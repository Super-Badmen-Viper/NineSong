package scene_audio

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
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

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	usercase_audio_util "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_util"
	"github.com/mozillazg/go-pinyin"
	ffmpeggo "github.com/u2takey/ffmpeg-go"
	driver "go.mongodb.org/mongo-driver/mongo"

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
	MediaRepo        scene_audio_db_interface.MediaFileRepository
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
	commonUtil *AudioCommonUtil
	statistics *AudioStatisticsUsecase
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
	workerCount := runtime.NumCPU() * 4
	commonUtil := NewAudioCommonUtil(detector)
	statistics := NewAudioStatisticsUsecase(
		artistRepo,
		albumRepo,
		mediaRepo,
		mediaCueRepo,
	)

	return &AudioProcessingUsecase{
		db:             db,
		fileRepo:       fileRepo,
		folderRepo:     folderRepo,
		detector:       detector,
		workerPool:     make(chan struct{}, workerCount),
		scanTimeout:    time.Duration(timeoutMinutes) * time.Minute,
		artistRepo:     artistRepo,
		albumRepo:      albumRepo,
		MediaRepo:      mediaRepo,
		tempRepo:       tempRepo,
		mediaCueRepo:   mediaCueRepo,
		wordCloudRepo:  wordCloudRepo,
		lyricsFileRepo: lyricsFileRepo,
		commonUtil:     commonUtil,
		statistics:     statistics,
	}
}

func (uc *AudioProcessingUsecase) ProcessMusicDirectory(
	ctx context.Context,
	libraryFolders []*domain_file_entity.LibraryFolderMetadata,
	taskProg *domain_util.TaskProgress,
	libraryTraversal bool, // 是否遍历传入的媒体库目录
	libraryStatistics bool, // 是否统计传入的媒体库目录
	libraryRefactoring bool, // 是否重构传入的媒体库目录（覆盖所有元数据）
	filesToProcess []string, // 指定要处理的文件集合，为空则遍历整个目录
) error {
	var finalErr error

	// 控制日志输出开关，默认关闭详细日志
	logOutputEnabled := false

	// 初始化扫描阶段权重（根据实际耗时比例设置默认值）
	uc.scanStageWeights = struct {
		traversal   float32
		statistics  float32
		refactoring float32
	}{}

	// 根据不同的扫描模式设置不同的阶段权重
	if !libraryTraversal && libraryStatistics && !libraryRefactoring {
		// 仅统计模式
		uc.scanStageWeights = struct {
			traversal   float32
			statistics  float32
			refactoring float32
		}{
			traversal:   0.1, // 文件处理阶段权重
			statistics:  0.8, // 统计阶段权重
			refactoring: 0.1, // 重构阶段权重
		}
	} else if libraryTraversal && libraryStatistics && libraryRefactoring {
		// 完整扫描模式
		uc.scanStageWeights = struct {
			traversal   float32
			statistics  float32
			refactoring float32
		}{
			traversal:   0.6, // 文件处理阶段权重
			statistics:  0.2, // 统计阶段权重
			refactoring: 0.2, // 重构阶段权重
		}
	} else {
		// 默认模式
		uc.scanStageWeights = struct {
			traversal   float32
			statistics  float32
			refactoring float32
		}{
			traversal:   0.7, // 文件处理阶段权重
			statistics:  0.2, // 统计阶段权重
			refactoring: 0.1, // 重构阶段权重
		}
	}

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
	processedFilesMap := make(map[string]bool) // 用于去重

	if len(filesToProcess) > 0 {
		// 如果指定了要处理的文件集合，统计实际需要处理的文件数
		for _, path := range filesToProcess {
			// 检查文件是否应该被处理
			if uc.shouldProcess(path, 1) {
				// 检查文件是否属于任何一个libraryFolder
				for _, folder := range libraryFolders {
					libraryFolderPath := strings.Replace(folder.FolderPath, "/", "\\", -1)
					if !strings.HasSuffix(libraryFolderPath, "\\") {
						libraryFolderPath += "\\"
					}

					if strings.HasPrefix(strings.Replace(path, "/", "\\", -1), libraryFolderPath) {
						// 规范化文件路径，避免重复计数
						normalizedPath := filepath.ToSlash(filepath.Clean(path))
						if !processedFilesMap[normalizedPath] {
							totalFiles++
							processedFilesMap[normalizedPath] = true
						}
						break
					}
				}
			}
		}
	} else {
		// 否则统计目录中的文件数
		for _, folder := range libraryFolders {
			count, err := uc.fastCountFilesInFolder(folder.FolderPath, folder.FolderType)
			if err != nil {
				log.Printf("统计文件数失败: %v", err)
				continue
			}
			totalFiles += count
		}
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

	// 声明folderInfo、excludeWavs、cueResourcesMap到外层作用域
	var folderInfo struct {
		libraryFolderID        primitive.ObjectID
		libraryFolderPath      string
		libraryFolderFileCount int
	}
	var excludeWavs map[string]struct{}
	var cueResourcesMap map[string]*scene_audio_db_models.CueConfig

	// 扫描前记录数据库统计字段，用以后续比对决定是否设置UpdatedAt
	artistCountsBeforeScanning, err := uc.artistRepo.GetAllCounts(ctx)
	//albumCountsBeforeScanning, err := uc.albumRepo.GetAllCounts(ctx)
	// 扫描前重置数据库统计字段 - 只有在重构模式下才重置
	if libraryRefactoring {
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

	// 并发处理管道
	var wgFile sync.WaitGroup
	errChan := make(chan error, 100)

	// 路径收集容器
	regularAudioPaths = make([]string, 0) // 常规音频路径
	cueAudioPaths = make([]string, 0)     // CUE音频路径
	// 路径去重容器
	addedCuePaths := make(map[string]bool)
	addedRegularPaths := make(map[string]bool)

	// 如果指定了要处理的文件集合，直接处理这些文件，不进行目录遍历
	if len(filesToProcess) > 0 {
		processedFilesMap := make(map[string]bool) // 用于去重
		// 初始化folderInfo
		folderInfo = struct {
			libraryFolderID        primitive.ObjectID
			libraryFolderPath      string
			libraryFolderFileCount int
		}{}

		// 存储需要排除的.wav文件路径
		excludeWavs = make(map[string]struct{})
		cueResourcesMap = make(map[string]*scene_audio_db_models.CueConfig)

		// 当指定了要处理的文件集合时，只收集这些文件的路径
		for _, path := range filesToProcess {
			// 规范化文件路径，转换为正斜杠格式，与数据库存储格式一致
			normalizedPath := filepath.ToSlash(filepath.Clean(path))
			// 检查文件是否应该被处理
			if uc.shouldProcess(path, 1) {
				// 检查文件扩展名
				ext := strings.ToLower(filepath.Ext(path))
				// 处理.cue文件
				if ext == ".cue" {
					// 查找同名音频文件
					dir := filepath.Dir(path)
					baseName := strings.TrimSuffix(filepath.Base(path), ext)
					audioExts := []string{".wav", ".ape", ".flac"} // 扩展支持的音频格式
					for _, audioExt := range audioExts {
						audioPath := filepath.Join(dir, baseName+audioExt)
						if _, err := os.Stat(audioPath); err == nil {
							// 规范化音频文件路径
							normalizedAudioPath := filepath.ToSlash(filepath.Clean(audioPath))
							if _, existsAdded := addedCuePaths[normalizedAudioPath]; !existsAdded {
								cueAudioPaths = append(cueAudioPaths, normalizedAudioPath)
								addedCuePaths[normalizedAudioPath] = true
							}
							break
						}
					}
				} else {
					// 处理常规音频文件
					if _, existsAdded := addedRegularPaths[normalizedPath]; !existsAdded {
						regularAudioPaths = append(regularAudioPaths, normalizedPath)
						addedRegularPaths[normalizedPath] = true
					}
				}
			}
		}

		for _, path := range filesToProcess {
			// 检查文件是否已经被处理过
			if processedFilesMap[path] {
				continue
			}
			processedFilesMap[path] = true

			// 检查文件是否应该被处理
			if !uc.shouldProcess(path, 1) {
				continue
			}

			// 查找文件所属的libraryFolder
			var folder *domain_file_entity.LibraryFolderMetadata
			for _, f := range libraryFolders {
				libraryFolderPath := strings.Replace(f.FolderPath, "/", "\\", -1)
				if !strings.HasSuffix(libraryFolderPath, "\\") {
					libraryFolderPath += "\\"
				}
				if strings.HasPrefix(strings.Replace(path, "/", "\\", -1), libraryFolderPath) {
					folder = f
					folderInfo.libraryFolderID = f.ID
					folderInfo.libraryFolderPath = libraryFolderPath
					break
				}
			}

			if folder == nil {
				continue // 文件不属于任何libraryFolder，跳过
			}

			ext := strings.ToLower(filepath.Ext(path))

			// 处理.cue文件
			if ext == ".cue" {
				// 收集.cue文件和关联资源
				res := &scene_audio_db_models.CueConfig{CuePath: path}

				// 查找同名音频文件
				audioFound := false
				audioExts := []string{".wav", ".ape", ".flac"} // 扩展支持的音频格式
				dir := filepath.Dir(path)
				baseName := strings.TrimSuffix(filepath.Base(path), ext)
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
					continue
				}

				// 优化：批量收集相关资源文件，减少文件系统访问次数
				dir = filepath.Dir(path)
				// 一次性获取目录下的所有文件，然后在内存中检查是否存在需要的资源文件
				fileInfos, err := os.ReadDir(dir)
				if err == nil {
					// 构建文件名到文件信息的映射
					fileMap := make(map[string]os.DirEntry)
					for _, info := range fileInfos {
						fileMap[strings.ToLower(info.Name())] = info
					}

					// 检查需要的资源文件
					if info, exists := fileMap["back.jpg"]; exists && !info.IsDir() {
						res.BackImage = filepath.Join(dir, "back.jpg")
					}
					if info, exists := fileMap["cover.jpg"]; exists && !info.IsDir() {
						res.CoverImage = filepath.Join(dir, "cover.jpg")
					}
					if info, exists := fileMap["disc.jpg"]; exists && !info.IsDir() {
						res.DiscImage = filepath.Join(dir, "disc.jpg")
					}
					if info, exists := fileMap["list.txt"]; exists && !info.IsDir() {
						res.ListFile = filepath.Join(dir, "list.txt")
					}
					if info, exists := fileMap["log.txt"]; exists && !info.IsDir() {
						res.LogFile = filepath.Join(dir, "log.txt")
					}
				}

				cueResourcesMap[path] = res

				// 关键修复：只收集需要处理的文件路径
				if res.AudioPath != "" {
					// 规范化音频文件路径，确保与数据库中存储的格式一致
					normalizedAudioPath := filepath.ToSlash(filepath.Clean(res.AudioPath))
					if _, existsAdded := addedCuePaths[normalizedAudioPath]; !existsAdded {
						cueAudioPaths = append(cueAudioPaths, normalizedAudioPath)
						addedCuePaths[normalizedAudioPath] = true
					}
				}

				// 仅当开启遍历时才执行文件处理和计数
				if libraryTraversal {
					wgFile.Add(1)
					folderInfo.libraryFolderFileCount++
					go uc.processFile(
						ctx, res,
						res.AudioPath, folderInfo.libraryFolderPath,
						coverTempPath, lyricsTempPathWithTimestamp,
						folder.ID, &wgFile, errChan, taskProg,
					)
				}
			} else {
				// 检查是否是被排除的.wav文件
				if _, excluded := excludeWavs[path]; excluded {
					continue
				}

				// 关键修复：只收集需要处理的文件路径
				// 规范化音频文件路径，确保与数据库中存储的格式一致
				normalizedPath := filepath.ToSlash(filepath.Clean(path))
				if _, existsAdded := addedRegularPaths[normalizedPath]; !existsAdded {
					regularAudioPaths = append(regularAudioPaths, normalizedPath)
					addedRegularPaths[normalizedPath] = true
				}

				// 仅当开启遍历时才执行文件处理和计数
				if libraryTraversal {
					wgFile.Add(1)
					folderInfo.libraryFolderFileCount++
					go uc.processFile(
						ctx, nil,
						path, folderInfo.libraryFolderPath,
						coverTempPath, lyricsTempPathWithTimestamp,
						folder.ID, &wgFile, errChan, taskProg,
					)
				}
			}
		}
	} else {
		// 否则，遍历整个目录
		for _, folder := range libraryFolders {
			libraryFolderPath := strings.Replace(folder.FolderPath, "/", "\\", -1)
			if !strings.HasSuffix(libraryFolderPath, "\\") {
				libraryFolderPath += "\\"
			}

			// 初始化folderInfo
			folderInfo = struct {
				libraryFolderID        primitive.ObjectID
				libraryFolderPath      string
				libraryFolderFileCount int
			}{
				libraryFolderID:        folder.ID,
				libraryFolderPath:      libraryFolderPath,
				libraryFolderFileCount: 0,
			}

			// 初始化excludeWavs和cueResourcesMap
			excludeWavs = make(map[string]struct{})
			cueResourcesMap = make(map[string]*scene_audio_db_models.CueConfig)

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
							// 规范化音频文件路径，确保与数据库中存储的格式一致
							normalizedAudioPath := filepath.ToSlash(filepath.Clean(res.AudioPath))
							if _, existsAdded := addedCuePaths[normalizedAudioPath]; !existsAdded {
								cueAudioPaths = append(cueAudioPaths, normalizedAudioPath)
								addedCuePaths[normalizedAudioPath] = true
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
						// 规范化音频文件路径，确保与数据库中存储的格式一致
						normalizedPath := filepath.ToSlash(filepath.Clean(path))
						if _, existsAdded := addedRegularPaths[normalizedPath]; !existsAdded {
							regularAudioPaths = append(regularAudioPaths, normalizedPath)
							addedRegularPaths[normalizedPath] = true
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
		log.Printf("开始媒体库统计阶段")

		// 更新当前阶段为统计阶段
		uc.scanMutex.Lock()
		uc.scanProgress = uc.scanStageWeights.traversal
		taskProg.Mu.Lock()
		processed := int32(float32(taskProg.TotalFiles) * uc.scanProgress)
		atomic.StoreInt32(&taskProg.ProcessedFiles, processed)
		taskProg.Mu.Unlock()
		uc.scanMutex.Unlock()

		// 根据扫描模式选择统计方式
		// 重构模式：先清零再遍历累加（准确计数）
		// 普通模式：增量更新（不清零）
		var err error
		if libraryRefactoring {
			log.Printf("执行全局重构统计（清零后遍历累加）")
			err = uc.statistics.RecalculateAllStatistics(ctx, taskProg)
		} else {
			log.Printf("执行增量统计更新（不清零）")
			err = uc.statistics.UpdateStatisticsIncremental(ctx, taskProg)
		}

		if err != nil {
			return fmt.Errorf("媒体库统计阶段失败: %w", err)
		}

		// 更新进度
		uc.scanMutex.Lock()
		uc.scanProgress = uc.scanStageWeights.traversal + uc.scanStageWeights.statistics
		processed = int32(float32(taskProg.TotalFiles) * uc.scanProgress)
		taskProg.Mu.Lock()
		atomic.StoreInt32(&taskProg.ProcessedFiles, processed)
		taskProg.Mu.Unlock()
		uc.scanMutex.Unlock()
		log.Printf("媒体库统计阶段完成")
	}

	// 区域3: 媒体库重构（仅当libraryRefactoring为true时执行）
	if libraryRefactoring {
		// 更新当前阶段为重构阶段
		uc.scanMutex.Lock()
		uc.scanProgress = uc.scanStageWeights.traversal + uc.scanStageWeights.statistics
		uc.scanMutex.Unlock()

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

		// 1. 执行艺术家统计 - 无论filesToProcess是否为空，都需要执行艺术家统计
		// 根据filesToProcess是否为空，选择不同的统计方式
		log.Printf("执行常规音频艺术家统计，共 %d 个艺术家", len(artistIDs))

		// 处理常规音频的艺术家统计
		for _, id := range artistIDs {
			artistID := id.Hex()

			// 确保结构体已初始化
			if _, exists := invalidArtistCounts[artistID]; !exists {
				invalidArtistCounts[artistID] = &artistStats{}
			}

			if len(filesToProcess) > 0 {
				// 当指定了要处理的文件集合时，执行详细的艺术家统计
				// 处理主艺术家统计
				mediaCount, err := uc.MediaRepo.InspectMediaCountByArtist(ctx, artistID, regularAudioPaths)
				if err != nil {
					if logOutputEnabled {
						log.Printf("常规音频主艺术家统计失败 (ID %s): %v", artistID, err)
					}
				} else {
					if mediaCount >= 0 {
						invalidArtistCounts[artistID].mediaCount += mediaCount
					} else {
						invalidArtistCounts[artistID].mediaCount++
						deleteArtistCounts[artistID]++
					}
				}

				// 处理合作艺术家统计
				guestMediaCount, err := uc.MediaRepo.InspectGuestMediaCountByArtist(ctx, artistID, regularAudioPaths)
				if err != nil {
					if logOutputEnabled {
						log.Printf("常规音频合作艺术家统计失败 (ID %s): %v", artistID, err)
					}
				} else {
					if guestMediaCount >= 0 {
						invalidArtistCounts[artistID].guestMediaCount += guestMediaCount
					} else {
						invalidArtistCounts[artistID].guestMediaCount++
						deleteArtistCounts[artistID]++
					}
				}
			} else {
				// 完整扫描模式下，直接从数据库获取统计信息
				mediaCount, err := uc.MediaRepo.MediaCountByArtist(ctx, artistID)
				if err != nil {
					if logOutputEnabled {
						log.Printf("获取常规音频主艺术家统计失败 (ID %s): %v", artistID, err)
					}
				} else {
					invalidArtistCounts[artistID].mediaCount = int(mediaCount)
				}

				guestMediaCount, err := uc.MediaRepo.GuestMediaCountByArtist(ctx, artistID)
				if err != nil {
					if logOutputEnabled {
						log.Printf("获取常规音频合作艺术家统计失败 (ID %s): %v", artistID, err)
					}
				} else {
					invalidArtistCounts[artistID].guestMediaCount = int(guestMediaCount)
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
				// 获取专辑中实际存在的媒体文件数量
				actualCount, err := uc.MediaRepo.MediaCountByAlbum(ctx, albumIDStr)
				if err != nil {
					log.Printf("获取专辑媒体数量失败 (ID %s): %v", albumIDStr, err)
					continue
				}

				// 检查专辑中不在有效路径列表中的媒体文件数量
				invalidCount, err := uc.MediaRepo.InspectMediaCountByAlbum(ctx, albumIDStr, regularAudioPaths)
				if err != nil {
					log.Printf("常规音频专辑统计失败 (ID %s): %v", albumIDStr, err)
					continue
				}

				// 计算专辑中有效媒体文件数量，将invalidCount转换为int64类型
				validCount := actualCount - int64(invalidCount)

				if validCount > 0 {
					// 专辑中有有效媒体文件，更新统计信息
					invalidArtistCounts[artistIDStr].albumCount += int(validCount)
					invalidAlbumCounts[albumIDStr] = int(validCount)
				} else {
					// 专辑中没有有效媒体文件，标记为删除
					invalidArtistCounts[artistIDStr].albumCount++
					invalidAlbumCounts[albumIDStr] = 0 // 改为0，表示需要删除
					deleteArtistCounts[artistIDStr]++
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
				count, err := uc.MediaRepo.InspectGuestMediaCountByAlbum(ctx, albumIDStr, regularAudioPaths)
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

		// 3. 清除常规音频的无效专辑（删除统计：有效单曲数为0的专辑）
		invalidAlbumNums := 0
		log.Printf("开始清理无效专辑，共检查 %d 个专辑", len(invalidAlbumCounts))

		// 批量处理无效专辑
		var albumsToDelete []primitive.ObjectID
		for albumID, validCount := range invalidAlbumCounts {
			// 只有当有效媒体文件数量为0时，才标记为删除
			if validCount <= 0 {
				albumObjID, err := primitive.ObjectIDFromHex(albumID)
				if err != nil {
					if logOutputEnabled {
						log.Printf("无效专辑ID (ID %s): %v", albumID, err)
					}
					continue
				}

				// 额外验证：再次检查专辑中是否确实没有有效媒体文件
				actualMediaCount, err := uc.MediaRepo.MediaCountByAlbum(ctx, albumID)
				if err != nil {
					if logOutputEnabled {
						log.Printf("获取专辑媒体数量失败 (ID %s): %v", albumID, err)
					}
					continue
				}

				// 最终保护：只有当实际媒体数量为0时，才删除专辑
				if actualMediaCount <= 0 {
					albumsToDelete = append(albumsToDelete, albumObjID)
				}
			}
		}

		// 批量删除无效专辑
		if len(albumsToDelete) > 0 {
			log.Printf("准备删除 %d 个无效专辑", len(albumsToDelete))
			for _, albumID := range albumsToDelete {
				if err := uc.albumRepo.DeleteByID(ctx, albumID); err != nil {
					if logOutputEnabled {
						log.Printf("删除无效专辑失败 (ID %s): %v", albumID.Hex(), err)
					}
				} else {
					invalidAlbumNums++
				}
			}
			log.Printf("已删除 %d 个无效专辑", invalidAlbumNums)
		} else {
			log.Printf("没有发现需要删除的无效专辑")
		}

		// 4. 执行CUE音频的艺术家统计 - 无论filesToProcess是否为空，都需要执行统计
		// 根据filesToProcess是否为空，选择不同的统计方式
		log.Printf("执行CUE音频艺术家统计，共 %d 个艺术家", len(artistIDs))

		// 处理CUE音频的艺术家统计
		for _, id := range artistIDs {
			artistID := id.Hex()

			// 确保结构体已初始化
			if _, exists := invalidArtistCounts[artistID]; !exists {
				invalidArtistCounts[artistID] = &artistStats{}
			}

			if len(filesToProcess) > 0 {
				// 当指定了要处理的文件集合时，执行详细的CUE音频艺术家统计
				// 处理CUE音频主艺术家统计
				cueMediaCount, err := uc.mediaCueRepo.InspectMediaCueCountByArtist(ctx, artistID, cueAudioPaths)
				if err != nil {
					if logOutputEnabled {
						log.Printf("CUE音频主艺术家统计失败 (ID %s): %v", artistID, err)
					}
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
					if logOutputEnabled {
						log.Printf("CUE音频合作艺术家统计失败 (ID %s): %v", artistID, err)
					}
				} else {
					if guestCueMediaCount >= 0 {
						invalidArtistCounts[artistID].guestCueMediaCount += guestCueMediaCount
					} else {
						invalidArtistCounts[artistID].guestCueMediaCount++
						deleteArtistCounts[artistID]++
					}
				}
			} else {
				// 完整扫描模式下，直接从数据库获取统计信息
				cueMediaCount, err := uc.mediaCueRepo.MediaCueCountByArtist(ctx, artistID)
				if err != nil {
					if logOutputEnabled {
						log.Printf("获取CUE音频主艺术家统计失败 (ID %s): %v", artistID, err)
					}
				} else {
					invalidArtistCounts[artistID].cueMediaCount = int(cueMediaCount)
				}

				guestCueMediaCount, err := uc.mediaCueRepo.GuestMediaCueCountByArtist(ctx, artistID)
				if err != nil {
					if logOutputEnabled {
						log.Printf("获取CUE音频合作艺术家统计失败 (ID %s): %v", artistID, err)
					}
				} else {
					invalidArtistCounts[artistID].guestCueMediaCount = int(guestCueMediaCount)
				}
			}
		}

		// 更新重构阶段进度
		uc.scanMutex.Lock()
		// 初始进入重构阶段时，进度应该是遍历和统计阶段的总和
		uc.scanProgress = uc.scanStageWeights.traversal + uc.scanStageWeights.statistics
		// 更新taskProg.ProcessedFiles以反映进度
		processed := int32(float32(taskProg.TotalFiles) * uc.scanProgress)
		taskProg.Mu.Lock()
		atomic.StoreInt32(&taskProg.ProcessedFiles, processed)
		taskProg.Mu.Unlock()
		uc.scanMutex.Unlock()

		// 5. 清除无效的艺术家（删除统计：【单曲+合作单曲+CD+合作CD+专辑+合作专辑都为0】的艺术家）
		artistIDStrCount := 0
		var artistsToDelete []primitive.ObjectID

		// 在重构阶段的各个关键点更新进度，确保进度平滑过渡
		// 1. 处理艺术家删除前，更新进度到重构阶段的25%
		uc.scanMutex.Lock()
		progress := uc.scanStageWeights.traversal +
			uc.scanStageWeights.statistics +
			uc.scanStageWeights.refactoring*0.25
		uc.scanProgress = progress
		// 更新taskProg.ProcessedFiles以反映进度
		processed = int32(float32(taskProg.TotalFiles) * uc.scanProgress)
		taskProg.Mu.Lock()
		atomic.StoreInt32(&taskProg.ProcessedFiles, processed)
		taskProg.Mu.Unlock()
		uc.scanMutex.Unlock()

		// 先收集需要删除的艺术家ID
		for _, artistID := range artistIDs {
			artistIDStr := artistID.Hex()

			// 获取艺术家的统计信息
			stats, exists := invalidArtistCounts[artistIDStr]
			if !exists {
				// 如果没有统计信息，可能是因为艺术家没有任何关联
				stats = &artistStats{}
			}

			// 检查艺术家是否有任何有效的媒体关联
			// 只有当所有统计项都为0时，才删除艺术家
			if stats.mediaCount == 0 &&
				stats.guestMediaCount == 0 &&
				stats.cueMediaCount == 0 &&
				stats.guestCueMediaCount == 0 &&
				stats.albumCount == 0 &&
				stats.guestAlbumCount == 0 {
				// 额外验证：再次检查艺术家是否确实没有任何关联
				// 检查常规音频
				mediaCount, _ := uc.MediaRepo.MediaCountByArtist(ctx, artistIDStr)
				guestMediaCount, _ := uc.MediaRepo.GuestMediaCountByArtist(ctx, artistIDStr)

				// 检查CUE音频
				cueMediaCount, _ := uc.mediaCueRepo.MediaCueCountByArtist(ctx, artistIDStr)
				guestCueMediaCount, _ := uc.mediaCueRepo.GuestMediaCueCountByArtist(ctx, artistIDStr)

				// 最终确认：所有统计项都为0
				if mediaCount == 0 && guestMediaCount == 0 && cueMediaCount == 0 && guestCueMediaCount == 0 {
					artistsToDelete = append(artistsToDelete, artistID)
					delete(invalidArtistCounts, artistIDStr)
				}
			}
		}

		// 批量删除无效艺术家
		if len(artistsToDelete) > 0 {
			// 添加保护机制：防止大量艺术家被误删
			artistDeletionThreshold := int(float64(len(artistIDs)) * 0.1) // 10%的阈值

			// 如果要删除的艺术家数量超过阈值，记录警告日志
			if len(artistsToDelete) > artistDeletionThreshold {
				log.Printf("警告：准备删除 %d 个艺术家，超过总艺术家数量的10%% (%d/%d)\n",
					len(artistsToDelete), len(artistsToDelete), len(artistIDs))
				log.Printf("当前删除操作将谨慎执行，确保只删除确实无效的艺术家")
			}

			log.Printf("准备删除 %d 个无效艺术家", len(artistsToDelete))

			// 执行删除操作
			for _, artistID := range artistsToDelete {
				// 再次验证：确保艺术家确实没有任何关联
				artistIDStr := artistID.Hex()
				mediaCount, _ := uc.MediaRepo.MediaCountByArtist(ctx, artistIDStr)
				guestMediaCount, _ := uc.MediaRepo.GuestMediaCountByArtist(ctx, artistIDStr)
				cueMediaCount, _ := uc.mediaCueRepo.MediaCueCountByArtist(ctx, artistIDStr)
				guestCueMediaCount, _ := uc.mediaCueRepo.GuestMediaCueCountByArtist(ctx, artistIDStr)

				// 最终确认：只有当所有统计项都为0时，才删除艺术家
				if mediaCount == 0 && guestMediaCount == 0 && cueMediaCount == 0 && guestCueMediaCount == 0 {
					err := uc.artistRepo.DeleteByID(ctx, artistID)
					if err != nil {
						if logOutputEnabled {
							log.Printf("艺术家删除失败 (ID %s): %v", artistID.Hex(), err)
						}
					} else {
						artistIDStrCount++
						log.Printf("已删除无效艺术家 (ID %s)", artistID.Hex())
					}
				} else {
					// 如果艺术家突然有了关联，跳过删除
					log.Printf("跳过删除艺术家 (ID %s)，因为发现了有效关联", artistID.Hex())
				}
			}
			log.Printf("已删除 %d 个无效艺术家", artistIDStrCount)
		}

		// 2. 处理完艺术家删除后，更新进度到重构阶段的50%
		uc.scanMutex.Lock()
		progress = uc.scanStageWeights.traversal +
			uc.scanStageWeights.statistics +
			uc.scanStageWeights.refactoring*0.5
		uc.scanProgress = progress
		// 更新taskProg.ProcessedFiles以反映进度
		processed = int32(float32(taskProg.TotalFiles) * uc.scanProgress)
		taskProg.Mu.Lock()
		atomic.StoreInt32(&taskProg.ProcessedFiles, processed)
		taskProg.Mu.Unlock()
		uc.scanMutex.Unlock()

		// 6. 清除无效的常规音频 - 只有在完整扫描模式下才执行
		if len(filesToProcess) == 0 {
			log.Printf("开始清理无效的常规音频")
			delResult, invalidMediaArtist, err := uc.MediaRepo.DeleteAllInvalid(ctx, regularAudioPaths)
			if err != nil {
				log.Printf("常规音频清理失败: %v", err)
				return fmt.Errorf("regular audio cleanup failed: %w", err)
			} else if delResult > 0 {
				log.Printf("已删除 %d 个无效的常规音频", delResult)
				// 更新艺术家计数
				for _, artist := range invalidMediaArtist {
					artistID := artist.ArtistID.Hex()
					// 确保结构体已初始化
					if _, exists := invalidArtistCounts[artistID]; !exists {
						invalidArtistCounts[artistID] = &artistStats{}
					}
					invalidArtistCounts[artistID].mediaCount = int(artist.Count)
				}
			}

			// 3. 处理完常规音频清理后，更新进度到重构阶段的70%
			uc.scanMutex.Lock()
			progress = uc.scanStageWeights.traversal +
				uc.scanStageWeights.statistics +
				uc.scanStageWeights.refactoring*0.7
			uc.scanProgress = progress
			// 更新taskProg.ProcessedFiles以反映进度
			processed = int32(float32(taskProg.TotalFiles) * uc.scanProgress)
			taskProg.Mu.Lock()
			atomic.StoreInt32(&taskProg.ProcessedFiles, processed)
			taskProg.Mu.Unlock()
			uc.scanMutex.Unlock()

			// 7. 清除无效的CUE音频 - 只有在完整扫描模式下才执行
			log.Printf("开始清理无效的CUE音频")
			delCueResult, invalidMediaCueArtist, err := uc.mediaCueRepo.DeleteAllInvalid(ctx, cueAudioPaths)
			if err != nil {
				log.Printf("CUE音频清理失败: %v", err)
				return fmt.Errorf("CUE audio cleanup failed: %w", err)
			} else if delCueResult > 0 {
				log.Printf("已删除 %d 个无效的CUE音频", delCueResult)
				// 更新艺术家计数
				for _, artist := range invalidMediaCueArtist {
					artistID := artist.ArtistID.Hex()
					// 确保结构体已初始化
					if _, exists := invalidArtistCounts[artistID]; !exists {
						invalidArtistCounts[artistID] = &artistStats{}
					}
					invalidArtistCounts[artistID].cueMediaCount = int(artist.Count)
				}
			}

			// 4. 处理完CUE音频清理后，更新进度到重构阶段的90%
			uc.scanMutex.Lock()
			progress = uc.scanStageWeights.traversal +
				uc.scanStageWeights.statistics +
				uc.scanStageWeights.refactoring*0.9
			uc.scanProgress = progress
			// 更新taskProg.ProcessedFiles以反映进度
			processed = int32(float32(taskProg.TotalFiles) * uc.scanProgress)
			taskProg.Mu.Lock()
			atomic.StoreInt32(&taskProg.ProcessedFiles, processed)
			taskProg.Mu.Unlock()
			uc.scanMutex.Unlock()
		}

		// 8. 重构媒体库艺术家操作数 - 移除了错误的ResetALLField调用
		if len(invalidArtistCounts) > 0 {
			log.Printf("开始重构 %d 个艺术家的操作数", len(invalidArtistCounts))
			// 注意：移除了错误的ResetALLField调用，该方法会重置所有艺术家的字段
			// 该方法不应在循环中调用，否则会清空所有艺术家的计数
		}

		// 清除无效的艺术家：应用艺术家操作数之后
		invalid, err := uc.artistRepo.DeleteAllInvalid(ctx)
		if err != nil {
			return err
		}
		if invalid > 0 {
			log.Printf("已删除 %d 个无效艺术家", invalid)
		}

		// 5. 重构阶段完成，更新进度到100%
		uc.scanMutex.Lock()
		// 完成所有阶段，进度应该是1.0（100%）
		uc.scanProgress = 1.0
		// 更新taskProg.ProcessedFiles以反映进度
		processed = int32(float32(taskProg.TotalFiles) * uc.scanProgress)
		taskProg.Mu.Lock()
		atomic.StoreInt32(&taskProg.ProcessedFiles, processed)
		taskProg.Mu.Unlock()
		uc.scanMutex.Unlock()
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
	allWords, err := uc.MediaRepo.GetHighFrequencyWords(ctx, 100)
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
	defer wg.Done()

	// 上下文取消检查
	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return
	default:
	}

	// 获取工作槽
	var workerSlotAcquired bool
	select {
	case uc.workerPool <- struct{}{}:
		workerSlotAcquired = true
		defer func() { <-uc.workerPool }()
	case <-ctx.Done():
		errChan <- ctx.Err()
		return
	}

	// 只有成功获取工作槽后才处理文件
	if workerSlotAcquired {
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
				log.Printf("音频提取失败: %s | %v", path, err)
				// 不要提前返回，继续执行后续清理
			} else {
				if mediaFile != nil && mediaFile.Title == "" {
					mediaFile.Title = "Unknown Title"
					mediaFile.OrderTitle = "Unknown Title"
					mediaFile.SortTitle = "Unknown Title"
				}
				if mediaFileCue != nil && mediaFileCue.Title == "" {
					mediaFileCue.Title = "Unknown Title"
				}

				if err := uc.processAudioHierarchy(ctx, artists, album, mediaFile, mediaFileCue); err != nil {
					log.Printf("音频层级处理失败: %s | %v", path, err)
					// 不要提前返回，继续执行后续清理
				} else {
					// 处理封面，包括FFmpeg处理
					if err := uc.processAudioCover(
						ctx,
						mediaFile,
						album,
						artists,
						mediaFileCue,
						path,
						coverTempPath,
					); err != nil {
						log.Printf("封面处理失败: %s | %v", path, err)
						// 不要提前返回，继续执行后续清理
					}

					// 处理歌词
					if err = uc.processAudioLyrics(
						ctx,
						mediaFile,
						mediaFileCue,
						path,
						lyricsTempPathWithTimestamp,
					); err != nil {
						log.Printf("歌词处理失败: %s | %v", path, err)
						// 不要提前返回，继续执行后续清理
					}
				}
			}

			// 所有处理完成后，更新任务级别的处理计数器
			// 确保FFmpeg处理完成后才更新计数器
			atomic.AddInt32(&taskProg.ProcessedFiles, 1)
		}
	}
}

// 批量创建元数据基本信息，减少数据库查询次数
func (uc *AudioProcessingUsecase) batchCreateMetadataBasicInfo(
	ctx context.Context,
	paths []string,
	libraryFolderID primitive.ObjectID,
) (map[string]*domain_file_entity.FileMetadata, error) {
	result := make(map[string]*domain_file_entity.FileMetadata)

	// 1. 批量查询已存在的文件元数据
	existingFilesMap := make(map[string]*domain_file_entity.FileMetadata)
	for _, path := range paths {
		existingFile, err := uc.fileRepo.FindByPath(ctx, path)
		if err != nil && !errors.Is(err, driver.ErrNoDocuments) {
			log.Printf("路径查询失败: %s | %v", path, err)
			continue
		}
		if existingFile != nil {
			existingFilesMap[path] = existingFile
			result[path] = existingFile
		}
	}

	// 2. 处理不存在的文件
	var filesToCreate []string
	for _, path := range paths {
		if _, exists := existingFilesMap[path]; !exists {
			filesToCreate = append(filesToCreate, path)
		}
	}

	if len(filesToCreate) == 0 {
		return result, nil
	}

	log.Printf("批量处理 %d 个新文件", len(filesToCreate))

	// 3. 并行处理文件创建
	var wg sync.WaitGroup
	var mu sync.Mutex
	var finalErr error

	for _, path := range filesToCreate {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			// 打开文件
			file, err := os.Open(p)
			if err != nil {
				log.Printf("文件打开失败: %s | %v", p, err)
				mu.Lock()
				if finalErr == nil {
					finalErr = fmt.Errorf("文件打开失败 %s: %w", p, err)
				}
				mu.Unlock()
				return
			}
			defer file.Close()

			// 获取文件信息
			stat, err := file.Stat()
			if err != nil {
				log.Printf("文件信息获取失败: %s | %v", p, err)
				mu.Lock()
				if finalErr == nil {
					finalErr = fmt.Errorf("文件信息获取失败 %s: %w", p, err)
				}
				mu.Unlock()
				return
			}

			// 检测文件类型
			fileType, err := uc.detector.DetectMediaType(p)
			if err != nil {
				log.Printf("文件类型检测失败: %s | %v", p, err)
				mu.Lock()
				if finalErr == nil {
					finalErr = fmt.Errorf("文件类型检测失败 %s: %w", p, err)
				}
				mu.Unlock()
				return
			}

			// 计算哈希值
			hash := sha256.New()
			if _, err := io.CopyBuffer(hash, file, make([]byte, 32*1024)); err != nil {
				log.Printf("哈希计算失败: %s | %v", p, err)
				mu.Lock()
				if finalErr == nil {
					finalErr = fmt.Errorf("哈希计算失败 %s: %w", p, err)
				}
				mu.Unlock()
				return
			}

			// 创建元数据
			normalizedPath := filepath.ToSlash(filepath.Clean(p))
			metadata := &domain_file_entity.FileMetadata{
				ID:        primitive.NewObjectID(),
				FolderID:  libraryFolderID,
				FilePath:  normalizedPath,
				FileType:  fileType,
				Size:      stat.Size(),
				ModTime:   stat.ModTime(),
				Checksum:  fmt.Sprintf("%x", hash.Sum(nil)),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// 保存到数据库
			if err := uc.fileRepo.Upsert(ctx, metadata); err != nil {
				log.Printf("文件写入失败: %s | %v", p, err)
				mu.Lock()
				if finalErr == nil {
					finalErr = fmt.Errorf("数据库写入失败 %s: %w", p, err)
				}
				mu.Unlock()
				return
			}

			// 添加到结果
			mu.Lock()
			result[p] = metadata
			mu.Unlock()
		}(path)
	}

	wg.Wait()

	return result, finalErr
}

// 创建单个元数据基本信息，调用批量处理函数
func (uc *AudioProcessingUsecase) createMetadataBasicInfo(
	path string,
	libraryFolderID primitive.ObjectID,
) (*domain_file_entity.FileMetadata, error) {
	result, err := uc.batchCreateMetadataBasicInfo(context.Background(), []string{path}, libraryFolderID)
	if err != nil {
		return nil, err
	}
	return result[path], nil
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
		if _, err := uc.MediaRepo.UpdateByID(ctx, media.ID, mediaUpdate); err != nil {
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
	if uc.MediaRepo == nil || uc.artistRepo == nil || uc.albumRepo == nil || uc.mediaCueRepo == nil {
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

	if artists != nil && len(artists) > 0 {
		for index, artist := range artists {
			if artist != nil && !artist.ID.IsZero() && index == 0 {
				if mediaFileCue != nil {
					log.Printf("艺术家(%s) CUE文件统计 - 文件大小: %d, 播放时长: %d", artist.ID.Hex(), mediaFileCue.Size, mediaFileCue.CueDuration)
				}
				if mediaFile != nil {
					log.Printf("艺术家(%s) 媒体文件统计 - 文件大小: %d, 播放时长: %d", artist.ID.Hex(), mediaFile.Size, mediaFile.Duration)
				}
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
	if uc.MediaRepo == nil {
		log.Print("专辑仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	existing, err := uc.MediaRepo.GetByID(ctx, media.ID)
	if err != nil {
		log.Printf("组合查询错误: %v", err)
	}

	if existing != nil {
		if media.UpdatedAt != existing.UpdatedAt {
			if _, err := uc.MediaRepo.Upsert(ctx, media); err != nil {
				log.Printf("单曲创建失败: %s | %v", media.Path, err)
				return err
			}
		}
	} else {
		if _, err := uc.MediaRepo.Upsert(ctx, media); err != nil {
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

// GetFilesWithMissingMetadata 获取缺失元数据的文件
func (uc *AudioProcessingUsecase) GetFilesWithMissingMetadata(ctx context.Context, folderPath string, folderType int) ([]*scene_audio_db_models.MediaFileMetadata, error) {
	return uc.MediaRepo.GetFilesWithMissingMetadata(ctx, folderPath, folderType)
}

// ClearAllData 清空所有媒体数据
func (uc *AudioProcessingUsecase) ClearAllData(ctx context.Context) error {
	// 1. 删除所有媒体文件
	if _, err := uc.MediaRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("删除媒体文件失败: %w", err)
	}

	// 2. 删除所有CUE媒体文件
	if _, err := uc.mediaCueRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("删除CUE媒体文件失败: %w", err)
	}

	// 3. 删除所有艺术家
	if _, err := uc.artistRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("删除艺术家失败: %w", err)
	}

	// 4. 删除所有专辑
	if _, err := uc.albumRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("删除专辑失败: %w", err)
	}

	// 5. 删除所有词云数据
	if err := uc.wordCloudRepo.AllDelete(ctx); err != nil {
		return fmt.Errorf("删除词云数据失败: %w", err)
	}

	// 6. 删除所有歌词文件
	if _, err := uc.lyricsFileRepo.CleanAll(ctx); err != nil {
		return fmt.Errorf("删除歌词文件失败: %w", err)
	}

	log.Printf("已清空所有媒体数据")
	return nil
}

func (uc *AudioProcessingUsecase) shouldProcess(path string, folderType int) bool {
	return uc.commonUtil.ShouldProcess(path, folderType)
}

func (uc *AudioProcessingUsecase) fastCountFilesInFolder(rootPath string, folderType int) (int, error) {
	return uc.commonUtil.FastCountFilesInFolder(rootPath, folderType)
}
