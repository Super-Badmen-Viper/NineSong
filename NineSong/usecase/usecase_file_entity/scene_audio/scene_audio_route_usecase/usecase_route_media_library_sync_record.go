package scene_audio_route_usecase

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_repository_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
)

// MediaLibrarySyncRecordUseCase 媒体库同步记录用例
// 负责处理与媒体库同步记录相关的业务逻辑
type MediaLibrarySyncRecordUseCase struct {
	mediaLibrarySyncRecordRepo scene_audio_repository_interface.MediaLibrarySyncRecordRepository
	timeout                    time.Duration
}

// NewMediaLibrarySyncRecordUseCase 创建新的媒体库同步记录用例实例
func NewMediaLibrarySyncRecordUseCase(
	mediaLibrarySyncRecordRepo scene_audio_repository_interface.MediaLibrarySyncRecordRepository,
	timeout time.Duration,
) *MediaLibrarySyncRecordUseCase {
	return &MediaLibrarySyncRecordUseCase{
		mediaLibrarySyncRecordRepo: mediaLibrarySyncRecordRepo,
		timeout:                    timeout,
	}
}

// CreateSyncRecord 创建媒体库同步记录
func (uc *MediaLibrarySyncRecordUseCase) CreateSyncRecord(ctx context.Context, syncRecord *scene_audio_db_models.MediaLibrarySyncRecord) (*scene_audio_route_models.MediaLibrarySyncRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if err := syncRecord.Validate(); err != nil {
		return nil, err
	}

	err := uc.mediaLibrarySyncRecordRepo.Create(ctx, syncRecord)
	if err != nil {
		return nil, err
	}

	// 转换为路由模型返回
	routeModel := scene_audio_route_models.ConvertDBToRouteModel(syncRecord)
	return routeModel, nil
}

// UpdateSyncRecord 更新媒体库同步记录
func (uc *MediaLibrarySyncRecordUseCase) UpdateSyncRecord(ctx context.Context, id string, syncRecord *scene_audio_db_models.MediaLibrarySyncRecord) (*scene_audio_route_models.MediaLibrarySyncRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	if err := syncRecord.Validate(); err != nil {
		return nil, err
	}

	err = uc.mediaLibrarySyncRecordRepo.UpdateByID(ctx, objectID, syncRecord)
	if err != nil {
		return nil, err
	}

	// 转换为路由模型返回
	routeModel := scene_audio_route_models.ConvertDBToRouteModel(syncRecord)
	return routeModel, nil
}

// GetSyncRecordByID 根据ID获取同步记录
func (uc *MediaLibrarySyncRecordUseCase) GetSyncRecordByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_route_models.MediaLibrarySyncRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	result, err := uc.mediaLibrarySyncRecordRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync record by ID: %w", err)
	}

	// 转换为路由模型返回
	routeModel := scene_audio_route_models.ConvertDBToRouteModel(result)
	return routeModel, nil
}

// GetSyncRecordByMediaFileID 根据媒体文件ID获取同步记录
func (uc *MediaLibrarySyncRecordUseCase) GetSyncRecordByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (*scene_audio_route_models.MediaLibrarySyncRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	result, err := uc.mediaLibrarySyncRecordRepo.FindByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync record by media file ID: %w", err)
	}

	// 转换为路由模型返回
	routeModel := scene_audio_route_models.ConvertDBToRouteModel(result)
	return routeModel, nil
}

// GetSyncRecordByMediaLibraryAudioID 根据媒体库音频文件ID获取同步记录
func (uc *MediaLibrarySyncRecordUseCase) GetSyncRecordByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (*scene_audio_route_models.MediaLibrarySyncRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	result, err := uc.mediaLibrarySyncRecordRepo.FindByMediaLibraryAudioID(ctx, mediaLibraryAudioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync record by media library audio ID: %w", err)
	}

	// 转换为路由模型返回
	routeModel := scene_audio_route_models.ConvertDBToRouteModel(result)
	return routeModel, nil
}

// GetSyncRecordsByStatus 根据状态获取同步记录列表
func (uc *MediaLibrarySyncRecordUseCase) GetSyncRecordsByStatus(ctx context.Context, status string) ([]*scene_audio_route_models.MediaLibrarySyncRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	dbResults, err := uc.mediaLibrarySyncRecordRepo.FindBySyncStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync records by status: %w", err)
	}

	// 转换为路由模型返回
	var routeResults []*scene_audio_route_models.MediaLibrarySyncRecord
	for _, dbResult := range dbResults {
		routeModel := scene_audio_route_models.ConvertDBToRouteModel(dbResult)
		routeResults = append(routeResults, routeModel)
	}
	return routeResults, nil
}

// GetSyncRecordsByType 根据同步类型获取同步记录列表
func (uc *MediaLibrarySyncRecordUseCase) GetSyncRecordsByType(ctx context.Context, syncType string) ([]*scene_audio_route_models.MediaLibrarySyncRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	dbResults, err := uc.mediaLibrarySyncRecordRepo.FindBySyncType(ctx, syncType)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync records by type: %w", err)
	}

	// 转换为路由模型返回
	var routeResults []*scene_audio_route_models.MediaLibrarySyncRecord
	for _, dbResult := range dbResults {
		routeModel := scene_audio_route_models.ConvertDBToRouteModel(dbResult)
		routeResults = append(routeResults, routeModel)
	}
	return routeResults, nil
}

// DeleteSyncRecordByID 根据ID删除同步记录
func (uc *MediaLibrarySyncRecordUseCase) DeleteSyncRecordByID(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	err := uc.mediaLibrarySyncRecordRepo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete sync record by ID: %w", err)
	}

	return nil
}

// UpdateSyncStatus 更新同步状态
func (uc *MediaLibrarySyncRecordUseCase) UpdateSyncStatus(ctx context.Context, id primitive.ObjectID, status string, progress float64, errorMessage string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	err := uc.mediaLibrarySyncRecordRepo.UpdateSyncStatus(ctx, id, status, progress, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	return nil
}

// UpdateSyncProgress 更新同步进度
func (uc *MediaLibrarySyncRecordUseCase) UpdateSyncProgress(ctx context.Context, id primitive.ObjectID, progress float64, lastUpdateTime primitive.DateTime) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	err := uc.mediaLibrarySyncRecordRepo.UpdateSyncProgress(ctx, id, progress, lastUpdateTime)
	if err != nil {
		return fmt.Errorf("failed to update sync progress: %w", err)
	}

	return nil
}

// UpdateSyncResult 更新同步结果
func (uc *MediaLibrarySyncRecordUseCase) UpdateSyncResult(ctx context.Context, id primitive.ObjectID, status string, progress float64, errorMessage string, filePath string, fileSize int64, checksum string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	err := uc.mediaLibrarySyncRecordRepo.UpdateSyncResult(ctx, id, status, progress, errorMessage, filePath, fileSize, checksum)
	if err != nil {
		return fmt.Errorf("failed to update sync result: %w", err)
	}

	return nil
}

// DeleteSyncRecordByMediaFileID 根据媒体文件ID删除同步记录
func (uc *MediaLibrarySyncRecordUseCase) DeleteSyncRecordByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	err := uc.mediaLibrarySyncRecordRepo.DeleteByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return fmt.Errorf("failed to delete sync record by media file ID: %w", err)
	}

	return nil
}

// DeleteSyncRecordByMediaLibraryAudioID 根据媒体库音频文件ID删除同步记录
func (uc *MediaLibrarySyncRecordUseCase) DeleteSyncRecordByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	err := uc.mediaLibrarySyncRecordRepo.DeleteByMediaLibraryAudioID(ctx, mediaLibraryAudioID)
	if err != nil {
		return fmt.Errorf("failed to delete sync record by media library audio ID: %w", err)
	}

	return nil
}

// GetSyncStatistics 获取同步统计信息
func (uc *MediaLibrarySyncRecordUseCase) GetSyncStatistics(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	stats, err := uc.mediaLibrarySyncRecordRepo.GetSyncStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync statistics: %w", err)
	}

	return stats, nil
}

// ExistsByMediaFileID 检查是否已存在关联指定媒体文件ID的同步记录
func (uc *MediaLibrarySyncRecordUseCase) ExistsByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	exists, err := uc.mediaLibrarySyncRecordRepo.ExistsByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return false, fmt.Errorf("failed to check if sync record exists by media file ID: %w", err)
	}

	return exists, nil
}

// ExistsByMediaLibraryAudioID 检查是否已存在关联指定媒体库音频文件ID的同步记录
func (uc *MediaLibrarySyncRecordUseCase) ExistsByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	exists, err := uc.mediaLibrarySyncRecordRepo.ExistsByMediaLibraryAudioID(ctx, mediaLibraryAudioID)
	if err != nil {
		return false, fmt.Errorf("failed to check if sync record exists by media library audio ID: %w", err)
	}

	return exists, nil
}
