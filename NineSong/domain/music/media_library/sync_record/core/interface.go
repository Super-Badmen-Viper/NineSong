package core

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaLibrarySyncRecordRepository 媒体库同步记录仓库接口
// 定义了对媒体库同步记录进行CRUD操作的方法
type MediaLibrarySyncRecordRepository interface {
	shared.BaseRepository[MediaLibrarySyncRecord]

	// FindByMediaFileID 根据媒体文件ID查找同步记录
	FindByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (*MediaLibrarySyncRecord, error)

	// FindByMediaLibraryAudioID 根据媒体库音频文件ID查找同步记录
	FindByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (*MediaLibrarySyncRecord, error)

	// FindBySyncStatus 根据同步状态查找同步记录
	FindBySyncStatus(ctx context.Context, syncStatus string) ([]*MediaLibrarySyncRecord, error)

	// FindBySyncType 根据同步类型查找同步记录
	FindBySyncType(ctx context.Context, syncType string) ([]*MediaLibrarySyncRecord, error)

	// UpdateSyncStatus 更新同步状态
	UpdateSyncStatus(ctx context.Context, id primitive.ObjectID, status string, progress float64, errorMessage string) error

	// UpdateSyncProgress 更新同步进度
	UpdateSyncProgress(ctx context.Context, id primitive.ObjectID, progress float64, lastUpdateTime primitive.DateTime) error

	// UpdateSyncResult 更新同步结果
	UpdateSyncResult(ctx context.Context, id primitive.ObjectID, status string, progress float64, errorMessage string, filePath string, fileSize int64, checksum string) error

	// DeleteByMediaFileID 根据媒体文件ID删除同步记录
	DeleteByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) error

	// DeleteByMediaLibraryAudioID 根据媒体库音频文件ID删除同步记录
	DeleteByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) error

	// ExistsByMediaFileID 检查是否已存在关联指定媒体文件ID的同步记录
	ExistsByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (bool, error)

	// ExistsByMediaLibraryAudioID 检查是否已存在关联指定媒体库音频文件ID的同步记录
	ExistsByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (bool, error)

	// GetSyncStatistics 获取同步统计信息
	GetSyncStatistics(ctx context.Context) (map[string]interface{}, error)
}
