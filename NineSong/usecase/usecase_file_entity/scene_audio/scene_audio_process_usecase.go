package scene_audio

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_util"
	"github.com/mozillazg/go-pinyin"
	ffmpeggo "github.com/u2takey/ffmpeg-go"
	driver "go.mongodb.org/mongo-driver/mongo"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/dhowden/tag"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AudioProcessingUsecase struct {
	db               mongo.Database
	fileRepo         domain_file_entity.FileRepository
	folderRepo       domain_file_entity.FolderRepository
	detector         domain_file_entity.FileDetector
	workerPool       chan struct{}
	scanTimeout      time.Duration
	audioExtractor   usercase_audio_util.AudioMetadataExtractorTaglib
	artistRepo       scene_audio_db_interface.ArtistRepository
	albumRepo        scene_audio_db_interface.AlbumRepository
	mediaRepo        scene_audio_db_interface.MediaFileRepository
	tempRepo         scene_audio_db_interface.TempRepository
	mediaCueRepo     scene_audio_db_interface.MediaFileCueRepository
	wordCloudRepo    scene_audio_db_interface.WordCloudDBRepository
	lyricsFileRepo   scene_audio_db_interface.LyricsFileRepository
	scanMutex        sync.RWMutex
	scanProgress     float32
	scanStageWeights struct {
		traversal   float32
		statistics  float32
		refactoring float32
	}
}

func NewAudioProcessingUsecase(
	db mongo.Database,
	fileRepo domain_file_entity.FileRepository,
	folderRepo domain_file_entity.FolderRepository,
	detector domain_file_entity.FileDetector,
	timeoutMinutes int,
	artistRepo scene_audio_db_interface.ArtistRepository,
	albumRepo scene_audio_db_interface.AlbumRepository,
	mediaRepo scene_audio_db_interface.MediaFileRepository,
	tempRepo scene_audio_db_interface.TempRepository,
	mediaCueRepo scene_audio_db_interface.MediaFileCueRepository,
	wordCloudRepo scene_audio_db_interface.WordCloudDBRepository,
	lyricsFileRepo scene_audio_db_interface.LyricsFileRepository,
) *AudioProcessingUsecase {
	workerCount := runtime.NumCPU()

	return &AudioProcessingUsecase{
		db:             db,
		fileRepo:       fileRepo,
		folderRepo:     folderRepo,
		detector:       detector,
		workerPool:     make(chan struct{}, workerCount),
		scanTimeout:    time.Duration(timeoutMinutes) * time.Minute,
		artistRepo:     artistRepo,
		albumRepo:      albumRepo,
		mediaRepo:      mediaRepo,
		tempRepo:       tempRepo,
		mediaCueRepo:   mediaCueRepo,
		wordCloudRepo:  wordCloudRepo,
		lyricsFileRepo: lyricsFileRepo,
	}
}

func (uc *AudioProcessingUsecase) ProcessMusicDirectory(
	ctx context.Context,
	libraryFolders []*domain_file_entity.LibraryFolderMetadata,
	taskProg *domain_util.TaskProgress,
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

	// 修复：统计总文件数并设置到任务进度
	totalFiles := 0
	for _, folder := range libraryFolders {
		count, err := fastCountFilesInFolder(folder.FolderPath)
		if err != nil {
			log.Printf("统计文件数失败: %v", err)
			continue
		}
		totalFiles += count
	}
	// 更新任务状态
	taskProg.Status = "processing"
	taskProg.AddTotalFiles(totalFiles)

	// 获取封面临时目录
	coverTempPath, _ := uc.tempRepo.GetTempPath(ctx, "cover")

	// 获取歌词临时目录
	lyricsTempPath, _ := uc.tempRepo.GetTempPath(ctx, "lyrics")
	// 直接使用时间戳作为目录名（无前缀）
	timestamp := time.Now().Format("20060102150405") // 格式：YYYYMMDDHHMMSS
	lyricsTempPathWithTimestamp := filepath.Join(lyricsTempPath, timestamp)
	// 创建时间戳目录
	if err := os.MkdirAll(lyricsTempPathWithTimestamp, 0755); err != nil {
		return fmt.Errorf("创建时间戳歌词目录失败: %w", err)
	}
	// 自动清理旧文件夹（保留最新5个）
	go uc.cleanupOldLyricsFoldersAsync(lyricsTempPath)
	//
	if _, err := uc.lyricsFileRepo.CleanAll(ctx); err != nil {
		log.Printf("清空歌词文件数据库失败: %v", err)
	}

	var libraryFolderNewInfos []struct {
		libraryFolderID        primitive.ObjectID
		libraryFolderPath      string
		libraryFolderFileCount int
	}
	var regularAudioPaths []string
	var cueAudioPaths []string

	// 扫描前记录数据库统计字段，用以后续比对决定是否设置UpdatedAt
	artistCountsBeforeScanning, err := uc.artistRepo.GetAllCounts(ctx)
	//albumCountsBeforeScanning, err := uc.albumRepo.GetAllCounts(ctx)
	// 扫描前重置数据库统计字段
	if libraryStatistics {
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
		cueResourcesMap := make(map[string]*scene_audio_db_models.CueConfig)

		// 遍历时统计总文件数
		for _, folderCount := range libraryFolders {
			total := 0
			err := filepath.Walk(folderCount.FolderPath, func(path string, info os.FileInfo, err error) error {
				select {
				case <-ctx.Done(): // 响应取消
					return ctx.Err()
				default:
				}
				if !info.IsDir() && uc.shouldProcess(path, 1) {
					total++
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		// 第一次遍历：收集.cue文件和关联资源
		err := filepath.Walk(folder.FolderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("访问路径 %s 出错: %v", path, err)
				return nil // 跳过错误继续遍历
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if info.IsDir() {
					return nil
				}

				// 修复：更新任务级别的遍历计数器
				atomic.AddInt32(&taskProg.WalkedFiles, 1)

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
							excludeWavs[audioPath] = struct{}{}
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
			if err != nil {
				log.Printf("访问路径 %s 出错: %v", path, err)
				return nil // 跳过错误继续遍历
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if info.IsDir() || !uc.shouldProcess(path, 1) {
					return nil
				}

				// 修复：更新任务级别的遍历计数器
				atomic.AddInt32(&taskProg.WalkedFiles, 1)

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
						go uc.processFile(
							ctx, res,
							res.AudioPath, libraryFolderPath,
							coverTempPath, lyricsTempPathWithTimestamp,
							folder.ID, &wgFile, errChan, taskProg,
						)
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
						go uc.processFile(
							ctx, nil,
							path, libraryFolderPath,
							coverTempPath, lyricsTempPathWithTimestamp,
							folder.ID, &wgFile, errChan, taskProg,
						)
					}
					return nil
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

	albumIDs, err := uc.albumRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取专辑ID列表失败: %w", err)
	}

	// 区域2: 媒体库统计（仅当libraryStatistics为true时执行）
	if libraryStatistics {
		maxConcurrency := 50
		sem := make(chan struct{}, maxConcurrency)
		var wgUpdate sync.WaitGroup

		// 定义艺术家计数器
		artistCounters := []struct {
			countMethod func(context.Context, string) (int64, error)
			counterName string
			countType   string
		}{
			{
				countMethod: uc.albumRepo.AlbumCountByArtist,
				counterName: "album_count",
				countType:   "艺术家-专辑",
			},
			{
				countMethod: uc.albumRepo.GuestAlbumCountByArtist,
				counterName: "guest_album_count",
				countType:   "艺术家-合作专辑",
			},
			{
				countMethod: uc.mediaRepo.MediaCountByArtist,
				counterName: "song_count",
				countType:   "艺术家-单曲",
			},
			{
				countMethod: uc.mediaRepo.GuestMediaCountByArtist,
				counterName: "guest_song_count",
				countType:   "艺术家-合作单曲",
			},
			{
				countMethod: uc.mediaCueRepo.MediaCueCountByArtist,
				counterName: "cue_count",
				countType:   "艺术家-光盘",
			},
			{
				countMethod: uc.mediaCueRepo.GuestMediaCueCountByArtist,
				counterName: "guest_cue_count",
				countType:   "艺术家-合作光盘",
			},
		}

		// 更新当前阶段为统计阶段
		uc.scanMutex.Lock()
		uc.scanProgress = uc.scanStageWeights.traversal // 文件处理已完成
		uc.scanMutex.Unlock()

		totalArtists := len(artistIDs)
		currentArtist := 0

		// 统计艺术家计数器
		for _, artistID := range artistIDs {
			wgUpdate.Add(1)
			go func(id primitive.ObjectID) {
				defer wgUpdate.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				strID := id.Hex()
				ctx := context.Background()

				// 统一处理艺术家计数器
				for _, counter := range artistCounters {
					count, err := counter.countMethod(ctx, strID)
					if err != nil {
						log.Printf("艺术家%s%s统计失败: %v", strID, counter.countType, err)
						continue
					}
					if _, err = uc.artistRepo.UpdateCounter(ctx, id, counter.counterName, int(count)); err != nil {
						log.Printf("艺术家%s%s计数更新失败: %v", strID, counter.countType, err)
					}
				}

				// 更新艺术家统计进度
				currentArtist++
				uc.scanMutex.Lock()
				traversalProgress := uc.scanStageWeights.traversal
				statisticsProgress := uc.scanStageWeights.statistics * float32(currentArtist) / float32(totalArtists) * 0.5
				uc.scanProgress = traversalProgress + statisticsProgress
				uc.scanMutex.Unlock()
			}(artistID)
		}
		wgUpdate.Wait()

		// ---- 新增专辑统计部分 ----
		totalAlbums := len(albumIDs)
		currentAlbum := 0
		var wgAlbum sync.WaitGroup

		// 定义专辑计数器
		albumCounters := []struct {
			countMethod func(context.Context, string) (int64, error)
			counterName string
			countType   string
		}{
			{
				countMethod: uc.mediaRepo.MediaCountByAlbum,
				counterName: "song_count",
				countType:   "专辑-歌曲",
			},
		}

		// 统计专辑计数器
		for _, albumID := range albumIDs {
			wgAlbum.Add(1)
			go func(id primitive.ObjectID) {
				defer wgAlbum.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				strID := id.Hex()
				ctx := context.Background()

				// 处理专辑计数器
				for _, counter := range albumCounters {
					count, err := counter.countMethod(ctx, strID)
					if err != nil {
						log.Printf("专辑%s%s统计失败: %v", strID, counter.countType, err)
						continue
					}
					if count == 0 {
						if err = uc.albumRepo.DeleteByID(ctx, id); err != nil {
							log.Printf("失效专辑%s%s删除失败: %v", strID, counter.countType, err)
						}
					} else if count > 0 {
						if _, err = uc.albumRepo.UpdateCounter(ctx, id, counter.counterName, int(count)); err != nil {
							log.Printf("专辑%s%s计数更新失败: %v", strID, counter.countType, err)
						}
					}
				}

				// 更新专辑统计进度
				currentAlbum++
				uc.scanMutex.Lock()
				traversalProgress := uc.scanStageWeights.traversal
				// 进度计算：艺术家部分(50%) + 专辑完成比例(50%)
				albumRatio := float32(currentAlbum) / float32(totalAlbums)
				statisticsProgress := uc.scanStageWeights.statistics * (0.5 + 0.5*albumRatio)
				uc.scanProgress = traversalProgress + statisticsProgress
				uc.scanMutex.Unlock()
			}(albumID)
		}
		wgAlbum.Wait()
	}

	// 区域3: 媒体库重构（仅当libraryRefactoring为true时执行）
	if libraryRefactoring {
		// 更新当前阶段为重构阶段
		uc.scanMutex.Lock()
		uc.scanProgress = uc.scanStageWeights.traversal + uc.scanStageWeights.statistics
		uc.scanMutex.Unlock()

		totalArtists := len(artistIDs)
		current := 0

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

			// 更新重构阶段进度
			current++
			uc.scanMutex.Lock()
			progress := uc.scanStageWeights.traversal +
				uc.scanStageWeights.statistics +
				uc.scanStageWeights.refactoring*float32(current)/float32(totalArtists)
			uc.scanProgress = progress
			uc.scanMutex.Unlock()
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

			// 更新重构阶段进度
			current++
			uc.scanMutex.Lock()
			progress := uc.scanStageWeights.traversal +
				uc.scanStageWeights.statistics +
				uc.scanStageWeights.refactoring*float32(current)/float32(totalArtists)
			uc.scanProgress = progress
			uc.scanMutex.Unlock()
		}

		// 5. 清除无效的艺术家（删除统计：【单曲/合作单曲+CD/合作CD都为0+专辑/合作专辑都为0】的艺术家）
		artistIDStrCount := 0
		for _, artistID := range artistIDs {
			artistIDStr := artistID.Hex()
			if deleteCount, exists := deleteArtistCounts[artistIDStr]; exists && deleteCount >= 6 {
				err := uc.artistRepo.DeleteByID(ctx, artistID)
				if err != nil {
					log.Printf("艺术家删除失败 (ID %s): %v", artistIDStr, err)
				} else {
					artistIDStrCount++
				}
				delete(invalidArtistCounts, artistIDStr)
			}

			// 更新重构阶段进度
			current++
			uc.scanMutex.Lock()
			progress := uc.scanStageWeights.traversal +
				uc.scanStageWeights.statistics +
				uc.scanStageWeights.refactoring*float32(current)/float32(totalArtists)
			uc.scanProgress = progress
			uc.scanMutex.Unlock()
		}
		log.Printf("已删除 %d 个无效艺术家", artistIDStrCount)

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

	// 执行词云处理
	err = uc.processMusicWordCloud(ctx)
	if err != nil {
		return err
	}

	// 获取扫描后的计数数据
	artistCountsAfterScanning, err := uc.artistRepo.GetAllCounts(ctx)
	//albumCountsAfterScanning, err := uc.albumRepo.GetAllCounts(ctx)

	// 比较并更新发生变化的艺术家和专辑
	if err := uc.updateChangedArtistsAndAlbums(
		ctx,
		artistCountsBeforeScanning,
		artistCountsAfterScanning,
		nil,
		nil,
	); err != nil {
		log.Printf("更新变更记录失败: %v", err)
	}

	return finalErr
}

// 比较并更新发生变化的艺术家和专辑
func (uc *AudioProcessingUsecase) updateChangedArtistsAndAlbums(
	ctx context.Context,
	beforeArtist []scene_audio_db_models.ArtistAlbumAndSongCounts,
	afterArtist []scene_audio_db_models.ArtistAlbumAndSongCounts,
	beforeAlbum []scene_audio_db_models.AlbumSongCounts,
	afterAlbum []scene_audio_db_models.AlbumSongCounts,
) error {
	// 4. 批量更新变更记录的UpdatedAt
	currentTime := time.Now().UTC()

	if beforeArtist != nil && afterArtist != nil {
		// 1. 创建ID到计数的映射以便快速查找
		beforeArtistMap := make(map[primitive.ObjectID]scene_audio_db_models.ArtistAlbumAndSongCounts)
		for _, a := range beforeArtist {
			beforeArtistMap[a.ID] = a
		}

		afterArtistMap := make(map[primitive.ObjectID]scene_audio_db_models.ArtistAlbumAndSongCounts)
		for _, a := range afterArtist {
			afterArtistMap[a.ID] = a
		}

		// 2. 识别变更的艺术家
		var changedArtists []primitive.ObjectID
		for id, after := range afterArtistMap {
			before, exists := beforeArtistMap[id]
			if !exists {
				// 新增艺术家
				changedArtists = append(changedArtists, id)
				continue
			}

			// 比较所有计数字段
			if before.AlbumCount != after.AlbumCount ||
				before.GuestAlbumCount != after.GuestAlbumCount ||
				before.SongCount != after.SongCount ||
				before.GuestSongCount != after.GuestSongCount ||
				before.CueCount != after.CueCount ||
				before.GuestCueCount != after.GuestCueCount {
				changedArtists = append(changedArtists, id)
			}
		}

		// 更新艺术家
		if len(changedArtists) > 0 {
			filter := bson.M{"_id": bson.M{"$in": changedArtists}}
			update := bson.M{"$set": bson.M{"updated_at": currentTime}}
			if _, err := uc.artistRepo.UpdateMany(ctx, filter, update); err != nil {
				return fmt.Errorf("艺术家更新时间失败: %w", err)
			}
			log.Printf("已更新 %d 位艺术家的更新时间", len(changedArtists))
		}
	}

	if beforeAlbum != nil && afterAlbum != nil {
		beforeAlbumMap := make(map[primitive.ObjectID]int)
		for _, a := range beforeAlbum {
			beforeAlbumMap[a.ID] = a.SongCount
		}

		// 3. 识别变更的专辑
		var changedAlbums []primitive.ObjectID
		for _, after := range afterAlbum {
			beforeCount, exists := beforeAlbumMap[after.ID]
			if !exists {
				// 新增专辑
				changedAlbums = append(changedAlbums, after.ID)
				continue
			}

			if beforeCount != after.SongCount {
				changedAlbums = append(changedAlbums, after.ID)
			}
		}

		// 更新专辑
		if len(changedAlbums) > 0 {
			filter := bson.M{"_id": bson.M{"$in": changedAlbums}}
			update := bson.M{"$set": bson.M{"updated_at": currentTime}}
			if _, err := uc.albumRepo.UpdateMany(ctx, filter, update); err != nil {
				return fmt.Errorf("专辑更新时间失败: %w", err)
			}
			log.Printf("已更新 %d 张专辑的更新时间", len(changedAlbums))
		}
	}

	return nil
}

func (uc *AudioProcessingUsecase) processMusicWordCloud(
	ctx context.Context,
) error {
	// 1. 删除现有索引
	if err := uc.wordCloudRepo.DropAllIndex(ctx); err != nil {
		// 非致命错误，记录日志但继续执行
		log.Printf("索引删除失败: %v", err)
	}

	// 2. 清空当前词云数据
	if err := uc.wordCloudRepo.AllDelete(ctx); err != nil {
		return fmt.Errorf("词云数据清空失败: %w", err)
	}

	// 3. 获取原始高频词数据
	allWords, err := uc.mediaRepo.GetHighFrequencyWords(ctx, 100)
	if err != nil {
		return fmt.Errorf("高频词获取失败: %w", err)
	}

	// 4. 数据处理流水线（排序+截取TopN）
	sort.Slice(allWords, func(i, j int) bool {
		return allWords[i].Count > allWords[j].Count
	})

	topN := 50
	if len(allWords) < topN {
		topN = len(allWords)
	}
	topWords := allWords[:topN]

	// 5. 构建最终结果集（使用指针切片）
	results := make([]*scene_audio_db_models.WordCloudMetadata, 0, topN)
	for i, word := range topWords {
		results = append(results, &scene_audio_db_models.WordCloudMetadata{
			ID:    primitive.NewObjectID(),
			Type:  word.Type,
			Name:  word.Name,
			Count: word.Count,
			Rank:  i + 1,
		})
	}

	// 6. 批量保存（直接传递指针切片）
	if _, err := uc.wordCloudRepo.BulkUpsert(ctx, results); err != nil {
		return fmt.Errorf("词云数据保存失败: %w", err)
	}

	// 7. 重建索引（补充unique参数）
	if err := uc.wordCloudRepo.CreateIndex(ctx, "name", true); err != nil {
		log.Printf("索引重建警告: %v", err)
	}

	return nil
}

func (uc *AudioProcessingUsecase) processFile(
	ctx context.Context,
	res *scene_audio_db_models.CueConfig,
	path string,
	libraryPath string,
	coverTempPath string,
	lyricsTempPathWithTimestamp string,
	libraryFolderID primitive.ObjectID,
	wg *sync.WaitGroup,
	errChan chan<- error,
	taskProg *domain_util.TaskProgress,
) {
	defer func() {
		// 修复：更新任务级别的处理计数器
		atomic.AddInt32(&taskProg.ProcessedFiles, 1)
		wg.Done()
	}()

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

		if mediaFile != nil && mediaFile.Title == "" {
			mediaFile.Title = "Unknown Title"
			mediaFile.OrderTitle = "Unknown Title"
			mediaFile.SortTitle = "Unknown Title"
		}
		if mediaFileCue != nil && mediaFileCue.Title == "" {
			mediaFileCue.Title = "Unknown Title"
		}

		if err := uc.processAudioHierarchy(ctx, artists, album, mediaFile, mediaFileCue); err != nil {
			return
		}

		if err := uc.processAudioCover(
			ctx,
			mediaFile,
			album,
			artists,
			mediaFileCue,
			path,
			coverTempPath,
		); err != nil {
			errChan <- fmt.Errorf("封面存储失败 %s: %w", path, err)
			return
		}

		if err = uc.processAudioLyrics(
			ctx,
			mediaFile,
			mediaFileCue,
			path,
			lyricsTempPathWithTimestamp,
		); err != nil {
			errChan <- fmt.Errorf("歌词存储失败 %s: %w", path, err)
			return
		}
	}
}

func (uc *AudioProcessingUsecase) createMetadataBasicInfo(
	path string,
	libraryFolderID primitive.ObjectID,
) (*domain_file_entity.FileMetadata, error) {
	// 1. 先查询是否已存在该路径文件
	existingFile, err := uc.fileRepo.FindByPath(context.Background(), path)
	if err != nil && !errors.Is(err, driver.ErrNoDocuments) {
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

func (uc *AudioProcessingUsecase) processAudioCover(
	ctx context.Context,
	media *scene_audio_db_models.MediaFileMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	artists []*scene_audio_db_models.ArtistMetadata,
	mediaCue *scene_audio_db_models.MediaFileCueMetadata,
	path string,
	coverBasePath string,
) error {
	if mediaCue != nil {
		// 处理CUE文件的封面逻辑（保持不变）
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
		// 创建封面存储目录
		mediaCoverDir := filepath.Join(coverBasePath, "media", media.ID.Hex())
		if err := os.MkdirAll(mediaCoverDir, 0755); err != nil {
			return fmt.Errorf("媒体目录创建失败: %w", err)
		}

		var coverPath string
		defer func() {
			if coverPath == "" {
				if err := os.RemoveAll(mediaCoverDir); err != nil {
					log.Printf("[WARN] 媒体目录删除失败 | 路径:%s | 错误:%v", mediaCoverDir, err)
				}
			}
		}()

		// 1. 优先检查本地已存在的封面文件[2,3](@ref)
		coverPath = uc.findLocalCover(path, mediaCoverDir)

		// 2. 如果未找到本地封面，尝试读取内嵌封面
		if coverPath == "" {
			file, err := os.Open(path)
			if err != nil {
				log.Printf("[WARN] 文件打开失败: %v", err)
			} else {
				defer func(file *os.File) {
					err := file.Close()
					if err != nil {
						return
					}
				}(file)
				if metadata, err := tag.ReadFrom(file); err == nil {
					coverPath = uc.extractEmbeddedCover(metadata, mediaCoverDir)
				}
			}
		}

		// 3. 如果仍未获取到封面，使用Ffmpeg提取内嵌封面[1,3,6](@ref)
		if coverPath == "" {
			coverPath = uc.extractCoverWithFFmpeg(path, mediaCoverDir)
		}

		// 更新媒体封面信息
		mediaUpdate := bson.M{
			"$set": bson.M{
				"medium_image_url": coverPath,
				"has_cover_art":    coverPath != "",
			},
		}
		if _, err := uc.mediaRepo.UpdateByID(ctx, media.ID, mediaUpdate); err != nil {
			return fmt.Errorf("媒体更新失败: %w", err)
		}

		// 4. 处理专辑封面（如果存在）
		if album != nil && coverPath != "" {
			uc.processAlbumCover(album, coverPath, coverBasePath, ctx)
		}

		// 5. 处理艺术家封面（如果存在）
		if coverPath != "" && !media.Compilation {
			artistCoverBaseDir := filepath.Join(coverBasePath, "artist")
			if err := os.MkdirAll(artistCoverBaseDir, 0755); err != nil {
				log.Printf("[WARN] 艺术家封面基础目录创建失败: %v", err)
			} else {
				for _, artist := range artists {
					if artist != nil {
						uc.processArtistCover(artist, coverPath, artistCoverBaseDir, ctx)
					}
				}
			}
		}
	}

	return nil
}

func (uc *AudioProcessingUsecase) processAudioLyrics(
	ctx context.Context,
	media *scene_audio_db_models.MediaFileMetadata,
	mediaCue *scene_audio_db_models.MediaFileCueMetadata,
	path string,
	lyricsTempPath string,
) error {
	// 1. 确保歌词根目录存在
	if _, err := os.Stat(lyricsTempPath); os.IsNotExist(err) {
		if err := os.MkdirAll(lyricsTempPath, 0755); err != nil {
			return fmt.Errorf("歌词根目录创建失败: %w", err)
		}
	}

	// 2. 处理常规媒体文件歌词
	if media != nil {
		if err := uc.processMediaLyrics(ctx, media, path, lyricsTempPath); err != nil {
			return fmt.Errorf("媒体文件歌词处理失败: %w", err)
		}
	}

	// 3. 处理CUE文件歌词
	if mediaCue != nil {
		if err := uc.processCueLyrics(ctx, mediaCue, lyricsTempPath); err != nil {
			return fmt.Errorf("CUE文件歌词处理失败: %w", err)
		}
	}

	return nil
}

// 处理常规媒体文件歌词（优化后）
func (uc *AudioProcessingUsecase) processMediaLyrics(
	ctx context.Context,
	media *scene_audio_db_models.MediaFileMetadata,
	path string,
	lyricsTempPath string,
) error {
	// 1. 精准定位同名歌词文件（核心优化）
	audioDir := filepath.Dir(path)
	audioName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	lrcPath := filepath.Join(audioDir, audioName+".lrc") // 直接构建路径避免全局扫描[4,7](@ref)

	// 2. 处理外部歌词文件（精准匹配）
	if _, err := os.Stat(lrcPath); err == nil {
		destFileName, err := uc.generateLyricFilename(media, lyricsTempPath, "lrc")
		if err != nil {
			return fmt.Errorf("歌词文件名生成失败: %w", err)
		}
		if destFileName == "" {
			return nil // 跳过情况
		}

		destPath := filepath.Join(lyricsTempPath, destFileName)

		// 冲突检测（避免覆盖）
		if _, err := os.Stat(destPath); err == nil {
			return nil
		}

		// 保存元数据到数据库
		artist, title := sanitizeMetadata(media.Artist, media.Title)
		if _, err := uc.lyricsFileRepo.UpdateLyricsFilePath(ctx, artist, title, "lrc", destPath); err != nil {
			log.Printf("歌词路径保存失败: %v", err)
		}

		// 安全复制文件（加锁防止并发冲突）[5](@ref)
		uc.scanMutex.Lock()
		defer uc.scanMutex.Unlock()
		if err := uc.copyLyricsFile(lrcPath, destPath); err != nil {
			return fmt.Errorf("歌词复制失败: %w", err)
		}
	}

	// 3. 处理内嵌歌词（独立流程）
	if media.Lyrics != "" {
		destFileName, err := uc.generateLyricFilename(media, lyricsTempPath, "embedded")
		if err != nil {
			return fmt.Errorf("歌词文件名生成失败: %w", err)
		}
		if destFileName == "" {
			return nil
		}

		destPath := filepath.Join(lyricsTempPath, destFileName)

		// 保存元数据到数据库
		artist, title := sanitizeMetadata(media.Artist, media.Title)
		if _, err := uc.lyricsFileRepo.UpdateLyricsFilePath(ctx, artist, title, "embedded", destPath); err != nil {
			log.Printf("歌词路径保存失败: %v", err)
		}

		// 原子写入（避免部分写入）
		if err := atomicWriteFile(destPath, []byte(media.Lyrics), 0644); err != nil {
			return fmt.Errorf("歌词写入失败: %w", err)
		}
	}

	return nil
}

// 辅助函数：原子写入文件（防止写入中断导致损坏）
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, perm); err != nil {
		return err
	}
	return os.Rename(tempPath, path) // 原子操作
}

// 辅助函数：清洗元数据特殊字符
func sanitizeMetadata(artist, title string) (string, string) {
	if artist == "" {
		artist = "UnknownArtist"
	}
	if title == "" {
		title = "UnknownTitle"
	}

	// 替换路径敏感字符[6](@ref)
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_")
	return replacer.Replace(artist), replacer.Replace(title)
}

// 处理CUE文件歌词（优化后）
func (uc *AudioProcessingUsecase) processCueLyrics(
	ctx context.Context,
	mediaCue *scene_audio_db_models.MediaFileCueMetadata,
	lyricsTempPath string,
) error {
	// 处理外部歌词文件（复制CUE文件目录所有.lrc文件）
	cueDir := filepath.Dir(mediaCue.CueResources.CuePath)
	var cueLrcFiles []string
	if entries, err := os.ReadDir(cueDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lrc") {
				cueLrcFiles = append(cueLrcFiles, filepath.Join(cueDir, entry.Name()))
			}
		}
	}

	// 复制所有找到的.lrc文件
	for _, srcPath := range cueLrcFiles {
		destFileName, err := uc.generateLyricFilename(mediaCue, lyricsTempPath, "lrc")
		if err != nil {
			log.Printf("文件名生成失败: %v", err)
			continue
		}

		if destFileName == "" { // 处理跳过情况
			continue
		}

		destPath := filepath.Join(lyricsTempPath, destFileName)

		if _, err := os.Stat(destPath); err == nil { // 冲突检测
			continue
		}

		// 保存歌词信息到数据库
		artist := mediaCue.Performer
		title := mediaCue.Title
		if artist == "" {
			artist = "UnknownArtist"
		}
		if title == "" {
			title = "UnknownTitle"
		}

		// 使用仓库保存歌词路径
		if _, err := uc.lyricsFileRepo.UpdateLyricsFilePath(ctx, artist, title, "lrc", destPath); err != nil {
			log.Printf("歌词路径保存失败: %v", err)
		}

		if err := uc.copyLyricsFile(srcPath, destPath); err != nil {
			log.Printf("CUE歌词复制失败: %s -> %s | %v", srcPath, destPath, err)
		}
	}

	return nil
}

// 生成标准化歌词文件名（统一函数）
func (uc *AudioProcessingUsecase) generateLyricFilename(
	media interface{},
	lyricsTempPath string,
	sourceType string,
) (string, error) {
	// 1. 提取艺术家和标题信息
	var artist, title string
	switch m := media.(type) {
	case *scene_audio_db_models.MediaFileMetadata:
		artist = m.Artist
		title = m.Title
	case *scene_audio_db_models.MediaFileCueMetadata:
		artist = m.Performer
		title = m.Title
	default:
		return "", fmt.Errorf("不支持的媒体类型")
	}

	// 2. 基础文件名生成
	if artist == "" {
		artist = "UnknownArtist"
	}
	if title == "" {
		title = "UnknownTitle"
	}

	// 替换特殊字符防止路径问题
	artist = strings.ReplaceAll(artist, "/", "-")
	title = strings.ReplaceAll(title, "/", "-")

	baseName := fmt.Sprintf("%s - %s", artist, title)
	fileName := baseName + ".lrc"

	// 3. 处理embedded类型 - 存在冲突直接退出
	if sourceType == "embedded" {
		fullPath := filepath.Join(lyricsTempPath, fileName)
		if _, err := os.Stat(fullPath); err == nil {
			return "", nil // 文件存在，返回空文件名表示跳过
		}
		return fileName, nil // 文件不存在，返回基础文件名
	}

	// 4. 处理lrc类型 - 完整冲突检测逻辑
	needRandom := artist == "UnknownArtist" || title == "UnknownTitle"
	randReader := rand.Reader
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 需要随机数时生成后缀
		if needRandom || attempt > 0 {
			randSuffix := make([]byte, 4)
			if _, err := randReader.Read(randSuffix); err != nil {
				return "", fmt.Errorf("随机数生成失败: %w", err)
			}
			fileName = fmt.Sprintf("%s-%X.lrc", baseName, randSuffix)
		}

		// 检查文件是否存在
		fullPath := filepath.Join(lyricsTempPath, fileName)
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				return fileName, nil // 文件不存在，返回可用文件名
			}
			return "", fmt.Errorf("文件检查失败: %w", err)
		}
	}

	// 5. 超过最大尝试次数时使用时间戳
	return fmt.Sprintf("%s-%d.lrc", baseName, time.Now().UnixNano()), nil
}

// 安全复制歌词文件（基于io.Copy）
func (uc *AudioProcessingUsecase) copyLyricsFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			log.Printf("文件关闭失败: %s | %v", src, err)
		}
	}(srcFile)

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {
			log.Printf("文件关闭失败: %s | %v", src, err)
		}
	}(destFile)

	_, err = io.Copy(destFile, srcFile)
	return err
}

// 在音频文件同目录查找封面文件[2](@ref)
func (uc *AudioProcessingUsecase) findLocalCover(audioPath, targetDir string) string {
	dir := filepath.Dir(audioPath)
	baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))

	// 可能的封面文件名模式
	coverPatterns := []string{
		"cover.jpg", "cover.png", "cover.jpeg", // 通用封面名
		baseName + ".jpg", baseName + ".png", baseName + ".jpeg", // 与音频同名的封面
		"folder.jpg", "folder.png", // Windows常见的封面名
	}

	for _, pattern := range coverPatterns {
		srcPath := filepath.Join(dir, pattern)
		if _, err := os.Stat(srcPath); err == nil {
			destPath := filepath.Join(targetDir, "cover"+filepath.Ext(srcPath))
			if err := uc.copyCoverFile(srcPath, destPath); err == nil {
				return destPath
			}
		}
	}
	return ""
}

// 复制封面文件
func (uc *AudioProcessingUsecase) copyCoverFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			return
		}
	}(srcFile)

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {
			return
		}
	}(destFile)

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}
	return nil
}

// 清理旧歌词目录（保留最新的5个）
func (uc *AudioProcessingUsecase) cleanupOldLyricsFoldersAsync(lyricsTempPath string) {
	// 1. 异步恢复机制防止panic崩溃[4](@ref)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[异步清理] 发生未预期错误: %v", r)
		}
	}()

	// 2. 执行实际清理逻辑
	if err := uc.cleanupOldLyricsFolders(lyricsTempPath); err != nil {
		log.Printf("[异步清理] 旧歌词目录清理失败: %v", err)
	}
}

// 优化后的清理逻辑（保持原有功能）
func (uc *AudioProcessingUsecase) cleanupOldLyricsFolders(lyricsTempPath string) error {
	// 1. 读取目录下所有项目
	entries, err := os.ReadDir(lyricsTempPath)
	if err != nil {
		return fmt.Errorf("读取歌词目录失败: %w", err)
	}

	// 2. 过滤时间戳目录
	var folders []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && domain_util.IsTimestampFolder(entry.Name()) {
			folders = append(folders, entry)
		}
	}

	// 3. 检查是否需要清理（超过10个才清理）
	if len(folders) <= 10 {
		log.Printf("[异步清理] 无需清理（当前目录数: %d）", len(folders))
		return nil
	}

	// 4. 按时间戳排序（从旧到新）[1](@ref)
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name() < folders[j].Name()
	})

	// 5. 保留最新的5个，其余删除
	deleteCount := len(folders) - 5
	for i := 0; i < deleteCount; i++ {
		folderPath := filepath.Join(lyricsTempPath, folders[i].Name())
		if err := os.RemoveAll(folderPath); err != nil {
			log.Printf("[异步清理] 删除失败 %s: %v", folderPath, err)
		} else {
			log.Printf("[异步清理] 已删除旧目录: %s", folderPath)
		}
	}
	return nil
}

// 从标签元数据提取内嵌封面[1](@ref)
func (uc *AudioProcessingUsecase) extractEmbeddedCover(metadata tag.Metadata, mediaCoverDir string) string {
	if pic := metadata.Picture(); pic != nil && len(pic.Data) > 0 {
		coverPath := filepath.Join(mediaCoverDir, "cover.jpg")
		if err := os.WriteFile(coverPath, pic.Data, 0644); err == nil {
			return coverPath
		}
	}
	return ""
}

// 使用FFmpeg提取封面（当其他方法失败时）[3,6](@ref)
func (uc *AudioProcessingUsecase) extractCoverWithFFmpeg(audioPath, mediaCoverDir string) string {
	// 1. 设置输出路径为PNG格式（WAV封面多为PNG）
	coverPath := filepath.Join(mediaCoverDir, "cover.png")

	// 2. 构建优化的FFmpeg命令
	cmd := ffmpeggo.Input(audioPath).
		Output(coverPath, ffmpeggo.KwArgs{
			"an":     "",     // 丢弃音频流
			"vcodec": "copy", // 直接复制视频流（封面数据）
			"y":      "",     // 覆盖现有文件
		}).
		OverWriteOutput()

	// 3. 执行并记录详细命令
	compiledCmd := cmd.Compile()
	log.Printf("执行FFmpeg封面提取命令: %s", strings.Join(compiledCmd.Args, " "))

	// 4. 运行命令并处理错误
	if err := cmd.Run(); err != nil {
		// 5. 特定错误处理
		if strings.Contains(err.Error(), "No video stream") {
			log.Printf("[WARN] 文件无内嵌封面: %s", audioPath)
		} else {
			log.Printf("[ERROR] FFmpeg封面提取失败: %v", err)
		}
		return ""
	}

	// 6. 验证结果
	if info, err := os.Stat(coverPath); err != nil || info.Size() == 0 {
		log.Printf("[WARN] 封面生成失败 | 错误:%v", err)
		return ""
	}

	return coverPath
}

// 处理专辑封面[1](@ref)
func (uc *AudioProcessingUsecase) processAlbumCover(
	album *scene_audio_db_models.AlbumMetadata,
	coverPath, coverBasePath string, ctx context.Context,
) {

	if album.MediumImageURL == "" {
		albumCoverDir := filepath.Join(coverBasePath, "album", album.ID.Hex())
		if err := os.MkdirAll(albumCoverDir, 0755); err != nil {
			log.Printf("[WARN] 专辑封面目录创建失败: %v", err)
			return
		}

		albumCoverPath := filepath.Join(albumCoverDir, "cover.jpg")
		if err := uc.copyCoverFile(coverPath, albumCoverPath); err != nil {
			log.Printf("[WARN] 专辑封面复制失败: %v", err)
			return
		}

		albumUpdate := bson.M{
			"$set": bson.M{
				"medium_image_url": albumCoverPath,
				"has_cover_art":    true,
			},
		}

		if _, err := uc.albumRepo.UpdateByID(ctx, album.ID, albumUpdate); err != nil {
			log.Printf("[WARN] 专辑元数据更新失败: %v", err)
		}
	}
}

// 处理艺术家封面[1](@ref)
func (uc *AudioProcessingUsecase) processArtistCover(
	artist *scene_audio_db_models.ArtistMetadata,
	coverPath, artistBaseDir string,
	ctx context.Context,
) {
	artistCoverDir := filepath.Join(artistBaseDir, artist.ID.Hex())
	artistCoverPath := filepath.Join(artistCoverDir, "cover.jpg")

	if _, err := os.Stat(artistCoverDir); errors.Is(err, fs.ErrNotExist) {
		// 1. 创建艺术家专属目录 (关键修复)
		if err := os.MkdirAll(artistCoverDir, 0755); err != nil {
			log.Printf("[ERROR] 艺术家封面目录创建失败 | 艺术家:%s | 错误:%v", artist.ID.Hex(), err)
			return
		}

		// 2. 复制封面文件
		if err := uc.copyCoverFile(coverPath, artistCoverPath); err != nil {
			log.Printf("[ERROR] 艺术家封面复制失败 | 艺术家:%s | 错误:%v", artist.ID.Hex(), err)
			return
		}

		// 3. 更新艺术家元数据
		artistUpdate := bson.M{
			"$set": bson.M{
				"medium_image_url": artistCoverPath,
				"has_cover_art":    true,
			},
		}

		if _, err := uc.artistRepo.UpdateByID(ctx, artist.ID, artistUpdate); err != nil {
			log.Printf("[ERROR] 艺术家元数据更新失败 | 艺术家:%s | 错误:%v", artist.ID.Hex(), err)
		}
	}

	return
}

func (uc *AudioProcessingUsecase) processAudioHierarchy(ctx context.Context,
	artists []*scene_audio_db_models.ArtistMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	mediaFile *scene_audio_db_models.MediaFileMetadata,
	mediaFileCue *scene_audio_db_models.MediaFileCueMetadata,
) (err error) {
	// 关键依赖检查
	if uc.mediaRepo == nil || uc.artistRepo == nil || uc.albumRepo == nil || uc.mediaCueRepo == nil {
		log.Print("音频仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	// 修复：检查mediaFile和mediaFileCue同时为nil的情况
	if mediaFile == nil && mediaFileCue == nil {
		log.Print("媒体文件元数据为空")
		return fmt.Errorf("mediaFile and mediaFileCue cannot both be nil")
	}

	// 直接保存无关联数据
	if artists == nil && album == nil {
		if mediaFile != nil {
			if err := uc.updateAudioMediaMetadata(ctx, mediaFile); err != nil {
				log.Printf("单曲处理失败: %s | %v", mediaFile.Path, err)
				return fmt.Errorf("单曲处理失败 | 名称:%s | 原因:%w", mediaFile.Path, err)
			}
		}
		if mediaFileCue != nil {
			if err := uc.updateAudioMediaCueMetadata(ctx, mediaFileCue); err != nil {
				return fmt.Errorf("单曲处理失败 | 名称:%s | 原因:%w", mediaFileCue.Path, err)
			}
			//if _, err := uc.mediaCueRepo.Upsert(ctx, mediaFileCue); err != nil {
			//	cuePath := "unknown path"
			//	if len(mediaFileCue.CueResources.CuePath) > 0 {
			//		cuePath = mediaFileCue.CueResources.CuePath
			//	}
			//	log.Printf("CUE文件保存失败: %s | %v", cuePath, err)
			//}
		}
		return nil
	}

	// 处理艺术家
	if artists != nil {
		for _, artist := range artists {
			// 修复：添加空指针检查
			if artist == nil {
				log.Print("艺术家元数据为空")
				continue
			}
			if artist.Name == "" {
				log.Print("艺术家名称为空")
				artist.Name = "Unknown"
			} else {
				artist.NamePinyin = pinyin.LazyConvert(artist.Name, nil)
				artist.NamePinyinFull = strings.Join(artist.NamePinyin, "")
			}
			if err := uc.updateAudioArtistMetadata(ctx, artist); err != nil {
				log.Printf("艺术家处理失败: %s | %v", artist.Name, err)
				return fmt.Errorf("艺术家处理失败 | 原因:%w", err)
			}
		}
	}

	// 处理专辑 - 修复：添加空专辑指针检查
	if album != nil {
		if album.Name == "" {
			log.Print("专辑名称为空")
			album.Name = "Unknown"
		}
		if album.Artist == "" {
			log.Print("专辑艺术家名称为空")
			album.Artist = "Unknown Artist"
		} else {
			album.ArtistPinyin = pinyin.LazyConvert(album.Artist, nil)
			album.ArtistPinyinFull = strings.Join(album.ArtistPinyin, "")
		}
		if err := uc.updateAudioAlbumMetadata(ctx, album); err != nil {
			log.Printf("专辑处理失败: %s | %v", album.Name, err)
			return fmt.Errorf("专辑处理失败 | 名称:%s | 原因:%w", album.Name, err)
		}
	}

	// 保存媒体文件
	if mediaFile != nil {
		if mediaFile.Artist == "" {
			log.Print("媒体文件艺术家名称为空")
			mediaFile.Artist = "Unknown Artist"
		} else {
			mediaFile.ArtistPinyin = pinyin.LazyConvert(mediaFile.Artist, nil)
			mediaFile.ArtistPinyinFull = strings.Join(mediaFile.ArtistPinyin, "")
		}
		if err := uc.updateAudioMediaMetadata(ctx, mediaFile); err != nil {
			log.Printf("单曲处理失败: %s | %v", mediaFile.Path, err)
			return fmt.Errorf("单曲处理失败 | 名称:%s | 原因:%w", mediaFile.Path, err)
		}
	}

	// 修复：分离媒体文件保存逻辑并添加空指针保护
	if mediaFileCue != nil {
		if mediaFileCue.Performer == "" {
			log.Print("CUE文件艺术家名称为空")
			mediaFileCue.Performer = "Unknown Artist"
		} else {
			mediaFileCue.PerformerPinyin = pinyin.LazyConvert(mediaFileCue.Performer, nil)
			mediaFileCue.PerformerPinyinFull = strings.Join(mediaFileCue.PerformerPinyin, "")
		}
		if err := uc.updateAudioMediaCueMetadata(ctx, mediaFileCue); err != nil {
			return fmt.Errorf("单曲处理失败 | 名称:%s | 原因:%w", mediaFileCue.Path, err)
		}
		//if _, err := uc.mediaCueRepo.Upsert(ctx, mediaFileCue); err != nil {
		//	cuePath := "unknown cue path"
		//	if len(mediaFileCue.CueResources.CuePath) > 0 {
		//		cuePath = mediaFileCue.CueResources.CuePath
		//	}
		//
		//	errorInfo := cuePath
		//	if album != nil {
		//		errorInfo += fmt.Sprintf(" 专辑:%s", album.Name)
		//	}
		//
		//	log.Printf("CUE最终保存失败: %s | %v", errorInfo, err)
		//}
	}

	// 异步统计更新
	go uc.updateAudioArtistAndAlbumStatistics(artists, album, mediaFile, mediaFileCue)
	return nil
}

func (uc *AudioProcessingUsecase) updateAudioArtistAndAlbumStatistics(
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

func (uc *AudioProcessingUsecase) updateAudioArtistMetadata(ctx context.Context, artist *scene_audio_db_models.ArtistMetadata) error {
	if uc.artistRepo == nil {
		log.Print("艺术家仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	existing, err := uc.artistRepo.GetByID(ctx, artist.ID)
	if err != nil {
		log.Printf("名称查询错误: %v", err)
	}
	// 仅同步ID，因为album会随着更新而新增字段
	if existing != nil {
		artist.ID = existing.ID
		artist.MediumImageURL = existing.MediumImageURL
		artist.HasCoverArt = existing.HasCoverArt
		artist.CreatedAt = existing.CreatedAt
		artist.UpdatedAt = existing.UpdatedAt // 先不更新该艺术家，媒体库重构阶段判断是否更新
	}
	// 再次插入，将版本更新的字段覆盖到现有文档
	if err := uc.artistRepo.Upsert(ctx, artist); err != nil {
		log.Printf("艺术家创建失败: %s | %v", artist.Name, err)
		return err
	}

	return nil
}

func (uc *AudioProcessingUsecase) updateAudioAlbumMetadata(ctx context.Context, album *scene_audio_db_models.AlbumMetadata) error {
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
		if album.UpdatedAt != existing.UpdatedAt {
			if err := uc.albumRepo.Upsert(ctx, album); err != nil {
				log.Printf("专辑创建失败: %s | %v", album.Name, err)
				return err
			}
		}
	} else {
		if err := uc.albumRepo.Upsert(ctx, album); err != nil {
			log.Printf("专辑创建失败: %s | %v", album.Name, err)
			return err
		}
	}

	return nil
}

func (uc *AudioProcessingUsecase) updateAudioMediaMetadata(ctx context.Context, media *scene_audio_db_models.MediaFileMetadata) error {
	if uc.mediaRepo == nil {
		log.Print("专辑仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	existing, err := uc.mediaRepo.GetByID(ctx, media.ID)
	if err != nil {
		log.Printf("组合查询错误: %v", err)
	}

	if existing != nil {
		if media.UpdatedAt != existing.UpdatedAt {
			if _, err := uc.mediaRepo.Upsert(ctx, media); err != nil {
				log.Printf("单曲创建失败: %s | %v", media.Path, err)
				return err
			}
		}
	} else {
		if _, err := uc.mediaRepo.Upsert(ctx, media); err != nil {
			log.Printf("单曲创建失败: %s | %v", media.Path, err)
			return err
		}
	}

	return nil
}

func (uc *AudioProcessingUsecase) updateAudioMediaCueMetadata(ctx context.Context, mediaCue *scene_audio_db_models.MediaFileCueMetadata) error {
	if uc.mediaCueRepo == nil {
		log.Print("专辑仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	existing, err := uc.mediaCueRepo.GetByID(ctx, mediaCue.ID)
	if err != nil {
		log.Printf("组合查询错误: %v", err)
	}

	if existing != nil {
		if mediaCue.UpdatedAt != existing.UpdatedAt {
			if _, err := uc.mediaCueRepo.Upsert(ctx, mediaCue); err != nil {
				log.Printf("单曲创建失败: %s | %v", mediaCue.Path, err)
				return err
			}
		}
	} else {
		if _, err := uc.mediaCueRepo.Upsert(ctx, mediaCue); err != nil {
			log.Printf("单曲创建失败: %s | %v", mediaCue.Path, err)
			return err
		}
	}

	return nil
}

func (uc *AudioProcessingUsecase) shouldProcess(path string, folderType int) bool {
	fileType, err := uc.detector.DetectMediaType(path)
	if err != nil {
		log.Printf("文件类型检测失败: %v", err)
		return false
	}
	if folderType == 1 && fileType == domain_file_entity.Audio {
		return true
	}
	if folderType == 2 && fileType == domain_file_entity.Video {
		return true
	}
	if folderType == 3 && fileType == domain_file_entity.Image {
		return true
	}
	if folderType == 4 && fileType == domain_file_entity.Document {
		return true
	}
	return false
}

// 使用快速目录统计函数
func fastCountFilesInFolder(rootPath string) (int, error) {
	count := 0
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}
