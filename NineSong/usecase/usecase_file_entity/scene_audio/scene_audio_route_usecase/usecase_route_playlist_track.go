package scene_audio_route_usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	usercase_audio_util "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type playlistTrackUsecase struct {
	repo         scene_audio_route_interface.PlaylistTrackRepository
	playlistRepo scene_audio_route_interface.PlaylistRepository
	tempRepo     scene_audio_db_interface.TempRepository
	timeout      time.Duration
}

func NewPlaylistTrackUsecase(
	repo scene_audio_route_interface.PlaylistTrackRepository,
	playlistRepo scene_audio_route_interface.PlaylistRepository,
	tempRepo scene_audio_db_interface.TempRepository,
	timeout time.Duration,
) scene_audio_route_interface.PlaylistTrackRepository {
	return &playlistTrackUsecase{
		repo:         repo,
		playlistRepo: playlistRepo,
		tempRepo:     tempRepo,
		timeout:      timeout,
	}
}

func (uc *playlistTrackUsecase) GetPlaylistTrackItems(
	ctx context.Context,
	start, end, sort, order, search, starred, albumId, artistId, year, playlistId string,
	suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter string,
) ([]scene_audio_route_models.MediaFileMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 参数验证
	if err := validateObjectID("playlistId", playlistId); err != nil {
		return nil, err
	}

	if _, err := strconv.Atoi(start); start != "" && err != nil {
		return nil, errors.New("invalid start parameter")
	}

	if _, err := strconv.Atoi(end); end != "" && err != nil {
		return nil, errors.New("invalid end parameter")
	}

	if albumId != "" {
		if err := validateObjectID("albumId", albumId); err != nil {
			return nil, err
		}
	}

	if artistId != "" {
		if err := validateObjectID("artistId", artistId); err != nil {
			return nil, err
		}
	}

	if year != "" {
		if _, err := strconv.Atoi(year); err != nil {
			return nil, errors.New("invalid year format")
		}
	}

	if starred != "" {
		if _, err := strconv.ParseBool(starred); err != nil {
			return nil, errors.New("invalid starred value")
		}
	}

	validSortFields := map[string]bool{
		"title":  true,
		"artist": true, "album": true,
		"year":       true,
		"rating":     true,
		"starred_at": true,
		"genre":      true,
		"play_count": true, "play_date": true,
		"duration": true, "bit_rate": true, "size": true,
		"created_at": true, "updated_at": true,
	}
	if !validSortFields[sort] {
		sort = "_id"
	}

	// 验证排序方向
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	return uc.repo.GetPlaylistTrackItems(
		ctx, start, end, sort, order,
		search, starred, albumId, artistId,
		year, playlistId,
		suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter,
	)
}

func (uc *playlistTrackUsecase) GetPlaylistTrackItemsMultipleSorting(
	ctx context.Context,
	start, end string,
	sortOrder []domain_util.SortOrder,
	search, starred, albumId, artistId, year, playlistId string,
	suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter string,
) ([]scene_audio_route_models.MediaFileMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 参数验证
	if err := validateObjectID("playlistId", playlistId); err != nil {
		return nil, err
	}

	if _, err := strconv.Atoi(start); start != "" && err != nil {
		return nil, errors.New("invalid start parameter")
	}

	if _, err := strconv.Atoi(end); end != "" && err != nil {
		return nil, errors.New("invalid end parameter")
	}

	if albumId != "" {
		if err := validateObjectID("albumId", albumId); err != nil {
			return nil, err
		}
	}

	if artistId != "" {
		if err := validateObjectID("artistId", artistId); err != nil {
			return nil, err
		}
	}

	if year != "" {
		if _, err := strconv.Atoi(year); err != nil {
			return nil, errors.New("invalid year format")
		}
	}

	if starred != "" {
		if _, err := strconv.ParseBool(starred); err != nil {
			return nil, errors.New("invalid starred value")
		}
	}

	// 验证排序参数
	validSortFields := map[string]bool{
		"index": true, "title": true, "artist": true, "album": true,
		"year": true, "rating": true, "starred_at": true, "genre": true,
		"play_count": true, "play_date": true, "duration": true,
		"bit_rate": true, "size": true, "created_at": true, "updated_at": true,
	}

	for _, so := range sortOrder {
		if !validSortFields[strings.ToLower(so.Sort)] {
			return nil, fmt.Errorf("invalid sort field: %s", so.Sort)
		}
		if so.Order != "asc" && so.Order != "desc" {
			return nil, fmt.Errorf("invalid sort order: %s", so.Order)
		}
	}

	return uc.repo.GetPlaylistTrackItemsMultipleSorting(
		ctx, start, end, sortOrder, search, starred, albumId, artistId, year, playlistId,
		suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter,
	)
}

func (uc *playlistTrackUsecase) GetPlaylistTrackFilterItemsCount(
	ctx context.Context,
) (*scene_audio_route_models.MediaFileFilterCounts, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	return uc.repo.GetPlaylistTrackFilterItemsCount(ctx)
}

func (uc *playlistTrackUsecase) AddPlaylistTrackItems(
	ctx context.Context,
	playlistId string,
	mediaFileIds string,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 参数验证
	if err := validateObjectID("playlistId", playlistId); err != nil {
		return false, err
	}

	if mediaFileIds == "" {
		return false, errors.New("empty media file ids")
	}

	// 验证媒体文件ID列表
	ids := strings.Split(mediaFileIds, ",")
	for _, id := range ids {
		if err := validateObjectID("mediaFileId", strings.TrimSpace(id)); err != nil {
			return false, err
		}
	}

	// 添加媒体项
	success, err := uc.repo.AddPlaylistTrackItems(ctx, playlistId, mediaFileIds)
	if err != nil {
		return false, err
	}

	// 只要成功添加了媒体项（无论播放列表是否为空），都触发封面生成
	if success {
		// 使用 goroutine 异步等待数据插入完成，然后生成封面
		go func() {
			// 等待一小段时间确保数据库插入完成
			time.Sleep(500 * time.Millisecond)
			uc.generatePlaylistCoverAsync(context.Background(), playlistId)
		}()
	}

	return success, nil
}

// generatePlaylistCoverAsync 异步生成播放列表封面
func (uc *playlistTrackUsecase) generatePlaylistCoverAsync(ctx context.Context, playlistId string) {
	go func() {
		coverCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// 获取封面基础路径
		coverBasePath, err := uc.tempRepo.GetTempPath(coverCtx, "cover")
		if err != nil {
			log.Printf("[ERROR] 生成播放列表封面失败: 获取封面基础路径失败: %v", err)
			return
		}

		// 获取第一首歌的封面路径
		firstTrackCoverPath, err := uc.playlistRepo.GetFirstTrackCoverImage(coverCtx, playlistId)
		if err != nil {
			log.Printf("[ERROR] 生成播放列表封面失败: 获取第一首歌封面失败: %v", err)
			return
		}

		if firstTrackCoverPath == "" {
			log.Printf("[WARN] 播放列表 %s 没有第一首歌封面，跳过封面生成", playlistId)
			return
		}

		// 创建封面生成器
		coverGenerator := usercase_audio_util.NewPlaylistCoverGenerator(coverBasePath)

		// 转换 playlistId 为 ObjectID
		objID, err := primitive.ObjectIDFromHex(playlistId)
		if err != nil {
			log.Printf("[ERROR] 生成播放列表封面失败: 无效的播放列表ID: %v", err)
			return
		}

		// 生成封面（总是强制重新生成）
		_, err = coverGenerator.GeneratePlaylistCover(
			coverCtx,
			firstTrackCoverPath,
			objID,
		)
		if err != nil {
			log.Printf("[ERROR] 生成播放列表 %s 封面失败: %v", playlistId, err)
			return
		}

		log.Printf("[INFO] 播放列表 %s 封面生成成功", playlistId)
	}()
}

func (uc *playlistTrackUsecase) RemovePlaylistTrackItems(
	ctx context.Context,
	playlistId string,
	mediaFileIds string,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 参数验证
	if err := validateObjectID("playlistId", playlistId); err != nil {
		return false, err
	}

	if mediaFileIds == "" {
		return false, errors.New("empty media file ids")
	}

	// 验证媒体文件ID列表
	ids := strings.Split(mediaFileIds, ",")
	for _, id := range ids {
		if err := validateObjectID("mediaFileId", strings.TrimSpace(id)); err != nil {
			return false, err
		}
	}

	// 删除媒体项
	success, err := uc.repo.RemovePlaylistTrackItems(ctx, playlistId, mediaFileIds)
	if err != nil {
		return false, err
	}

	// 只要成功删除了媒体项，都触发封面重新生成
	if success {
		// 使用 goroutine 异步等待数据删除完成，然后生成封面
		go func() {
			// 等待一小段时间确保数据库删除完成
			time.Sleep(500 * time.Millisecond)
			uc.generatePlaylistCoverAsync(context.Background(), playlistId)
		}()
	}

	return success, nil
}

func (uc *playlistTrackUsecase) SortPlaylistTrackItems(
	ctx context.Context,
	playlistId string,
	mediaFileIds string,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 参数验证
	if err := validateObjectID("playlistId", playlistId); err != nil {
		return false, err
	}

	if mediaFileIds == "" {
		return false, errors.New("empty media file ids")
	}

	// 验证媒体文件ID列表
	ids := strings.Split(mediaFileIds, ",")
	for _, id := range ids {
		if err := validateObjectID("mediaFileId", strings.TrimSpace(id)); err != nil {
			return false, err
		}
	}

	// 排序媒体项
	success, err := uc.repo.SortPlaylistTrackItems(ctx, playlistId, mediaFileIds)
	if err != nil {
		return false, err
	}

	// 只要成功排序了媒体项，都触发封面重新生成
	if success {
		// 使用 goroutine 异步等待数据更新完成，然后生成封面
		go func() {
			// 等待一小段时间确保数据库更新完成
			time.Sleep(500 * time.Millisecond)
			uc.generatePlaylistCoverAsync(context.Background(), playlistId)
		}()
	}

	return success, nil
}

func validateObjectID(field, value string) error {
	if _, err := primitive.ObjectIDFromHex(value); err != nil {
		return fmt.Errorf("invalid %s format", field)
	}
	return nil
}
