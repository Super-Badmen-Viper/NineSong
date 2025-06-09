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
	"strings"
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
	mediaCueRepo   scene_audio_db_interface.MediaFileCueRepository
}

func NewFileUsecase(
	fileRepo domain_file_entity.FileRepository,
	folderRepo domain_file_entity.FolderRepository,
	detector domain_file_entity.FileDetector,
	timeoutMinutes int,

	artistRepo scene_audio_db_interface.ArtistRepository,
	albumRepo scene_audio_db_interface.AlbumRepository,
	mediaRepo scene_audio_db_interface.MediaFileRepository,
	tempRepo scene_audio_db_interface.TempRepository,
	mediaCueRepo scene_audio_db_interface.MediaFileCueRepository,
) *FileUsecase {
	workerCount := runtime.NumCPU() * 2
	if workerCount < 4 {
		workerCount = 4
	}

	return &FileUsecase{
		fileRepo:     fileRepo,
		folderRepo:   folderRepo,
		detector:     detector,
		workerPool:   make(chan struct{}, workerCount),
		scanTimeout:  time.Duration(timeoutMinutes) * time.Minute,
		artistRepo:   artistRepo,
		albumRepo:    albumRepo,
		mediaRepo:    mediaRepo,
		tempRepo:     tempRepo,
		mediaCueRepo: mediaCueRepo,
	}
}

func (uc *FileUsecase) ProcessDirectory(
	ctx context.Context,
	dirPaths []string,
	targetTypes []domain_file_entity.FileTypeNo,
	ScanModel int,
) error {
	if uc.folderRepo == nil {
		log.Printf("folderRepo未初始化")
		return fmt.Errorf("系统未正确初始化")
	}

	var libraryFolders []*domain_file_entity.LibraryFolderMetadata

	// 1. 处理多个目录路径
	for _, dirPath := range dirPaths {
		folder, err := uc.folderRepo.FindLibrary(ctx, dirPath, targetTypes)
		if err != nil {
			log.Printf("文件夹查询失败: %v", err)
			return fmt.Errorf("folder query failed: %w", err)
		}

		// 2. 根据扫描模式处理目录
		if ScanModel == 0 || ScanModel == 2 {
			if folder == nil {
				newFolder := &domain_file_entity.LibraryFolderMetadata{
					ID:          primitive.NewObjectID(),
					FolderPath:  dirPath,
					FileTypes:   targetTypes,
					FileCount:   0,
					LastScanned: time.Now(),
				}
				if err := uc.folderRepo.Insert(ctx, newFolder); err != nil {
					log.Printf("文件夹创建失败: %v", err)
					return fmt.Errorf("folder creation failed: %w", err)
				}
				libraryFolders = append(libraryFolders, newFolder)
			} else {
				libraryFolders = append(libraryFolders, folder)
			}
		}
	}

	// 3. 获取整个媒体库（修复模式使用）
	if ScanModel == 1 {
		library, err := uc.folderRepo.GetAllByType(ctx, 1)
		if err != nil {
			log.Printf("文件夹查询失败: %v", err)
			return fmt.Errorf("folder query failed: %w", err)
		}
		libraryFolders = library
	}

	// 4. 设置目标文件类型
	uc.targetMutex.Lock()
	uc.targetTypes = make(map[domain_file_entity.FileTypeNo]struct{})
	for _, ft := range targetTypes {
		uc.targetTypes[ft] = struct{}{}
	}
	uc.targetMutex.Unlock()

	// 5. 根据扫描模式执行处理
	switch ScanModel {
	case 0: // 扫描新的和有修改的文件
		if isFileTypeSliceEqual(targetTypes, domain_file_entity.LibraryMusicType) {
			err := uc.ProcessMusicDirectory(ctx, libraryFolders, true, true, false)
			if err != nil {
				return err
			}
		}
	case 1: // 搜索缺失的元数据
		if isFileTypeSliceEqual(targetTypes, domain_file_entity.LibraryMusicType) {
			err := uc.ProcessMusicDirectory(ctx, libraryFolders, false, true, true)
			if err != nil {
				return err
			}
		}
	case 2: // 覆盖所有元数据
		if isFileTypeSliceEqual(targetTypes, domain_file_entity.LibraryMusicType) {
			err := uc.ProcessMusicDirectory(ctx, libraryFolders, true, true, true)
			if err != nil {
				return err
			}
		}
	}

	log.Printf("媒体库扫描完成，共处理%d个目录", len(dirPaths))
	return nil
}
func isFileTypeSliceEqual(a, b []domain_file_entity.FileTypeNo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (uc *FileUsecase) ProcessMusicDirectory(
	ctx context.Context,
	libraryFolders []*domain_file_entity.LibraryFolderMetadata,
	libraryTraversal bool, // 是否遍历传入的媒体库目录
	libraryStatistics bool, // 是否统计传入的媒体库目录
	libraryRefactoring bool, // 是否重构传入的媒体库目录（覆盖所有元数据）
) error {
	var finalErr error

	// 更新文件夹统计与状态
	for _, folderInfo := range libraryFolders {
		if updateErr := uc.folderRepo.UpdateStats(
			ctx, folderInfo.ID, folderInfo.FileCount, domain_file_entity.StatusActive,
		); updateErr != nil {
			log.Printf("媒体库：%s (ID %s) - 统计更新失败: %v", folderInfo.FolderPath, folderInfo.ID, updateErr)
			return fmt.Errorf("stats update failed: %w", updateErr)
		}
	}

	coverTempPath, _ := uc.tempRepo.GetTempPath(ctx, "cover")

	var libraryFolderNewInfos []struct {
		libraryFolderID        primitive.ObjectID
		libraryFolderPath      string
		libraryFolderFileCount int
	}
	var regularAudioPaths []string
	var cueAudioPaths []string

	// 扫描前重置数据库统计字段
	_, err := uc.albumRepo.ResetALLField(ctx)
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

	// 路径收集容器
	regularAudioPaths = make([]string, 0) // 常规音频路径
	cueAudioPaths = make([]string, 0)     // CUE音频路径
	// 路径去重容器
	addedCuePaths := make(map[string]bool)
	addedRegularPaths := make(map[string]bool)

	for _, folder := range libraryFolders {
		libraryFolderPath := strings.Replace(folder.FolderPath, "/", "\\", -1)
		if !strings.HasSuffix(libraryFolderPath, "\\") {
			libraryFolderPath += "\\"
		}

		folderInfo := struct {
			libraryFolderID        primitive.ObjectID
			libraryFolderPath      string
			libraryFolderFileCount int
		}{
			libraryFolderID:        folder.ID,
			libraryFolderPath:      libraryFolderPath,
			libraryFolderFileCount: 0,
		}

		// 并发处理管道
		var wgFile sync.WaitGroup
		errChan := make(chan error, 100)

		// 存储需要排除的.wav文件路径
		excludeWavs := make(map[string]struct{})
		// 存储cue文件及其资源
		cueResourcesMap := make(map[string]*scene_audio_db_models.CueConfig)

		// 第一次遍历：收集.cue文件和关联资源
		err = filepath.Walk(folder.FolderPath, func(path string, info os.FileInfo, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if info.IsDir() {
					return nil
				}

				ext := strings.ToLower(filepath.Ext(path))
				dir := filepath.Dir(path)
				baseName := strings.TrimSuffix(filepath.Base(path), ext)

				// 处理.cue文件
				if ext == ".cue" {
					res := &scene_audio_db_models.CueConfig{CuePath: path}

					// 查找同名音频文件
					audioFound := false
					audioExts := []string{".wav", ".ape", ".flac"} // 扩展支持的音频格式
					for _, audioExt := range audioExts {
						audioPath := filepath.Join(dir, baseName+audioExt)
						if _, err := os.Stat(audioPath); err == nil {
							res.AudioPath = audioPath
							excludeWavs[audioPath] = struct{}{} // 添加到排除列表
							audioFound = true
							break // 找到一种音频格式即停止
						}
					}

					// 未找到音频文件则跳过
					if !audioFound {
						return nil
					}

					// 收集相关资源文件[9](@ref)
					resourceFiles := []string{"back.jpg", "cover.jpg", "disc.jpg", "list.txt", "log.txt"}
					for _, f := range resourceFiles {
						p := filepath.Join(dir, f)
						if _, err := os.Stat(p); err == nil {
							switch f {
							case "back.jpg":
								res.BackImage = p
							case "cover.jpg":
								res.CoverImage = p
							case "disc.jpg":
								res.DiscImage = p
							case "list.txt":
								res.ListFile = p
							case "log.txt":
								res.LogFile = p
							}
						}
					}

					cueResourcesMap[path] = res
				}
				return nil
			}
		})
		if err != nil {
			log.Printf("文件夹%v，文件遍历错误: %v，请检查该文件夹是否存在", folder.FolderPath, err)
		}

		// 第二次遍历：处理文件
		err = filepath.Walk(folder.FolderPath, func(path string, info os.FileInfo, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if info.IsDir() || !uc.shouldProcess(path) {
					return nil
				}

				// 检查是否是被排除的.wav文件
				if _, excluded := excludeWavs[path]; excluded {
					return nil
				}

				ext := strings.ToLower(filepath.Ext(path))

				// 处理.cue文件
				if ext == ".cue" {
					res, exists := cueResourcesMap[path]
					if !exists {
						log.Printf("未找到cue文件对应的音频资源: %s", path)
						return nil
					}

					// 关键修复：路径收集不受libraryTraversal影响
					if res.AudioPath != "" {
						if _, existsAdded := addedCuePaths[res.AudioPath]; !existsAdded {
							cueAudioPaths = append(cueAudioPaths, res.AudioPath)
							addedCuePaths[res.AudioPath] = true
						}
					}

					// 仅当开启遍历时才执行文件处理和计数
					if libraryTraversal {
						wgFile.Add(1)
						folderInfo.libraryFolderFileCount++
						go uc.processFile(ctx, res, res.AudioPath, libraryFolderPath, coverTempPath, folder.ID, &wgFile, errChan)
					}
				} else {
					// 关键修复：路径收集不受libraryTraversal影响
					if _, existsAdded := addedRegularPaths[path]; !existsAdded {
						regularAudioPaths = append(regularAudioPaths, path)
						addedRegularPaths[path] = true
					}

					// 仅当开启遍历时才执行文件处理和计数
					if libraryTraversal {
						wgFile.Add(1)
						folderInfo.libraryFolderFileCount++
						go uc.processFile(ctx, nil, path, libraryFolderPath, coverTempPath, folder.ID, &wgFile, errChan)
					}
				}
				return nil
			}
		})
		if err != nil {
			log.Printf("文件夹%v，文件遍历错误: %v，请检查该文件夹是否存在", folder.FolderPath, err)
		}

		// 等待所有任务完成
		go func() {
			wgFile.Wait()
			close(errChan)
		}()

		// 收集错误
		for err := range errChan {
			log.Printf("文件处理错误: %v", err)
			if finalErr == nil {
				finalErr = err
			} else {
				finalErr = fmt.Errorf("%v; %w", finalErr, err)
			}
		}

		libraryFolderNewInfos = append(libraryFolderNewInfos, folderInfo)
	}

	artistIDs, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取艺术家ID列表失败: %w", err)
	}

	// 区域2: 媒体库统计（仅当libraryStatistics为true时执行）
	if libraryStatistics {
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
			{
				countMethod: uc.mediaCueRepo.MediaCueCountByArtist,
				counterName: "cue_count",
				countType:   "光盘",
			},
			{
				countMethod: uc.mediaCueRepo.GuestMediaCueCountByArtist,
				counterName: "guest_cue_count",
				countType:   "合作光盘",
			},
		}
		for _, artistID := range artistIDs {
			wgUpdate.Add(1)
			go func(id primitive.ObjectID) {
				defer wgUpdate.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				strID := id.Hex()
				ctx := context.Background()

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
			}(artistID)
		}
		wgUpdate.Wait()
	}

	// 区域3: 媒体库重构（仅当libraryRefactoring为true时执行）
	if libraryRefactoring {
		// 定义统一的结构体类型
		type artistStats struct {
			mediaCount         int
			guestMediaCount    int
			cueMediaCount      int
			guestCueMediaCount int
			albumCount         int
			guestAlbumCount    int
		}

		// 使用指针类型map
		invalidArtistCounts := make(map[string]*artistStats)
		deleteArtistCounts := make(map[string]int)
		invalidAlbumCounts := make(map[string]int)

		artistAlbums, err := uc.albumRepo.GetArtistAlbumsMap(ctx)
		if err != nil {
			return fmt.Errorf("获取专辑ID列表失败: %w", err)
		}

		artistGuestAlbums, err := uc.albumRepo.GetArtistGuestAlbumsMap(ctx)
		if err != nil {
			return fmt.Errorf("获取专辑ID列表失败: %w", err)
		}

		// 1. 处理常规音频的艺术家统计
		for _, id := range artistIDs {
			artistID := id.Hex()

			// 确保结构体已初始化[1,2](@ref)
			if _, exists := invalidArtistCounts[artistID]; !exists {
				invalidArtistCounts[artistID] = &artistStats{}
			}

			// 处理主艺术家统计
			mediaCount, err := uc.mediaRepo.InspectMediaCountByArtist(ctx, artistID, regularAudioPaths)
			if err != nil {
				log.Printf("常规音频主艺术家统计失败 (ID %s): %v", artistID, err)
			} else {
				if mediaCount >= 0 {
					invalidArtistCounts[artistID].mediaCount += mediaCount
				} else {
					invalidArtistCounts[artistID].mediaCount++
					deleteArtistCounts[artistID]++
				}
			}

			// 处理合作艺术家统计
			guestMediaCount, err := uc.mediaRepo.InspectGuestMediaCountByArtist(ctx, artistID, regularAudioPaths)
			if err != nil {
				log.Printf("常规音频合作艺术家统计失败 (ID %s): %v", artistID, err)
			} else {
				if guestMediaCount >= 0 {
					invalidArtistCounts[artistID].guestMediaCount += guestMediaCount
				} else {
					invalidArtistCounts[artistID].guestMediaCount++
					deleteArtistCounts[artistID]++
				}
			}
		}

		// 2. 处理常规音频的专辑统计
		for artistID, albums := range artistAlbums {
			artistIDStr := artistID.Hex()

			// 确保结构体已初始化[1,2](@ref)
			if _, exists := invalidArtistCounts[artistIDStr]; !exists {
				invalidArtistCounts[artistIDStr] = &artistStats{}
			}

			for _, albumID := range albums {
				albumIDStr := albumID.Hex()
				count, err := uc.mediaRepo.InspectMediaCountByAlbum(ctx, albumIDStr, regularAudioPaths)
				if err != nil {
					log.Printf("常规音频专辑统计失败 (ID %s): %v", albumIDStr, err)
					continue
				}

				if count > 0 {
					// 直接修改指针指向的结构体字段[7](@ref)
					invalidArtistCounts[artistIDStr].albumCount += count
					//
					invalidAlbumCounts[albumIDStr] += count
				} else if count < 0 {
					// 删除艺术家统计
					invalidArtistCounts[artistIDStr].albumCount++
					// 标记为删除
					invalidAlbumCounts[albumIDStr] = -1
					//
					deleteArtistCounts[artistIDStr]++
				} else {
					delete(invalidAlbumCounts, albumIDStr)
				}
			}
		}

		// 2.1 处理常规音频的合作专辑统计
		for artistID, albums := range artistGuestAlbums {
			artistIDStr := artistID.Hex()

			// 确保结构体已初始化[1,2](@ref)
			if _, exists := invalidArtistCounts[artistIDStr]; !exists {
				invalidArtistCounts[artistIDStr] = &artistStats{}
			}

			for _, albumID := range albums {
				albumIDStr := albumID.Hex()
				count, err := uc.mediaRepo.InspectGuestMediaCountByAlbum(ctx, albumIDStr, regularAudioPaths)
				if err != nil {
					log.Printf("常规音频专辑统计失败 (ID %s): %v", albumIDStr, err)
					continue
				}

				if count > 0 {
					invalidArtistCounts[artistIDStr].guestAlbumCount += count
				} else if count < 0 {
					invalidArtistCounts[artistIDStr].guestAlbumCount++
					deleteArtistCounts[artistIDStr]++
				}
			}
		}

		// 3. 清除常规音频的无效专辑（删除统计：单曲数为0的专辑）
		invalidAlbumNums := 0
		for albumID, operand := range invalidAlbumCounts {
			deleted, err := uc.albumRepo.InspectAlbumMediaCountByAlbum(ctx, albumID, operand)
			if err != nil {
				log.Printf("专辑清理失败 (ID %s): %v", albumID, err)
			}
			if deleted {
				invalidAlbumNums++
			}
		}
		log.Printf("已删除 %v 个无效专辑\n", invalidAlbumNums)

		// 4. 处理CUE音频的艺术家统计
		for _, id := range artistIDs {
			artistID := id.Hex()

			// 确保结构体已初始化[1,2](@ref)
			if _, exists := invalidArtistCounts[artistID]; !exists {
				invalidArtistCounts[artistID] = &artistStats{}
			}

			// 处理CUE音频主艺术家统计
			cueMediaCount, err := uc.mediaCueRepo.InspectMediaCueCountByArtist(ctx, artistID, cueAudioPaths)
			if err != nil {
				log.Printf("CUE音频主艺术家统计失败 (ID %s): %v", artistID, err)
			} else {
				if cueMediaCount >= 0 {
					invalidArtistCounts[artistID].cueMediaCount += cueMediaCount
				} else {
					invalidArtistCounts[artistID].cueMediaCount++
					deleteArtistCounts[artistID]++
				}
			}

			// 处理CUE音频合作艺术家统计
			guestCueMediaCount, err := uc.mediaCueRepo.InspectGuestMediaCueCountByArtist(ctx, artistID, cueAudioPaths)
			if err != nil {
				log.Printf("CUE音频合作艺术家统计失败 (ID %s): %v", artistID, err)
			} else {
				if guestCueMediaCount >= 0 {
					invalidArtistCounts[artistID].guestCueMediaCount += guestCueMediaCount
				} else {
					invalidArtistCounts[artistID].guestCueMediaCount++
					deleteArtistCounts[artistID]++
				}
			}
		}

		// 5. 清除无效的艺术家（删除统计：【单曲/合作单曲+CD/合作CD都为0+专辑/合作专辑都为0】的艺术家）
		for _, artistID := range artistIDs {
			artistIDStr := artistID.Hex()
			if deleteCount, exists := deleteArtistCounts[artistIDStr]; exists && deleteCount >= 6 {
				err := uc.artistRepo.DeleteByID(ctx, artistID)
				if err != nil {
					log.Printf("艺术家删除失败 (ID %s): %v", artistIDStr, err)
				} else {
					log.Printf("已删除无效艺术家 (ID %s)", artistIDStr)
				}
				delete(invalidArtistCounts, artistIDStr)
			}
		}

		// 6. 清除无效的常规音频
		delResult, invalidMediaArtist, err := uc.mediaRepo.DeleteAllInvalid(ctx, regularAudioPaths)
		if err != nil {
			log.Printf("常规音频清理失败: %v", err)
			return fmt.Errorf("regular audio cleanup failed: %w", err)
		} else if delResult > 0 {
			for _, artist := range invalidMediaArtist {
				artistID := artist.ArtistID.Hex()
				// 确保结构体已初始化
				if _, exists := invalidArtistCounts[artistID]; !exists {
					invalidArtistCounts[artistID] = &artistStats{}
				}
				invalidArtistCounts[artistID].mediaCount = int(artist.Count)
			}
		}

		// 7. 清除无效的CUE音频
		delCueResult, invalidMediaCueArtist, err := uc.mediaCueRepo.DeleteAllInvalid(ctx, cueAudioPaths)
		if err != nil {
			log.Printf("CUE音频清理失败: %v", err)
			return fmt.Errorf("CUE audio cleanup failed: %w", err)
		} else if delCueResult > 0 {
			for _, artist := range invalidMediaCueArtist {
				artistID := artist.ArtistID.Hex()
				// 确保结构体已初始化
				if _, exists := invalidArtistCounts[artistID]; !exists {
					invalidArtistCounts[artistID] = &artistStats{}
				}
				invalidArtistCounts[artistID].cueMediaCount = int(artist.Count)
			}
		}

		// 8. 重构媒体库艺术家操作数
		artistOperands := make(map[string]struct {
			albumCount         int
			guestAlbumCount    int
			mediaCount         int
			guestMediaCount    int
			cueMediaCount      int
			cueGuestMediaCount int
		})

		// 收集常规、CUE音频艺术家操作数
		for artistID, stats := range invalidArtistCounts {
			op := artistOperands[artistID]
			op.albumCount = stats.albumCount
			op.guestAlbumCount = stats.guestAlbumCount
			op.mediaCount = stats.mediaCount
			op.guestMediaCount = stats.guestMediaCount
			op.cueMediaCount = stats.cueMediaCount
			op.cueGuestMediaCount = stats.guestCueMediaCount
			artistOperands[artistID] = op
		}

		// 应用艺术家操作数。不能为0，否则将删除艺术家
		for artistID, operands := range artistOperands {
			// 专辑计数处理
			if operands.albumCount > 0 {
				operands.albumCount, err = uc.artistRepo.InspectAlbumCountByArtist(ctx, artistID, operands.albumCount)
				if err != nil {
					log.Printf("艺术家专辑计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 合作专辑计数处理
			if operands.guestAlbumCount > 0 {
				operands.guestAlbumCount, err = uc.artistRepo.InspectGuestAlbumCountByArtist(ctx, artistID, operands.guestAlbumCount)
				if err != nil {
					log.Printf("艺术家合作专辑计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 单曲计数处理
			if operands.mediaCount > 0 {
				operands.mediaCount, err = uc.artistRepo.InspectMediaCountByArtist(ctx, artistID, operands.mediaCount)
				if err != nil {
					log.Printf("艺术家单曲计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 合作单曲计数处理
			if operands.guestMediaCount > 0 {
				operands.guestMediaCount, err = uc.artistRepo.InspectGuestMediaCountByArtist(ctx, artistID, operands.guestMediaCount)
				if err != nil {
					log.Printf("艺术家合作单曲计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 光盘计数处理
			if operands.cueMediaCount > 0 {
				operands.cueMediaCount, err = uc.artistRepo.InspectMediaCueCountByArtist(ctx, artistID, operands.cueMediaCount)
				if err != nil {
					log.Printf("艺术家光盘计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 合作光盘计数处理
			if operands.cueGuestMediaCount > 0 {
				operands.cueGuestMediaCount, err = uc.artistRepo.InspectGuestMediaCueCountByArtist(ctx, artistID, operands.cueGuestMediaCount)
				if err != nil {
					log.Printf("艺术家合作光盘计数处理失败 (ID %s): %v", artistID, err)
				}
			}
		}

		// 清除无效的艺术家：应用艺术家操作数之后
		invalid, err := uc.artistRepo.DeleteAllInvalid(ctx)
		if err != nil {
			return err
		}
		if invalid > 0 {
			log.Printf("已删除 %d 个无效艺术家", invalid)
		}
	}

	// 更新文件夹统计（仅在执行遍历时更新）
	if libraryTraversal {
		for _, folderInfo := range libraryFolderNewInfos {
			if updateErr := uc.folderRepo.UpdateStats(
				ctx, folderInfo.libraryFolderID, folderInfo.libraryFolderFileCount, domain_file_entity.StatusActive,
			); updateErr != nil {
				log.Printf("媒体库：%s (ID %s) - 统计更新失败: %v", folderInfo.libraryFolderPath, folderInfo.libraryFolderID, updateErr)
				return fmt.Errorf("stats update failed: %w", updateErr)
			}
		}
	}

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
	res *scene_audio_db_models.CueConfig,
	path string,
	libraryPath string,
	coverTempPath string,
	libraryFolderID primitive.ObjectID,
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
	metadata, err := uc.createMetadataBasicInfo(path, libraryFolderID)
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
		mediaFile, album, artists, mediaFileCue, err := uc.audioExtractor.Extract(path, libraryPath, metadata, res)
		if err != nil {
			return
		}

		if err := uc.processAudioHierarchy(ctx, artists, album, mediaFile, mediaFileCue); err != nil {
			return
		}

		if err := uc.processAudioMediaFilesAndAlbumCover(
			ctx,
			mediaFile,
			album,
			mediaFileCue,
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
	libraryFolderID primitive.ObjectID,
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
		FolderID:  libraryFolderID,
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
	mediaCue *scene_audio_db_models.MediaFileCueMetadata,
	path string,
	coverBasePath string,
) error {
	if mediaCue != nil {
		mediaCueUpdate := bson.M{
			"$set": bson.M{
				"back_image_url":  mediaCue.CueResources.BackImage,
				"cover_image_url": mediaCue.CueResources.CoverImage,
				"disc_image_url":  mediaCue.CueResources.DiscImage,
				"has_cover_art":   mediaCue.CueResources.CoverImage != "",
			},
		}
		if _, err := uc.mediaCueRepo.UpdateByID(ctx, mediaCue.ID, mediaCueUpdate); err != nil {
			return fmt.Errorf("媒体更新失败: %w", err)
		}
	} else {
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
	}

	return nil
}

func (uc *FileUsecase) processAudioHierarchy(ctx context.Context,
	artists []*scene_audio_db_models.ArtistMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	mediaFile *scene_audio_db_models.MediaFileMetadata,
	mediaFileCue *scene_audio_db_models.MediaFileCueMetadata,
) error {
	// 关键依赖检查
	if uc.mediaRepo == nil || uc.artistRepo == nil || uc.albumRepo == nil || uc.mediaCueRepo == nil {
		log.Print("音频仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	if mediaFile == nil && mediaFileCue == nil {
		log.Print("媒体文件元数据为空")
		return fmt.Errorf("mediaFile cannot be nil")
	}

	// 直接保存无关联数据
	if artists == nil && album == nil {
		if mediaFile != nil {
			if mediaFile, err := uc.mediaRepo.Upsert(ctx, mediaFile); err != nil {
				log.Printf("歌曲保存失败: %s | %v", mediaFile.Path, err)
				return fmt.Errorf("歌曲元数据保存失败 | 路径:%s | %w", mediaFile.Path, err)
			}
		}
		if mediaFileCue != nil {
			if mediaFileCue, err := uc.mediaCueRepo.Upsert(ctx, mediaFileCue); err != nil {
				if mediaFileCue != nil {
					log.Printf("CUE文件保存失败: %s | %v", mediaFileCue.CueResources.CuePath, err)
				} else {
					log.Printf("CUE文件保存失败: %s | %v", mediaFileCue, err)
				}
			}
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
	if mediaFile != nil {
		if mediaFile, err := uc.mediaRepo.Upsert(ctx, mediaFile); err != nil {
			errorInfo := fmt.Sprintf("路径:%s", mediaFile.Path)
			if album != nil {
				errorInfo += fmt.Sprintf(" 专辑:%s", album.Name)
			}
			log.Printf("最终保存失败: %s | %v", errorInfo, err)
			return fmt.Errorf("歌曲写入失败 %s | %w", errorInfo, err)
		}
	}
	if mediaFileCue != nil {
		if mediaFileCue, err := uc.mediaCueRepo.Upsert(ctx, mediaFileCue); err != nil {
			if mediaFileCue != nil {
				errorInfo := fmt.Sprintf("路径:%s", mediaFileCue.CueResources.CuePath)
				if album != nil {
					errorInfo += fmt.Sprintf(" 专辑:%s", album.Name)
				}
			} else {
				log.Printf("最终保存失败: %s | %v", mediaFileCue, err)
			}
			return fmt.Errorf("歌曲写入失败 %s | %w", mediaFileCue, err)
		}
	}

	// 异步统计更新
	go uc.updateAudioArtistAndAlbumStatistics(artists, album, mediaFile, mediaFileCue)
	return nil
}

func (uc *FileUsecase) updateAudioArtistAndAlbumStatistics(
	artists []*scene_audio_db_models.ArtistMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	mediaFile *scene_audio_db_models.MediaFileMetadata,
	mediaFileCue *scene_audio_db_models.MediaFileCueMetadata,
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
				if mediaFileCue != nil {
					if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "size", mediaFileCue.Size); err != nil {
						log.Printf("专辑大小统计更新失败: %v", err)
					}
					if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "duration", int(mediaFileCue.CueDuration)); err != nil {
						log.Printf("专辑播放时间统计更新失败: %v", err)
					}
				}
				if mediaFile != nil {
					if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "size", mediaFile.Size); err != nil {
						log.Printf("专辑大小统计更新失败: %v", err)
					}
					if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "duration", int(mediaFile.Duration)); err != nil {
						log.Printf("专辑播放时间统计更新失败: %v", err)
					}
				}
			}
		}
	}

	if album != nil && !album.ID.IsZero() {
		albumID = album.ID
	}
	if !albumID.IsZero() {
		if mediaFile != nil {
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
	// 仅同步ID，因为album会随着更新而新增字段
	if existing != nil {
		artist.ID = existing.ID
	}
	// 再次插入，将版本更新的字段覆盖到现有文档
	if err := uc.artistRepo.Upsert(ctx, artist); err != nil {
		log.Printf("艺术家创建失败: %s | %v", artist.Name, err)
		return err
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
	// 仅同步ID，因为album会随着更新而新增字段
	if existing != nil {
		album.ID = existing.ID
	}
	// 再次插入，将版本更新的字段覆盖到现有文档
	if err := uc.albumRepo.Upsert(ctx, album); err != nil {
		log.Printf("专辑创建失败: %s | %v", album.Name, err)
		return err
	}

	return nil
}
