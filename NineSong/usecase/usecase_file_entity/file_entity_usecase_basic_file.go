package usecase_file_entity

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_db_usecase"
	"github.com/dhowden/tag"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileUsecase struct {
	fileRepo    domain_file_entity.FileRepository
	folderRepo  domain_file_entity.FolderRepository
	detector    domain_file_entity.FileDetector
	targetTypes map[domain_file_entity.FileTypeNo]struct{}
	targetMutex sync.RWMutex
	workerPool  chan struct{}
	scanTimeout time.Duration

	audioExtractor scene_audio_db_usecase.AudioMetadataExtractorTaglib
	artistRepo     scene_audio_db_interface.ArtistRepository
	albumRepo      scene_audio_db_interface.AlbumRepository
	mediaRepo      scene_audio_db_interface.MediaFileRepository
	tempRepo       scene_audio_db_interface.TempRepository
}

func NewFileUsecase(
	fileRepo domain_file_entity.FileRepository,
	folderRepo domain_file_entity.FolderRepository,
	detector domain_file_entity.FileDetector,
	timeoutMinutes int,

	// 音频处理依赖项
	artistRepo scene_audio_db_interface.ArtistRepository,
	albumRepo scene_audio_db_interface.AlbumRepository,
	mediaRepo scene_audio_db_interface.MediaFileRepository,
	tempRepo scene_audio_db_interface.TempRepository,
) *FileUsecase {
	workerCount := runtime.NumCPU() * 2
	if workerCount < 4 {
		workerCount = 4
	}

	return &FileUsecase{
		fileRepo:    fileRepo,
		folderRepo:  folderRepo,
		detector:    detector,
		workerPool:  make(chan struct{}, workerCount),
		scanTimeout: time.Duration(timeoutMinutes) * time.Minute,
		artistRepo:  artistRepo,
		albumRepo:   albumRepo,
		mediaRepo:   mediaRepo,
		tempRepo:    tempRepo,
	}
}

func (uc *FileUsecase) ProcessDirectory(ctx context.Context, dirPath string, targetTypes []domain_file_entity.FileTypeNo) error {
	// 防御性检查
	if uc.folderRepo == nil {
		log.Printf("folderRepo未初始化")
		return fmt.Errorf("系统未正确初始化")
	}

	coverTempPath, _ := uc.tempRepo.GetTempPath(ctx, "cover")
	//steamPath, _ := uc.tempRepo.GetTempPath(ctx, "stream")

	folder, err := uc.folderRepo.FindByPath(ctx, dirPath)
	if err != nil {
		log.Printf("文件夹查询失败: %v", err)
		return fmt.Errorf("folder query failed: %w", err)
	}

	if folder == nil {
		newFolder := &domain_file_entity.FolderMetadata{
			ID:         primitive.NewObjectID(),
			FolderPath: dirPath,
			FolderMeta: domain_file_entity.FolderMeta{
				FileCount:   0,
				LastScanned: time.Now(),
			},
		}
		if err := uc.folderRepo.Insert(ctx, newFolder); err != nil {
			log.Printf("文件夹创建失败: %v", err)
			return fmt.Errorf("folder creation failed: %w", err)
		}
		folder = newFolder
	}

	// 设置目标文件类型
	uc.targetMutex.Lock()
	uc.targetTypes = make(map[domain_file_entity.FileTypeNo]struct{})
	for _, ft := range targetTypes {
		uc.targetTypes[ft] = struct{}{}
	}
	uc.targetMutex.Unlock()

	// 扫描前先清空下数据库表内的统计数据，扫描过程中会重新统计一次
	_, err = uc.albumRepo.ResetALLField(ctx)
	if err != nil {
		return err
	}
	_, err = uc.albumRepo.ResetField(ctx, "AllArtistIDs")
	if err != nil {
		return err
	}
	_, err = uc.artistRepo.ResetALLField(ctx)
	if err != nil {
		return err
	}

	// 并发处理管道
	var wgFile sync.WaitGroup
	errChan := make(chan error, 100)
	fileCount := 0

	// 遍历文件系统
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("文件遍历错误: %v", err)
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if info.IsDir() || !uc.shouldProcess(path) {
				return nil
			}

			wgFile.Add(1)
			fileCount++
			go uc.processFile(ctx, path, coverTempPath, folder.ID, &wgFile, errChan)
			return nil
		}
	})

	// 等待所有任务完成
	go func() {
		wgFile.Wait()
		close(errChan)
	}()

	// 收集错误
	var finalErr error
	for err := range errChan {
		log.Printf("文件处理错误: %v", err)
		if finalErr == nil {
			finalErr = err
		} else {
			finalErr = fmt.Errorf("%v; %w", finalErr, err)
		}
	}

	// 更新艺术家专辑统计数量
	ids, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取艺术家ID列表失败: %w", err)
	}
	maxConcurrency := 50
	sem := make(chan struct{}, maxConcurrency)
	var wgUpdate sync.WaitGroup
	counters := []struct {
		countMethod func(context.Context, string) (int64, error)
		counterName string
		countType   string
	}{
		{
			countMethod: uc.albumRepo.AlbumCountByArtist,
			counterName: "album_count",
			countType:   "专辑",
		},
		{
			countMethod: uc.albumRepo.GuestAlbumCountByArtist,
			counterName: "guest_album_count",
			countType:   "合作专辑",
		},
		{
			countMethod: uc.mediaRepo.MediaCountByArtist,
			counterName: "song_count",
			countType:   "单曲",
		},
		{
			countMethod: uc.mediaRepo.GuestMediaCountByArtist,
			counterName: "guest_song_count",
			countType:   "合作单曲",
		},
	}
	for _, artistID := range ids {
		wgUpdate.Add(1)
		go func(id primitive.ObjectID) {
			defer wgUpdate.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			strID := id.Hex()
			ctx := context.Background() // 根据实际情况决定是否传递原始context

			// 统一处理计数器更新
			for _, counter := range counters {
				count, err := counter.countMethod(ctx, strID)
				if err != nil {
					log.Printf("艺术家%s%s统计失败: %v", strID, counter.countType, err)
					continue
				}

				if _, err = uc.artistRepo.UpdateCounter(ctx, id, counter.counterName, int(count)); err != nil {
					log.Printf("艺术家%s%s计数更新失败: %v", strID, counter.countType, err)
				}
			}
		}(artistID) // 传递当前artistID的副本给goroutine
	}
	wgUpdate.Wait()

	// 更新文件夹统计
	if updateErr := uc.folderRepo.UpdateStats(ctx, folder.ID, fileCount); updateErr != nil {
		log.Printf("统计更新失败: %v", updateErr)
		return fmt.Errorf("stats update failed: %w", updateErr)
	}

	log.Printf("音乐库扫描完毕。。。。。。。。。。。。。")

	return finalErr
}

func (uc *FileUsecase) shouldProcess(path string) bool {
	fileType, err := uc.detector.DetectMediaType(path)
	if err != nil {
		log.Printf("文件类型检测失败: %v", err)
		return false
	}

	uc.targetMutex.RLock()
	defer uc.targetMutex.RUnlock()
	_, exists := uc.targetTypes[fileType]
	return exists
}

func (uc *FileUsecase) processFile(
	ctx context.Context,
	path string,
	coverTempPath string,
	folderID primitive.ObjectID,
	wg *sync.WaitGroup,
	errChan chan<- error,
) {
	defer wg.Done()

	// 上下文取消检查
	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return
	default:
	}

	// 获取工作槽
	select {
	case uc.workerPool <- struct{}{}:
		defer func() { <-uc.workerPool }()
	case <-ctx.Done():
		errChan <- ctx.Err()
		return
	}

	// 文件类型检测
	fileType, err := uc.detector.DetectMediaType(path)
	if err != nil {
		log.Printf("文件检测失败: %s | %v", path, err)
		errChan <- fmt.Errorf("文件检测失败 %s: %w", path, err)
		return
	}

	// 创建基础元数据
	metadata, err := uc.createMetadataBasicInfo(path, folderID)
	if err != nil {
		log.Printf("元数据创建失败: %s | %v", path, err)
		errChan <- fmt.Errorf("文件处理失败 %s: %w", path, err)
		return
	}

	// 保存基础文件信息
	if err := uc.fileRepo.Upsert(ctx, metadata); err != nil {
		log.Printf("文件写入失败: %s | %v", path, err)
		errChan <- fmt.Errorf("数据库写入失败 %s: %w", path, err)
		return
	}

	// 处理音频文件
	if fileType == domain_file_entity.Audio {
		mediaFile, album, artists, err := uc.audioExtractor.Extract(path, metadata)
		if err != nil {
			return
		}

		if err := uc.processAudioHierarchy(ctx, artists, album, mediaFile); err != nil {
			return
		}

		if err := uc.processAudioMediaFilesAndAlbumCover(
			ctx,
			mediaFile,
			album,
			path,
			coverTempPath,
		); err != nil {
			errChan <- fmt.Errorf("文件存储失败 %s: %w", path, err)
			return
		}
	}
}

func (uc *FileUsecase) createMetadataBasicInfo(
	path string,
	folderID primitive.ObjectID,
) (*domain_file_entity.FileMetadata, error) {
	// 1. 先查询是否已存在该路径文件
	existingFile, err := uc.fileRepo.FindByPath(context.Background(), path)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Printf("路径查询失败: %s | %v", path, err)
		return nil, fmt.Errorf("路径查询失败: %w", err)
	}

	// 2. 已存在则直接返回
	if existingFile != nil {
		return existingFile, nil
	}

	// 3. 不存在时执行原流程
	file, err := os.Open(path)
	if err != nil {
		log.Printf("文件打开失败: %s | %v", path, err)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("文件关闭失败: %s | %v", path, err)
		}
	}(file)

	stat, err := file.Stat()
	if err != nil {
		log.Printf("文件信息获取失败: %s | %v", path, err)
		return nil, err
	}

	fileType, err := uc.detector.DetectMediaType(path)
	if err != nil {
		log.Printf("文件类型检测失败: %s | %v", path, err)
		return nil, err
	}

	hash := sha256.New()
	if _, err := io.CopyBuffer(hash, file, make([]byte, 32*1024)); err != nil {
		log.Printf("哈希计算失败: %s | %v", path, err)
		return nil, err
	}

	normalizedPath := filepath.ToSlash(filepath.Clean(path))

	return &domain_file_entity.FileMetadata{
		ID:        primitive.NewObjectID(),
		FolderID:  folderID,
		FilePath:  normalizedPath,
		FileType:  fileType,
		Size:      stat.Size(),
		ModTime:   stat.ModTime(),
		Checksum:  fmt.Sprintf("%x", hash.Sum(nil)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (uc *FileUsecase) processAudioMediaFilesAndAlbumCover(
	ctx context.Context,
	media *scene_audio_db_models.MediaFileMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	path string,
	coverBasePath string,
) error {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("文件关闭失败[%s]: %v", path, err)
		}
	}(file)
	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return nil
	}

	mediaCoverDir := filepath.Join(coverBasePath, "media", media.ID.Hex())
	if err := os.MkdirAll(mediaCoverDir, 0755); err != nil {
		return fmt.Errorf("媒体目录创建失败: %w", err)
	}

	var coverPath string
	defer func() {
		if coverPath == "" {
			err := os.RemoveAll(mediaCoverDir)
			if err != nil {
				log.Printf("[WARN] 媒体目录删除失败 | 路径:%s | 错误:%v", mediaCoverDir, err)
				return
			}
		}
	}()

	if pic := metadata.Picture(); pic != nil && len(pic.Data) > 0 {
		targetPath := filepath.Join(mediaCoverDir, "cover.jpg")
		if err := os.WriteFile(targetPath, pic.Data, 0644); err != nil {
			return fmt.Errorf("封面写入失败: %w", err)
		}
		coverPath = targetPath
	}

	mediaUpdate := bson.M{
		"$set": bson.M{
			"medium_image_url": coverPath,
			"has_cover_art":    coverPath != "",
		},
	}
	if _, err := uc.mediaRepo.UpdateByID(ctx, media.ID, mediaUpdate); err != nil {
		return fmt.Errorf("媒体更新失败: %w", err)
	}

	if album != nil && coverPath != "" {
		if album.MediumImageURL == "" {
			albumCoverDir := filepath.Join(coverBasePath, "album", album.ID.Hex())
			if err := os.MkdirAll(albumCoverDir, 0755); err != nil {
				log.Printf("[WARN] 专辑封面目录创建失败 | AlbumID:%s | 错误:%v",
					album.ID.Hex(), err)
			} else if pic := metadata.Picture(); pic != nil && len(pic.Data) > 0 {
				albumCoverPath := filepath.Join(albumCoverDir, "cover.jpg")
				if err := os.WriteFile(albumCoverPath, pic.Data, 0644); err != nil {
					log.Printf("[WARN] 专辑封面写入失败 | AlbumID:%s | 路径:%s | 错误:%v",
						album.ID.Hex(), albumCoverPath, err)
				} else {
					coverPath = albumCoverPath // 优先使用专辑级封面路径
				}
			}

			albumUpdate := bson.M{
				"$set": bson.M{
					"medium_image_url": coverPath,
					"has_cover_art":    true,
					"updated_at":       time.Now().UTC(),
				},
			}
			if _, err := uc.albumRepo.UpdateByID(ctx, album.ID, albumUpdate); err != nil {
				log.Printf("[WARN] 专辑元数据更新失败 | ID:%s | 错误:%v",
					album.ID.Hex(), err)
			}
		}
	}

	return nil
}

func (uc *FileUsecase) processAudioHierarchy(ctx context.Context,
	artists []*scene_audio_db_models.ArtistMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	mediaFile *scene_audio_db_models.MediaFileMetadata,
) error {
	// 关键依赖检查
	if uc.mediaRepo == nil || uc.artistRepo == nil || uc.albumRepo == nil {
		log.Print("音频仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	if mediaFile == nil {
		log.Print("媒体文件元数据为空")
		return fmt.Errorf("mediaFile cannot be nil")
	}

	// 直接保存无关联数据
	if artists == nil && album == nil {
		if mediaFile, err := uc.mediaRepo.Upsert(ctx, mediaFile); err != nil {
			log.Printf("歌曲保存失败: %s | %v", mediaFile.Path, err)
			return fmt.Errorf("歌曲元数据保存失败 | 路径:%s | %w", mediaFile.Path, err)
		}
		return nil
	}

	// 处理艺术家
	if artists != nil {
		for _, artist := range artists {
			if artist.Name == "" {
				log.Print("艺术家名称为空")
				artist.Name = "Unknown"
			}
			if err := uc.updateAudioArtistMetadata(ctx, artist); err != nil {
				log.Printf("艺术家处理失败: %s | %v", artist.Name, err)
				return fmt.Errorf("艺术家处理失败 | 原因:%w", err)
			}
		}
	}

	// 处理专辑
	if album != nil {
		if album.Name == "" {
			log.Print("专辑名称为空")
			album.Name = "Unknown"
		}
		if err := uc.updateAudioAlbumMetadata(ctx, album); err != nil {
			log.Printf("专辑处理失败: %s | %v", album.Name, err)
			return fmt.Errorf("专辑处理失败 | 名称:%s | 原因:%w", album.Name, err)
		}
	}

	// 保存媒体文件
	if mediaFile, err := uc.mediaRepo.Upsert(ctx, mediaFile); err != nil {
		errorInfo := fmt.Sprintf("路径:%s", mediaFile.Path)
		if album != nil {
			errorInfo += fmt.Sprintf(" 专辑:%s", album.Name)
		}
		log.Printf("最终保存失败: %s | %v", errorInfo, err)
		return fmt.Errorf("歌曲写入失败 %s | %w", errorInfo, err)
	}

	// 异步统计更新
	go uc.updateAudioArtistAndAlbumStatistics(artists, album, mediaFile)
	return nil
}

func (uc *FileUsecase) updateAudioArtistAndAlbumStatistics(
	artists []*scene_audio_db_models.ArtistMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	mediaFile *scene_audio_db_models.MediaFileMetadata,
) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("统计更新发生panic: %v", r)
		}
	}()

	ctx := context.Background()
	var artistID, albumID primitive.ObjectID

	if artists != nil {
		for index, artist := range artists {
			if artist != nil && !artist.ID.IsZero() {
				artistID = artist.ID
			}
			if !artistID.IsZero() && index == 0 {
				if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "size", mediaFile.Size); err != nil {
					log.Printf("专辑大小统计更新失败: %v", err)
				}
				if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "duration", int(mediaFile.Duration)); err != nil {
					log.Printf("专辑播放时间统计更新失败: %v", err)
				}
			}
		}
	}

	if album != nil && !album.ID.IsZero() {
		albumID = album.ID
	}
	if !albumID.IsZero() {
		if _, err := uc.albumRepo.UpdateCounter(ctx, albumID, "song_count", 1); err != nil {
			log.Printf("专辑单曲数量统计更新失败: %v", err)
		}
		if _, err := uc.albumRepo.UpdateCounter(ctx, albumID, "size", mediaFile.Size); err != nil {
			log.Printf("专辑大小统计更新失败: %v", err)
		}
		if _, err := uc.albumRepo.UpdateCounter(ctx, albumID, "duration", int(mediaFile.Duration)); err != nil {
			log.Printf("专辑播放时间统计更新失败: %v", err)
		}
	}
}

func (uc *FileUsecase) updateAudioArtistMetadata(ctx context.Context, artist *scene_audio_db_models.ArtistMetadata) error {
	if uc.artistRepo == nil {
		log.Print("艺术家仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	existing, err := uc.artistRepo.GetByName(ctx, artist.Name)
	if err != nil {
		log.Printf("名称查询错误: %v", err)
	}
	if existing != nil {
		artist = existing
	} else {
		if err := uc.artistRepo.Upsert(ctx, artist); err != nil {
			log.Printf("艺术家创建失败: %s | %v", artist.Name, err)
			return err
		}
	}

	return nil
}

func (uc *FileUsecase) updateAudioAlbumMetadata(ctx context.Context, album *scene_audio_db_models.AlbumMetadata) error {
	if uc.albumRepo == nil {
		log.Print("专辑仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	filter := bson.M{
		"_id":       album.ID,
		"artist_id": album.ArtistID,
		"name":      album.Name,
	}

	existing, err := uc.albumRepo.GetByFilter(ctx, filter)
	if err != nil {
		log.Printf("组合查询错误: %v", err)
	}
	if existing != nil {
		album = existing
	} else {
		if err := uc.albumRepo.Upsert(ctx, album); err != nil {
			log.Printf("专辑创建失败: %s | %v", album.Name, err)
			return err
		}
	}

	return nil
}
