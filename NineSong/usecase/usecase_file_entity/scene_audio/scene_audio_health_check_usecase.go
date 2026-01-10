package scene_audio

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LibraryHealthCheckUsecase struct {
	artistRepo   scene_audio_db_interface.ArtistRepository
	albumRepo    scene_audio_db_interface.AlbumRepository
	mediaRepo    scene_audio_db_interface.MediaFileRepository
	mediaCueRepo scene_audio_db_interface.MediaFileCueRepository
}

func NewLibraryHealthCheckUsecase(
	artistRepo scene_audio_db_interface.ArtistRepository,
	albumRepo scene_audio_db_interface.AlbumRepository,
	mediaRepo scene_audio_db_interface.MediaFileRepository,
	mediaCueRepo scene_audio_db_interface.MediaFileCueRepository,
) *LibraryHealthCheckUsecase {
	return &LibraryHealthCheckUsecase{
		artistRepo:   artistRepo,
		albumRepo:    albumRepo,
		mediaRepo:    mediaRepo,
		mediaCueRepo: mediaCueRepo,
	}
}

func (uc *LibraryHealthCheckUsecase) PerformHealthCheck(
	ctx context.Context,
	libraryFolders []*domain_file_entity.LibraryFolderMetadata,
	taskProg *domain_util.TaskProgress,
) error {
	log.Printf("LibraryHealthCheck: 开始媒体库健康检查")

	if taskProg == nil {
		return fmt.Errorf("taskProg 不能为空")
	}

	taskProg.Mu.Lock()
	taskProg.Status = "processing"
	taskProg.Stage = "health_check"
	taskProg.Initialized = true
	taskProg.TotalFiles = 100
	taskProg.ProcessedFiles = 0
	taskProg.Mu.Unlock()

	libraryPaths := make([]string, 0, len(libraryFolders))
	for _, folder := range libraryFolders {
		if folder != nil && folder.FolderPath != "" {
			libraryPaths = append(libraryPaths, folder.FolderPath)
		}
	}

	log.Printf("LibraryHealthCheck: 开始清理无效媒体文件")
	deletedMedia, err := uc.CleanupInvalidMediaFiles(ctx, libraryPaths)
	if err != nil {
		log.Printf("LibraryHealthCheck: 清理无效媒体文件失败: %v", err)
	} else {
		log.Printf("LibraryHealthCheck: 已清理 %d 个无效媒体文件", deletedMedia)
	}

	atomic.StoreInt32(&taskProg.ProcessedFiles, 15)

	log.Printf("LibraryHealthCheck: 开始清理无效CUE文件")
	deletedCue, err := uc.CleanupInvalidCueFiles(ctx, libraryPaths)
	if err != nil {
		log.Printf("LibraryHealthCheck: 清理无效CUE文件失败: %v", err)
	} else {
		log.Printf("LibraryHealthCheck: 已清理 %d 个无效CUE文件", deletedCue)
	}

	atomic.StoreInt32(&taskProg.ProcessedFiles, 20)

	log.Printf("LibraryHealthCheck: 开始清理无效实体")

	deletedArtists, err := uc.CleanupInvalidArtists(ctx)
	if err != nil {
		log.Printf("LibraryHealthCheck: 清理无效艺术家失败: %v", err)
	} else {
		log.Printf("LibraryHealthCheck: 已清理 %d 个无效艺术家", deletedArtists)
	}

	atomic.StoreInt32(&taskProg.ProcessedFiles, 45)

	deletedAlbums, err := uc.CleanupInvalidAlbums(ctx)
	if err != nil {
		log.Printf("LibraryHealthCheck: 清理无效专辑失败: %v", err)
	} else {
		log.Printf("LibraryHealthCheck: 已清理 %d 个无效专辑", deletedAlbums)
	}

	atomic.StoreInt32(&taskProg.ProcessedFiles, 60)

	log.Printf("LibraryHealthCheck: 开始重建艺术家-专辑关联关系")

	if err := uc.RefactorArtistAlbums(ctx); err != nil {
		log.Printf("LibraryHealthCheck: 重构艺术家-专辑映射失败: %v", err)
	} else {
		log.Printf("LibraryHealthCheck: 艺术家-专辑映射重构完成")
	}

	atomic.StoreInt32(&taskProg.ProcessedFiles, 80)

	log.Printf("LibraryHealthCheck: 开始统计更新")

	if err := uc.UpdateStatistics(ctx, taskProg); err != nil {
		log.Printf("LibraryHealthCheck: 更新统计失败: %v", err)
	} else {
		log.Printf("LibraryHealthCheck: 统计更新完成")
	}

	atomic.StoreInt32(&taskProg.ProcessedFiles, 100)
	taskProg.Mu.Lock()
	taskProg.Status = "completed"
	taskProg.Stage = "completed"
	taskProg.Mu.Unlock()

	log.Printf("LibraryHealthCheck: 媒体库健康检查完成")
	return nil
}

func (uc *LibraryHealthCheckUsecase) CleanupInvalidMediaFiles(ctx context.Context, libraryPaths []string) (int, error) {
	log.Printf("开始清理无效媒体文件（数据库中有记录但文件不存在）")

	allMedia, err := uc.mediaRepo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取媒体文件列表失败: %w", err)
	}

	var deletedCount int
	for _, media := range allMedia {
		if media == nil || media.Path == "" {
			continue
		}

		if !isFileExists(media.Path) {
			if err := uc.mediaRepo.DeleteByID(ctx, media.ID); err != nil {
				log.Printf("删除无效媒体文件失败 (ID: %s, Path: %s): %v", media.ID.Hex(), media.Path, err)
			} else {
				deletedCount++
				log.Printf("已删除无效媒体文件: %s", media.Path)
			}
		}
	}

	log.Printf("无效媒体文件清理完成，共删除 %d 个", deletedCount)
	return deletedCount, nil
}

func (uc *LibraryHealthCheckUsecase) CleanupInvalidCueFiles(ctx context.Context, libraryPaths []string) (int, error) {
	log.Printf("开始清理无效CUE文件（数据库中有记录但文件不存在）")

	allCue, err := uc.mediaCueRepo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取CUE文件列表失败: %w", err)
	}

	var deletedCount int
	for _, cue := range allCue {
		if cue == nil || cue.Path == "" {
			continue
		}

		if !isFileExists(cue.Path) {
			if err := uc.mediaCueRepo.DeleteByID(ctx, cue.ID); err != nil {
				log.Printf("删除无效CUE文件失败 (ID: %s, Path: %s): %v", cue.ID.Hex(), cue.Path, err)
			} else {
				deletedCount++
				log.Printf("已删除无效CUE文件: %s", cue.Path)
			}
		}
	}

	log.Printf("无效CUE文件清理完成，共删除 %d 个", deletedCount)
	return deletedCount, nil
}

func isFileExists(path string) bool {
	normalizedPath := strings.ReplaceAll(path, "/", "\\")
	if _, err := os.Stat(normalizedPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (uc *LibraryHealthCheckUsecase) CleanupInvalidArtists(ctx context.Context) (int, error) {
	log.Printf("开始清理无效艺术家（没有媒体文件的艺术家）")

	artistIDs, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取艺术家ID列表失败: %w", err)
	}

	var deletedCount int
	for _, artistID := range artistIDs {
		artistIDStr := artistID.Hex()

		mediaCount, _ := uc.mediaRepo.MediaCountByArtist(ctx, artistIDStr)
		guestMediaCount, _ := uc.mediaRepo.GuestMediaCountByArtist(ctx, artistIDStr)

		if mediaCount == 0 && guestMediaCount == 0 {
			if err := uc.artistRepo.DeleteByID(ctx, artistID); err != nil {
				log.Printf("删除无效艺术家失败 (ID: %s): %v", artistIDStr, err)
			} else {
				deletedCount++
				log.Printf("已删除无效艺术家: %s", artistIDStr)
			}
		}
	}

	log.Printf("无效艺术家清理完成，共删除 %d 个", deletedCount)
	return deletedCount, nil
}

func (uc *LibraryHealthCheckUsecase) CleanupInvalidAlbums(ctx context.Context) (int, error) {
	log.Printf("开始清理无效专辑（没有媒体文件的专辑）")

	albumIDs, err := uc.albumRepo.GetAllIDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取专辑ID列表失败: %w", err)
	}

	var deletedCount int
	for _, albumID := range albumIDs {
		albumIDStr := albumID.Hex()

		songCount, _ := uc.mediaRepo.MediaCountByAlbum(ctx, albumIDStr)

		if songCount == 0 {
			if err := uc.albumRepo.DeleteByID(ctx, albumID); err != nil {
				log.Printf("删除无效专辑失败 (ID: %s): %v", albumIDStr, err)
			} else {
				deletedCount++
				log.Printf("已删除无效专辑: %s", albumIDStr)
			}
		}
	}

	log.Printf("无效专辑清理完成，共删除 %d 个", deletedCount)
	return deletedCount, nil
}

func (uc *LibraryHealthCheckUsecase) RefactorArtistAlbums(ctx context.Context) error {
	log.Printf("开始重构艺术家-专辑关联关系")

	artistIDs, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取艺术家ID列表失败: %w", err)
	}

	albumIDs, err := uc.albumRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取专辑ID列表失败: %w", err)
	}

	albumArtistMap := make(map[primitive.ObjectID]string)
	for _, albumID := range albumIDs {
		album, err := uc.albumRepo.GetByID(ctx, albumID)
		if err != nil {
			continue
		}
		if album != nil && album.ArtistID != "" {
			albumArtistMap[albumID] = album.ArtistID
		}
	}

	var wg sync.WaitGroup

	for _, artistID := range artistIDs {
		wg.Add(1)
		go func(id primitive.ObjectID) {
			defer wg.Done()

			var albumIDs []primitive.ObjectID
			for albumID, artistIDStr := range albumArtistMap {
				if artistIDStr == id.Hex() {
					albumIDs = append(albumIDs, albumID)
				}
			}

			if len(albumIDs) > 0 {
				update := bson.M{"$set": bson.M{"album_ids": albumIDs}}
				if _, err := uc.artistRepo.UpdateByID(ctx, id, update); err != nil {
					log.Printf("艺术家 %s 更新专辑列表失败: %v", id.Hex(), err)
				} else {
					log.Printf("艺术家 %s 专辑列表已更新 (共 %d 个专辑)", id.Hex(), len(albumIDs))
				}
			}
		}(artistID)
	}

	wg.Wait()
	log.Printf("艺术家-专辑关联关系重构完成")
	return nil
}

func (uc *LibraryHealthCheckUsecase) UpdateStatistics(ctx context.Context, taskProg *domain_util.TaskProgress) error {
	log.Printf("开始更新统计（增量模式）")

	artistIDs, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取艺术家ID列表失败: %w", err)
	}

	totalArtists := len(artistIDs)
	var processedArtists int32

	maxConcurrency := 50
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, artistID := range artistIDs {
		wg.Add(1)
		go func(id primitive.ObjectID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			artistIDStr := id.Hex()

			albumCount, _ := uc.albumRepo.AlbumCountByArtist(ctx, artistIDStr)
			guestAlbumCount, _ := uc.albumRepo.GuestAlbumCountByArtist(ctx, artistIDStr)
			songCount, _ := uc.mediaRepo.MediaCountByArtist(ctx, artistIDStr)
			guestSongCount, _ := uc.mediaRepo.GuestMediaCountByArtist(ctx, artistIDStr)

			uc.artistRepo.SetCounter(ctx, id, "album_count", int(albumCount))
			uc.artistRepo.SetCounter(ctx, id, "guest_album_count", int(guestAlbumCount))
			uc.artistRepo.SetCounter(ctx, id, "song_count", int(songCount))
			uc.artistRepo.SetCounter(ctx, id, "guest_song_count", int(guestSongCount))

			processedArtists++
			if taskProg != nil && totalArtists > 0 {
				progress := 0.8 + 0.2*float32(processedArtists)/float32(totalArtists)
				atomic.StoreInt32(&taskProg.ProcessedFiles, int32(float32(taskProg.TotalFiles)*progress))
			}
		}(artistID)
	}

	wg.Wait()

	albumIDs, err := uc.albumRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取专辑ID列表失败: %w", err)
	}

	totalAlbums := len(albumIDs)
	var processedAlbums int32

	for _, albumID := range albumIDs {
		wg.Add(1)
		go func(id primitive.ObjectID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			albumIDStr := id.Hex()
			songCount, _ := uc.mediaRepo.MediaCountByAlbum(ctx, albumIDStr)

			update := bson.M{"$set": bson.M{"song_count": int(songCount)}}
			if _, err := uc.albumRepo.UpdateByID(ctx, id, update); err != nil {
				log.Printf("专辑 %s 更新 song_count 失败: %v", albumIDStr, err)
			}

			processedAlbums++
			if taskProg != nil && totalAlbums > 0 {
				progress := 0.8 + 0.2*float32(processedAlbums)/float32(totalAlbums)
				atomic.StoreInt32(&taskProg.ProcessedFiles, int32(float32(taskProg.TotalFiles)*progress))
			}
		}(albumID)
	}

	wg.Wait()

	log.Printf("统计更新完成，共处理 %d 个艺术家, %d 个专辑", len(artistIDs), len(albumIDs))
	return nil
}
