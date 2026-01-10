package scene_audio

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AudioStatisticsUsecase struct {
	artistRepo   scene_audio_db_interface.ArtistRepository
	albumRepo    scene_audio_db_interface.AlbumRepository
	mediaRepo    scene_audio_db_interface.MediaFileRepository
	mediaCueRepo scene_audio_db_interface.MediaFileCueRepository
}

func NewAudioStatisticsUsecase(
	artistRepo scene_audio_db_interface.ArtistRepository,
	albumRepo scene_audio_db_interface.AlbumRepository,
	mediaRepo scene_audio_db_interface.MediaFileRepository,
	mediaCueRepo scene_audio_db_interface.MediaFileCueRepository,
) *AudioStatisticsUsecase {
	return &AudioStatisticsUsecase{
		artistRepo:   artistRepo,
		albumRepo:    albumRepo,
		mediaRepo:    mediaRepo,
		mediaCueRepo: mediaCueRepo,
	}
}

func (uc *AudioStatisticsUsecase) IncrementArtistCounter(ctx context.Context, artistID string, counterName string) error {
	_, err := uc.artistRepo.UpdateCounter(ctx, primitive.NilObjectID, counterName, 1)
	if err != nil {
		return err
	}
	return nil
}

func (uc *AudioStatisticsUsecase) DecrementArtistCounter(ctx context.Context, artistID string, counterName string) error {
	_, err := uc.artistRepo.UpdateCounter(ctx, primitive.NilObjectID, counterName, -1)
	if err != nil {
		return err
	}
	return nil
}

func (uc *AudioStatisticsUsecase) SetArtistCounter(ctx context.Context, artistID primitive.ObjectID, counterName string, value int) error {
	return uc.artistRepo.SetCounter(ctx, artistID, counterName, value)
}

func (uc *AudioStatisticsUsecase) UpdateArtistStatisticsByTraversal(ctx context.Context, taskProg *domain_util.TaskProgress, baseProgress float32) error {
	log.Printf("开始通过遍历方式更新艺术家统计")

	artistIDs, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取艺术家ID列表失败: %w", err)
	}

	if len(artistIDs) == 0 {
		log.Printf("没有艺术家需要统计")
		return nil
	}

	totalArtists := len(artistIDs)
	var processedArtists int32
	var mu sync.Mutex

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

			artistCounters := []struct {
				counterName string
			}{
				{"album_count"},
				{"guest_album_count"},
				{"song_count"},
				{"guest_song_count"},
				{"cue_count"},
				{"guest_cue_count"},
			}

			for _, counter := range artistCounters {
				count, err := uc.getArtistCounterByTraversal(ctx, artistIDStr, counter.counterName)
				if err != nil {
					log.Printf("艺术家(%s) %s 统计失败: %v", artistIDStr, counter.counterName, err)
					continue
				}
				if err := uc.artistRepo.SetCounter(ctx, id, counter.counterName, count); err != nil {
					log.Printf("艺术家(%s) %s 更新失败: %v", artistIDStr, counter.counterName, err)
				}
			}

			if taskProg != nil {
				mu.Lock()
				processedArtists++
				progress := baseProgress + 0.25*float32(processedArtists)/float32(totalArtists)
				taskProg.Mu.Lock()
				taskProg.ProcessedFiles = int32(float32(taskProg.TotalFiles) * progress)
				taskProg.Mu.Unlock()
				mu.Unlock()
			}
		}(artistID)
	}

	wg.Wait()
	log.Printf("艺术家统计更新完成，共处理 %d 个艺术家", len(artistIDs))
	return nil
}

func (uc *AudioStatisticsUsecase) getArtistCounterByTraversal(ctx context.Context, artistIDStr, counterName string) (int, error) {
	switch counterName {
	case "album_count":
		return uc.countAlbumsWithMediaByArtist(ctx, artistIDStr, false)
	case "guest_album_count":
		return uc.countAlbumsWithMediaByArtist(ctx, artistIDStr, true)
	case "song_count":
		count, err := uc.mediaRepo.MediaCountByArtist(ctx, artistIDStr)
		if err != nil {
			return 0, err
		}
		return int(count), nil
	case "guest_song_count":
		count, err := uc.mediaRepo.GuestMediaCountByArtist(ctx, artistIDStr)
		if err != nil {
			return 0, err
		}
		return int(count), nil
	case "cue_count":
		count, err := uc.mediaCueRepo.MediaCueCountByArtist(ctx, artistIDStr)
		if err != nil {
			return 0, err
		}
		return int(count), nil
	case "guest_cue_count":
		count, err := uc.mediaCueRepo.GuestMediaCueCountByArtist(ctx, artistIDStr)
		if err != nil {
			return 0, err
		}
		return int(count), nil
	default:
		return 0, fmt.Errorf("未知的计数器名称: %s", counterName)
	}
}

func (uc *AudioStatisticsUsecase) countAlbumsWithMediaByArtist(ctx context.Context, artistIDStr string, isGuest bool) (int, error) {
	var albumIDs []primitive.ObjectID

	if isGuest {
		guestAlbums, err := uc.albumRepo.GetArtistGuestAlbumsMap(ctx)
		if err != nil {
			return 0, err
		}
		for id, albums := range guestAlbums {
			if id.Hex() == artistIDStr {
				albumIDs = albums
				break
			}
		}
	} else {
		artistAlbums, err := uc.albumRepo.GetArtistAlbumsMap(ctx)
		if err != nil {
			return 0, err
		}
		for id, albums := range artistAlbums {
			if id.Hex() == artistIDStr {
				albumIDs = albums
				break
			}
		}
	}

	if len(albumIDs) == 0 {
		return 0, nil
	}

	count := 0
	for _, albumID := range albumIDs {
		mediaCount, err := uc.mediaRepo.MediaCountByAlbum(ctx, albumID.Hex())
		if err != nil {
			continue
		}
		if mediaCount > 0 {
			count++
		}
	}

	return count, nil
}

func (uc *AudioStatisticsUsecase) UpdateAlbumStatisticsByTraversal(ctx context.Context, taskProg *domain_util.TaskProgress, baseProgress float32) error {
	log.Printf("开始通过遍历方式更新专辑统计")

	albumIDs, err := uc.albumRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取专辑ID列表失败: %w", err)
	}

	if len(albumIDs) == 0 {
		log.Printf("没有专辑需要统计")
		return nil
	}

	totalAlbums := len(albumIDs)
	var processedAlbums int32
	var mu sync.Mutex

	maxConcurrency := 50
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, albumID := range albumIDs {
		wg.Add(1)
		go func(id primitive.ObjectID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			albumIDStr := id.Hex()

			count, err := uc.mediaRepo.MediaCountByAlbum(ctx, albumIDStr)
			if err != nil {
				log.Printf("专辑(%s) song_count 统计失败: %v", albumIDStr, err)
				return
			}

			update := bson.M{"$set": bson.M{"song_count": int(count)}}
			if _, err := uc.albumRepo.UpdateByID(ctx, id, update); err != nil {
				log.Printf("专辑(%s) song_count 更新失败: %v", albumIDStr, err)
			}

			if taskProg != nil {
				mu.Lock()
				processedAlbums++
				progress := baseProgress + 0.25*float32(processedAlbums)/float32(totalAlbums)
				taskProg.Mu.Lock()
				taskProg.ProcessedFiles = int32(float32(taskProg.TotalFiles) * progress)
				taskProg.Mu.Unlock()
				mu.Unlock()
			}
		}(albumID)
	}

	wg.Wait()
	log.Printf("专辑统计更新完成，共处理 %d 个专辑", len(albumIDs))
	return nil
}

func (uc *AudioStatisticsUsecase) ResetAllArtistCounters(ctx context.Context) error {
	log.Printf("重置所有艺术家计数器")
	_, err := uc.artistRepo.ResetALLField(ctx)
	return err
}

func (uc *AudioStatisticsUsecase) ResetAllAlbumCounters(ctx context.Context) error {
	log.Printf("重置所有专辑计数器")
	_, err := uc.albumRepo.ResetALLField(ctx)
	return err
}

func (uc *AudioStatisticsUsecase) RecalculateAllStatistics(ctx context.Context, taskProg *domain_util.TaskProgress) error {
	log.Printf("开始完整重算所有统计（清零后遍历累加）")

	if taskProg == nil {
		return nil
	}

	taskProg.Mu.Lock()
	taskProg.Stage = "statistics"
	taskProg.Mu.Unlock()

	if err := uc.ResetAllArtistCounters(ctx); err != nil {
		return fmt.Errorf("重置艺术家计数器失败: %w", err)
	}

	if err := uc.ResetAllAlbumCounters(ctx); err != nil {
		return fmt.Errorf("重置专辑计数器失败: %w", err)
	}

	if err := uc.UpdateArtistStatisticsByTraversal(ctx, taskProg, 0.5); err != nil {
		return fmt.Errorf("更新艺术家统计失败: %w", err)
	}

	taskProg.Mu.Lock()
	taskProg.ProcessedFiles = int32(float32(taskProg.TotalFiles) * 0.75)
	taskProg.Mu.Unlock()

	if err := uc.UpdateAlbumStatisticsByTraversal(ctx, taskProg, 0.75); err != nil {
		return fmt.Errorf("更新专辑统计失败: %w", err)
	}

	taskProg.Mu.Lock()
	taskProg.ProcessedFiles = taskProg.TotalFiles
	taskProg.Mu.Unlock()

	log.Printf("完整统计重算完成")
	return nil
}

func (uc *AudioStatisticsUsecase) UpdateStatisticsIncremental(ctx context.Context, taskProg *domain_util.TaskProgress) error {
	log.Printf("开始增量更新统计（不清零，只更新变化的统计）")

	if taskProg == nil {
		return nil
	}

	taskProg.Mu.Lock()
	taskProg.Stage = "statistics"
	taskProg.ProcessedFiles = int32(float32(taskProg.TotalFiles) * 0.5)
	taskProg.Mu.Unlock()

	if err := uc.UpdateArtistStatisticsByTraversal(ctx, taskProg, 0.5); err != nil {
		log.Printf("增量更新艺术家统计失败: %v", err)
	}

	taskProg.Mu.Lock()
	taskProg.ProcessedFiles = int32(float32(taskProg.TotalFiles) * 0.75)
	taskProg.Mu.Unlock()

	if err := uc.UpdateAlbumStatisticsByTraversal(ctx, taskProg, 0.75); err != nil {
		log.Printf("增量更新专辑统计失败: %v", err)
	}

	taskProg.Mu.Lock()
	taskProg.ProcessedFiles = taskProg.TotalFiles
	taskProg.Mu.Unlock()

	log.Printf("增量统计更新完成")
	return nil
}

func (uc *AudioStatisticsUsecase) UpdateAllArtistStatistics(ctx context.Context) error {
	return uc.UpdateArtistStatisticsByTraversal(ctx, nil, 0)
}

func (uc *AudioStatisticsUsecase) UpdateAllAlbumStatistics(ctx context.Context) error {
	return uc.UpdateAlbumStatisticsByTraversal(ctx, nil, 0)
}

func (uc *AudioStatisticsUsecase) CleanupInvalidArtists(ctx context.Context) (int, error) {
	artistIDs, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return 0, err
	}

	var deletedCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, artistID := range artistIDs {
		wg.Add(1)
		go func(id primitive.ObjectID) {
			defer wg.Done()

			artistIDStr := id.Hex()

			mediaCount, _ := uc.mediaRepo.MediaCountByArtist(ctx, artistIDStr)
			guestMediaCount, _ := uc.mediaRepo.GuestMediaCountByArtist(ctx, artistIDStr)
			cueCount, _ := uc.mediaCueRepo.MediaCueCountByArtist(ctx, artistIDStr)
			guestCueCount, _ := uc.mediaCueRepo.GuestMediaCueCountByArtist(ctx, artistIDStr)

			if mediaCount == 0 && guestMediaCount == 0 && cueCount == 0 && guestCueCount == 0 {
				if err := uc.artistRepo.DeleteByID(ctx, id); err != nil {
					log.Printf("删除无效艺术家失败 (ID %s): %v", artistIDStr, err)
				} else {
					mu.Lock()
					deletedCount++
					mu.Unlock()
				}
			}
		}(artistID)
	}

	wg.Wait()
	return deletedCount, nil
}

func (uc *AudioStatisticsUsecase) CleanupInvalidAlbums(ctx context.Context) (int, error) {
	albumIDs, err := uc.albumRepo.GetAllIDs(ctx)
	if err != nil {
		return 0, err
	}

	var deletedCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, albumID := range albumIDs {
		wg.Add(1)
		go func(id primitive.ObjectID) {
			defer wg.Done()

			albumIDStr := id.Hex()
			mediaCount, _ := uc.mediaRepo.MediaCountByAlbum(ctx, albumIDStr)

			if mediaCount == 0 {
				if err := uc.albumRepo.DeleteByID(ctx, id); err != nil {
					log.Printf("删除无效专辑失败 (ID %s): %v", albumIDStr, err)
				} else {
					mu.Lock()
					deletedCount++
					mu.Unlock()
				}
			}
		}(albumID)
	}

	wg.Wait()
	return deletedCount, nil
}

func (uc *AudioStatisticsUsecase) PerformFullStatistics(ctx context.Context, totalFiles int32, scanProgress *float32, scanMutex sync.Locker) error {
	return uc.RecalculateAllStatistics(ctx, nil)
}

var processedFiles int32

func (uc *AudioStatisticsUsecase) GetArtistStatistics(ctx context.Context, artistID string) (*ArtistStatistics, error) {
	albumCount, err := uc.albumRepo.AlbumCountByArtist(ctx, artistID)
	if err != nil {
		return nil, err
	}
	guestAlbumCount, err := uc.albumRepo.GuestAlbumCountByArtist(ctx, artistID)
	if err != nil {
		return nil, err
	}
	songCount, err := uc.mediaRepo.MediaCountByArtist(ctx, artistID)
	if err != nil {
		return nil, err
	}
	guestSongCount, err := uc.mediaRepo.GuestMediaCountByArtist(ctx, artistID)
	if err != nil {
		return nil, err
	}
	cueCount, err := uc.mediaCueRepo.MediaCueCountByArtist(ctx, artistID)
	if err != nil {
		return nil, err
	}
	guestCueCount, err := uc.mediaCueRepo.GuestMediaCueCountByArtist(ctx, artistID)
	if err != nil {
		return nil, err
	}

	return &ArtistStatistics{
		AlbumCount:      int(albumCount),
		GuestAlbumCount: int(guestAlbumCount),
		SongCount:       int(songCount),
		GuestSongCount:  int(guestSongCount),
		CueCount:        int(cueCount),
		GuestCueCount:   int(guestCueCount),
	}, nil
}

func (uc *AudioStatisticsUsecase) GetAlbumStatistics(ctx context.Context, albumID string) (*AlbumStatistics, error) {
	songCount, err := uc.mediaRepo.MediaCountByAlbum(ctx, albumID)
	if err != nil {
		return nil, err
	}

	return &AlbumStatistics{
		SongCount: int(songCount),
	}, nil
}

type ArtistStatistics struct {
	AlbumCount      int
	GuestAlbumCount int
	SongCount       int
	GuestSongCount  int
	CueCount        int
	GuestCueCount   int
}

type AlbumStatistics struct {
	SongCount int
}

func (uc *AudioStatisticsUsecase) RefactorArtistAlbums(ctx context.Context) error {
	log.Printf("开始重构艺术家-专辑映射关系")

	artistAlbums, err := uc.albumRepo.GetArtistAlbumsMap(ctx)
	if err != nil {
		return fmt.Errorf("获取艺术家专辑映射失败: %w", err)
	}

	artistGuestAlbums, err := uc.albumRepo.GetArtistGuestAlbumsMap(ctx)
	if err != nil {
		return fmt.Errorf("获取艺术家合作专辑映射失败: %w", err)
	}

	maxConcurrency := 50
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for artistID, albums := range artistAlbums {
		wg.Add(1)
		go func(id primitive.ObjectID, albumIDs []primitive.ObjectID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			artistIDStr := id.Hex()

			var artistIDPairs []scene_audio_db_models.ArtistIDPair
			for _, albumID := range albumIDs {
				album, err := uc.albumRepo.GetByID(ctx, albumID)
				if err != nil {
					continue
				}
				if album != nil {
					artistIDPairs = append(artistIDPairs, scene_audio_db_models.ArtistIDPair{
						ArtistName: album.Artist,
						ArtistID:   album.ArtistID,
					})
				}
			}

			if err := uc.artistRepo.UpdateArtistAlbums(ctx, id, artistIDPairs); err != nil {
				log.Printf("更新艺术家专辑关联失败 (ID %s): %v", artistIDStr, err)
			}
		}(artistID, albums)
	}

	for artistID, albums := range artistGuestAlbums {
		wg.Add(1)
		go func(id primitive.ObjectID, albumIDs []primitive.ObjectID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			artistIDStr := id.Hex()

			var artistIDPairs []scene_audio_db_models.ArtistIDPair
			for _, albumID := range albumIDs {
				album, err := uc.albumRepo.GetByID(ctx, albumID)
				if err != nil {
					continue
				}
				if album != nil {
					artistIDPairs = append(artistIDPairs, scene_audio_db_models.ArtistIDPair{
						ArtistName: album.Artist,
						ArtistID:   album.ArtistID,
					})
				}
			}

			if err := uc.artistRepo.UpdateArtistGuestAlbums(ctx, id, artistIDPairs); err != nil {
				log.Printf("更新艺术家合作专辑关联失败 (ID %s): %v", artistIDStr, err)
			}
		}(artistID, albums)
	}

	wg.Wait()
	log.Printf("艺术家-专辑映射关系重构完成")
	return nil
}

func (uc *AudioStatisticsUsecase) RebuildArtistAlbumRelations(ctx context.Context) error {
	log.Printf("开始重建艺术家-专辑关联关系")

	albumIDs, err := uc.albumRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取所有专辑ID失败: %w", err)
	}

	artistAlbumMap := make(map[primitive.ObjectID][]primitive.ObjectID)
	artistGuestAlbumMap := make(map[primitive.ObjectID][]primitive.ObjectID)

	for _, albumID := range albumIDs {
		album, err := uc.albumRepo.GetByID(ctx, albumID)
		if err != nil || album == nil {
			continue
		}

		if album.ArtistID != "" {
			artistID, err := primitive.ObjectIDFromHex(album.ArtistID)
			if err == nil {
				artistAlbumMap[artistID] = append(artistAlbumMap[artistID], album.ID)
			}
		}

		if len(album.AllArtistIDs) > 0 {
			for _, artistIDPair := range album.AllArtistIDs {
				if artistIDPair.ArtistID != "" && artistIDPair.ArtistID != album.ArtistID {
					artistID, err := primitive.ObjectIDFromHex(artistIDPair.ArtistID)
					if err == nil {
						artistGuestAlbumMap[artistID] = append(artistGuestAlbumMap[artistID], album.ID)
					}
				}
			}
		}
	}

	maxConcurrency := 50
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for artistID, albumIDs := range artistAlbumMap {
		wg.Add(1)
		go func(id primitive.ObjectID, ids []primitive.ObjectID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var pairs []scene_audio_db_models.ArtistIDPair
			for _, albumID := range ids {
				album, err := uc.albumRepo.GetByID(ctx, albumID)
				if err != nil || album == nil {
					continue
				}
				pairs = append(pairs, scene_audio_db_models.ArtistIDPair{
					ArtistName: album.Artist,
					ArtistID:   album.ArtistID,
				})
			}

			if err := uc.artistRepo.UpdateArtistAlbums(ctx, id, pairs); err != nil {
				log.Printf("更新艺术家(%s)专辑关联失败: %v", id.Hex(), err)
			}
		}(artistID, albumIDs)
	}

	for artistID, albumIDs := range artistGuestAlbumMap {
		wg.Add(1)
		go func(id primitive.ObjectID, ids []primitive.ObjectID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var pairs []scene_audio_db_models.ArtistIDPair
			for _, albumID := range ids {
				album, err := uc.albumRepo.GetByID(ctx, albumID)
				if err != nil || album == nil {
					continue
				}
				pairs = append(pairs, scene_audio_db_models.ArtistIDPair{
					ArtistName: album.Artist,
					ArtistID:   album.ArtistID,
				})
			}

			if err := uc.artistRepo.UpdateArtistGuestAlbums(ctx, id, pairs); err != nil {
				log.Printf("更新艺术家(%s)合作专辑关联失败: %v", id.Hex(), err)
			}
		}(artistID, albumIDs)
	}

	wg.Wait()

	if err := uc.UpdateArtistStatisticsByTraversal(ctx, nil, 0); err != nil {
		log.Printf("更新艺术家统计失败: %v", err)
	}

	log.Printf("艺术家-专辑关联关系重建完成")
	return nil
}

func (uc *AudioStatisticsUsecase) GetAllArtistsWithZeroCounts(ctx context.Context) ([]primitive.ObjectID, error) {
	artistIDs, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return nil, err
	}

	var zeroCountArtists []primitive.ObjectID

	for _, artistID := range artistIDs {
		artistIDStr := artistID.Hex()

		mediaCount, _ := uc.mediaRepo.MediaCountByArtist(ctx, artistIDStr)
		guestMediaCount, _ := uc.mediaRepo.GuestMediaCountByArtist(ctx, artistIDStr)
		cueCount, _ := uc.mediaCueRepo.MediaCueCountByArtist(ctx, artistIDStr)
		guestCueCount, _ := uc.mediaCueRepo.GuestMediaCueCountByArtist(ctx, artistIDStr)
		albumCount, _ := uc.albumRepo.AlbumCountByArtist(ctx, artistIDStr)
		guestAlbumCount, _ := uc.albumRepo.GuestAlbumCountByArtist(ctx, artistIDStr)

		if mediaCount == 0 && guestMediaCount == 0 && cueCount == 0 && guestCueCount == 0 && albumCount == 0 && guestAlbumCount == 0 {
			zeroCountArtists = append(zeroCountArtists, artistID)
		}
	}

	return zeroCountArtists, nil
}

func (uc *AudioStatisticsUsecase) GetAllAlbumsWithZeroCounts(ctx context.Context) ([]primitive.ObjectID, error) {
	albumIDs, err := uc.albumRepo.GetAllIDs(ctx)
	if err != nil {
		return nil, err
	}

	var zeroCountAlbums []primitive.ObjectID

	for _, albumID := range albumIDs {
		albumIDStr := albumID.Hex()
		mediaCount, _ := uc.mediaRepo.MediaCountByAlbum(ctx, albumIDStr)

		if mediaCount == 0 {
			zeroCountAlbums = append(zeroCountAlbums, albumID)
		}
	}

	return zeroCountAlbums, nil
}

func (uc *AudioStatisticsUsecase) BatchDeleteInvalidArtists(ctx context.Context, artistIDs []primitive.ObjectID) (int, error) {
	if len(artistIDs) == 0 {
		return 0, nil
	}

	var deletedCount int
	for _, artistID := range artistIDs {
		if err := uc.artistRepo.DeleteByID(ctx, artistID); err != nil {
			log.Printf("删除艺术家失败 (ID %s): %v", artistID.Hex(), err)
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}

func (uc *AudioStatisticsUsecase) BatchDeleteInvalidAlbums(ctx context.Context, albumIDs []primitive.ObjectID) (int, error) {
	if len(albumIDs) == 0 {
		return 0, nil
	}

	var deletedCount int
	for _, albumID := range albumIDs {
		if err := uc.albumRepo.DeleteByID(ctx, albumID); err != nil {
			log.Printf("删除专辑失败 (ID %s): %v", albumID.Hex(), err)
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}
