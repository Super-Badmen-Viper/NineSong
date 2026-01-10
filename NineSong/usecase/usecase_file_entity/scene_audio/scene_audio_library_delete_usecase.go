package scene_audio

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
)

type LibraryDeleteResult struct {
	DeletedMediaCount     int
	DeletedMediaCueCount  int
	DeletedArtistCount    int
	DeletedAlbumCount     int
	DeletedLyricsCount    int
	DeletedWordCloudCount int
	SkippedMediaCount     int
	SkippedCueCount       int
	Errors                []string
}

type LibraryDeleteUsecase struct {
	mediaRepo      scene_audio_db_interface.MediaFileRepository
	mediaCueRepo   scene_audio_db_interface.MediaFileCueRepository
	artistRepo     scene_audio_db_interface.ArtistRepository
	albumRepo      scene_audio_db_interface.AlbumRepository
	lyricsFileRepo scene_audio_db_interface.LyricsFileRepository
	wordCloudRepo  scene_audio_db_interface.WordCloudDBRepository
	statistics     *AudioStatisticsUsecase
}

func NewLibraryDeleteUsecase(
	mediaRepo scene_audio_db_interface.MediaFileRepository,
	mediaCueRepo scene_audio_db_interface.MediaFileCueRepository,
	artistRepo scene_audio_db_interface.ArtistRepository,
	albumRepo scene_audio_db_interface.AlbumRepository,
	lyricsFileRepo scene_audio_db_interface.LyricsFileRepository,
	wordCloudRepo scene_audio_db_interface.WordCloudDBRepository,
) *LibraryDeleteUsecase {
	statistics := NewAudioStatisticsUsecase(
		artistRepo,
		albumRepo,
		mediaRepo,
		mediaCueRepo,
	)

	return &LibraryDeleteUsecase{
		mediaRepo:      mediaRepo,
		mediaCueRepo:   mediaCueRepo,
		artistRepo:     artistRepo,
		albumRepo:      albumRepo,
		lyricsFileRepo: lyricsFileRepo,
		wordCloudRepo:  wordCloudRepo,
		statistics:     statistics,
	}
}

func (uc *LibraryDeleteUsecase) DeleteByLibraryPaths(ctx context.Context, libraryPaths []string) (*LibraryDeleteResult, error) {
	result := &LibraryDeleteResult{}

	log.Printf("开始清理被删除媒体库的元数据，涉及 %d 个路径", len(libraryPaths))

	regularPaths, cuePaths, err := uc.collectAllAudioPaths(libraryPaths)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("收集音频路径失败: %v", err))
		return result, err
	}

	log.Printf("收集到 %d 个常规音频路径, %d 个CUE音频路径", len(regularPaths), len(cuePaths))

	if err := uc.cleanupMediaFiles(ctx, regularPaths, cuePaths, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("清理媒体文件失败: %v", err))
	}

	if err := uc.cleanupInvalidEntities(ctx, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("清理无效实体失败: %v", err))
	}

	log.Printf("开始统计阶段 - 增量更新艺术家和专辑计数器")

	if err := uc.statistics.UpdateStatisticsIncremental(ctx, nil); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("更新统计失败: %v", err))
	}

	log.Printf("媒体库清理完成: 删除了 %d 个媒体文件, %d 个CUE文件, %d 个艺术家, %d 个专辑",
		result.DeletedMediaCount, result.DeletedMediaCueCount, result.DeletedArtistCount, result.DeletedAlbumCount)

	return result, nil
}

func (uc *LibraryDeleteUsecase) collectAllAudioPaths(libraryPaths []string) (regularPaths []string, cuePaths []string, err error) {
	regularPaths = make([]string, 0)
	cuePaths = make([]string, 0)
	addedRegularPaths := make(map[string]bool)
	addedCuePaths := make(map[string]bool)
	excludeWavs := make(map[string]struct{})
	cueResourcesMap := make(map[string]struct{})

	for _, libPath := range libraryPaths {
		err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			dir := filepath.Dir(path)
			baseName := strings.TrimSuffix(filepath.Base(path), ext)

			if ext == ".cue" {
				audioExts := []string{".wav", ".ape", ".flac"}
				for _, audioExt := range audioExts {
					audioPath := filepath.Join(dir, baseName+audioExt)
					if _, err := os.Stat(audioPath); err == nil {
						excludeWavs[audioPath] = struct{}{}
						break
					}
				}
				cueResourcesMap[path] = struct{}{}
			}

			return nil
		})
		if err != nil {
			log.Printf("收集路径失败 %s: %v", libPath, err)
		}
	}

	for _, libPath := range libraryPaths {
		err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			if _, excluded := excludeWavs[path]; excluded {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			normalizedPath := filepath.ToSlash(filepath.Clean(path))

			if ext == ".cue" {
				if _, exists := cueResourcesMap[path]; exists {
					if _, added := addedCuePaths[normalizedPath]; !added {
						cuePaths = append(cuePaths, normalizedPath)
						addedCuePaths[normalizedPath] = true
					}
				}
			} else if isAudioExt(ext) {
				if _, added := addedRegularPaths[normalizedPath]; !added {
					regularPaths = append(regularPaths, normalizedPath)
					addedRegularPaths[normalizedPath] = true
				}
			}

			return nil
		})
		if err != nil {
			log.Printf("收集路径失败 %s: %v", libPath, err)
		}
	}

	return regularPaths, cuePaths, nil
}

func isAudioExt(ext string) bool {
	audioExts := map[string]bool{".mp3": true, ".flac": true, ".wav": true, ".ape": true, ".m4a": true, ".ogg": true, ".aac": true, ".wma": true, ".opus": true, ".alac": true}
	return audioExts[ext]
}

func (uc *LibraryDeleteUsecase) cleanupMediaFiles(ctx context.Context, regularPaths, cuePaths []string, result *LibraryDeleteResult) error {
	if len(regularPaths) > 0 {
		delResult, _, err := uc.mediaRepo.DeleteAllInvalid(ctx, regularPaths)
		if err != nil {
			log.Printf("删除无效媒体文件失败: %v", err)
			result.Errors = append(result.Errors, fmt.Sprintf("删除媒体文件失败: %v", err))
		} else {
			result.DeletedMediaCount = int(delResult)
			log.Printf("已删除 %d 个无效媒体文件", delResult)
		}
	}

	if len(cuePaths) > 0 {
		delResult, _, err := uc.mediaCueRepo.DeleteAllInvalid(ctx, cuePaths)
		if err != nil {
			log.Printf("删除无效CUE媒体文件失败: %v", err)
			result.Errors = append(result.Errors, fmt.Sprintf("删除CUE媒体文件失败: %v", err))
		} else {
			result.DeletedMediaCueCount = int(delResult)
			log.Printf("已删除 %d 个无效CUE媒体文件", delResult)
		}
	}

	return nil
}

func (uc *LibraryDeleteUsecase) cleanupInvalidEntities(ctx context.Context, result *LibraryDeleteResult) error {
	log.Printf("开始清理无效实体 - 艺术家和专辑")

	deletedArtists, err := uc.statistics.CleanupInvalidArtists(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("清理无效艺术家失败: %v", err))
	} else {
		result.DeletedArtistCount = deletedArtists
		log.Printf("已删除 %d 个无效艺术家", deletedArtists)
	}

	deletedAlbums, err := uc.statistics.CleanupInvalidAlbums(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("清理无效专辑失败: %v", err))
	} else {
		result.DeletedAlbumCount = deletedAlbums
		log.Printf("已删除 %d 个无效专辑", deletedAlbums)
	}

	return nil
}

func (uc *LibraryDeleteUsecase) ClearAllData(ctx context.Context) error {
	log.Printf("开始清空所有媒体数据")

	if _, err := uc.mediaRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("删除媒体文件失败: %w", err)
	}

	if _, err := uc.mediaCueRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("删除CUE媒体文件失败: %w", err)
	}

	if _, err := uc.artistRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("删除艺术家失败: %w", err)
	}

	if _, err := uc.albumRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("删除专辑失败: %w", err)
	}

	if err := uc.wordCloudRepo.AllDelete(ctx); err != nil {
		return fmt.Errorf("删除词云数据失败: %w", err)
	}

	if _, err := uc.lyricsFileRepo.CleanAll(ctx); err != nil {
		return fmt.Errorf("删除歌词文件失败: %w", err)
	}

	log.Printf("已清空所有媒体数据")
	return nil
}

func (uc *LibraryDeleteUsecase) ProcessWordCloud(ctx context.Context) error {
	if err := uc.wordCloudRepo.DropAllIndex(ctx); err != nil {
		log.Printf("词云索引删除失败: %v", err)
	}

	if err := uc.wordCloudRepo.AllDelete(ctx); err != nil {
		return fmt.Errorf("词云数据清空失败: %w", err)
	}

	return uc.wordCloudRepo.CreateIndex(ctx, "name", true)
}

func (uc *LibraryDeleteUsecase) CleanupOrphanedMedia(ctx context.Context, libraryPaths []string, taskProg *domain_util.TaskProgress) (int, error) {
	log.Printf("开始清理孤儿媒体文件，剩余媒体库: %d 个", len(libraryPaths))

	if taskProg != nil {
		taskProg.Mu.Lock()
		taskProg.Stage = "cleanup_orphaned"
		taskProg.ProcessedFiles = int32(float32(taskProg.TotalFiles) * 0.8)
		taskProg.Mu.Unlock()
	}

	var deletedCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 获取所有媒体文件
	allMedia, err := uc.mediaRepo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取所有媒体文件失败: %w", err)
	}

	// 获取所有CUE媒体文件
	allCueMedia, err := uc.mediaCueRepo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取所有CUE媒体文件失败: %w", err)
	}

	log.Printf("检查 %d 个常规媒体和 %d 个CUE媒体文件", len(allMedia), len(allCueMedia))

	// 构建库路径集合（用于快速查找）
	libraryPathSet := make(map[string]bool)
	for _, path := range libraryPaths {
		normalizedPath := strings.Replace(path, "/", "\\", -1)
		if !strings.HasSuffix(normalizedPath, "\\") {
			normalizedPath += "\\"
		}
		libraryPathSet[normalizedPath] = true
	}

	// 检查并删除孤儿的常规媒体文件
	for _, media := range allMedia {
		wg.Add(1)
		go func(mediaItem *scene_audio_db_models.MediaFileMetadata) {
			defer wg.Done()

			if mediaItem == nil || mediaItem.Path == "" {
				return
			}

			mediaPath := strings.Replace(mediaItem.Path, "/", "\\", -1)
			isOrphan := true

			for libPath := range libraryPathSet {
				if strings.HasPrefix(mediaPath, libPath) {
					isOrphan = false
					break
				}
			}

			if isOrphan {
				log.Printf("发现孤儿媒体文件: %s", mediaPath)
				if err := uc.mediaRepo.DeleteByID(ctx, mediaItem.ID); err != nil {
					log.Printf("删除孤儿媒体文件失败: %v", err)
				} else {
					mu.Lock()
					deletedCount++
					mu.Unlock()
					log.Printf("已删除孤儿媒体文件: %s", mediaPath)
				}
			}
		}(media)
	}

	// 检查并删除孤儿的CUE媒体文件
	for _, cueMedia := range allCueMedia {
		wg.Add(1)
		go func(cueItem *scene_audio_db_models.MediaFileCueMetadata) {
			defer wg.Done()

			if cueItem == nil || cueItem.Path == "" {
				return
			}

			cuePath := strings.Replace(cueItem.Path, "/", "\\", -1)
			isOrphan := true

			for libPath := range libraryPathSet {
				if strings.HasPrefix(cuePath, libPath) {
					isOrphan = false
					break
				}
			}

			if isOrphan {
				log.Printf("发现孤儿CUE媒体文件: %s", cuePath)
				if err := uc.mediaCueRepo.DeleteByID(ctx, cueItem.ID); err != nil {
					log.Printf("删除孤儿CUE媒体文件失败: %v", err)
				} else {
					mu.Lock()
					deletedCount++
					mu.Unlock()
					log.Printf("已删除孤儿CUE媒体文件: %s", cuePath)
				}
			}
		}(cueMedia)
	}

	wg.Wait()

	// 统计阶段：增量更新艺术家和专辑的计数器
	log.Printf("统计阶段：增量更新艺术家和专辑的计数器")

	if err := uc.statistics.UpdateStatisticsIncremental(ctx, taskProg); err != nil {
		log.Printf("增量更新统计失败: %v", err)
	} else {
		log.Printf("增量统计更新完成")
	}

	// 重构阶段：更新艺术家-专辑映射关系
	log.Printf("重构阶段：更新艺术家-专辑映射关系")
	if err := uc.statistics.RefactorArtistAlbums(ctx); err != nil {
		log.Printf("重构艺术家-专辑映射失败: %v", err)
	} else {
		log.Printf("艺术家-专辑映射重构完成")
	}

	// 清理无效的艺术家和专辑
	log.Printf("清理无效的艺术家和专辑")
	if deletedArtists, err := uc.statistics.CleanupInvalidArtists(ctx); err != nil {
		log.Printf("清理无效艺术家失败: %v", err)
	} else {
		log.Printf("已清理 %d 个无效艺术家", deletedArtists)
	}

	if deletedAlbums, err := uc.statistics.CleanupInvalidAlbums(ctx); err != nil {
		log.Printf("清理无效专辑失败: %v", err)
	} else {
		log.Printf("已清理 %d 个无效专辑", deletedAlbums)
	}

	log.Printf("清理孤儿媒体文件完成，共删除 %d 个", deletedCount)

	if taskProg != nil {
		taskProg.Mu.Lock()
		taskProg.Stage = "completed"
		taskProg.ProcessedFiles = taskProg.TotalFiles
		taskProg.Status = "completed"
		taskProg.Mu.Unlock()
	}

	return deletedCount, nil
}
