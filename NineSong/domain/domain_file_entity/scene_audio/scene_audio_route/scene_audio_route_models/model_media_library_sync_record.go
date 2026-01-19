package scene_audio_route_models

import (
	scene_audio_db_models "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaLibrarySyncRecord 媒体库同步记录响应模型
// 用于API响应中的媒体库同步记录数据传输
type MediaLibrarySyncRecord struct {
	ID                  primitive.ObjectID `json:"id"`
	MediaFileID         primitive.ObjectID `json:"media_file_id"`          // 关联的媒体文件ID
	MediaLibraryAudioID primitive.ObjectID `json:"media_library_audio_id"` // 关联的媒体库音频文件ID
	SyncStatus          string             `json:"sync_status"`            // 同步状态: pending, uploading, downloading, synced, failed
	SyncType            string             `json:"sync_type"`              // 同步类型: upload, download, both
	LastSyncTime        primitive.DateTime `json:"last_sync_time"`         // 最后同步时间
	CreatedAt           primitive.DateTime `json:"created_at"`
	UpdatedAt           primitive.DateTime `json:"updated_at"`
	ErrorMessage        string             `json:"error_message,omitempty"` // 错误信息（如果有）
	Progress            float64            `json:"progress"`                // 同步进度百分比 (0-100)
	FileSize            int64              `json:"file_size"`               // 文件大小
	FilePath            string             `json:"file_path"`               // 文件路径
	FileName            string             `json:"file_name"`               // 文件名
	Checksum            string             `json:"checksum"`                // 文件校验和
}

// ConvertDBToRouteModel 将数据库模型转换为路由响应模型
func ConvertDBToRouteModel(dbModel *scene_audio_db_models.MediaLibrarySyncRecord) *MediaLibrarySyncRecord {
	return &MediaLibrarySyncRecord{
		ID:                  dbModel.ID,
		MediaFileID:         dbModel.MediaFileID,
		MediaLibraryAudioID: dbModel.MediaLibraryAudioID,
		SyncStatus:          dbModel.SyncStatus,
		SyncType:            dbModel.SyncType,
		LastSyncTime:        dbModel.LastSyncTime,
		CreatedAt:           dbModel.CreatedAt,
		UpdatedAt:           dbModel.UpdatedAt,
		ErrorMessage:        dbModel.ErrorMessage,
		Progress:            dbModel.Progress,
		FileSize:            dbModel.FileSize,
		FilePath:            dbModel.FilePath,
		FileName:            dbModel.FileName,
		Checksum:            dbModel.Checksum,
	}
}
