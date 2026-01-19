package scene_audio_db_models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaLibrarySyncRecord 媒体库同步记录模型
// 用于跟踪音频文件的上传、下载和同步状态
type MediaLibrarySyncRecord struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MediaFileID         primitive.ObjectID `bson:"media_file_id" json:"media_file_id" validate:"required"`                   // 关联的媒体文件ID
	MediaLibraryAudioID primitive.ObjectID `bson:"media_library_audio_id" json:"media_library_audio_id" validate:"required"` // 关联的媒体库音频文件ID
	SyncStatus          string             `bson:"sync_status" json:"sync_status" validate:"required"`                       // 同步状态: pending, uploading, downloading, synced, failed
	SyncType            string             `bson:"sync_type" json:"sync_type" validate:"required"`                           // 同步类型: upload, download, both
	LastSyncTime        primitive.DateTime `bson:"last_sync_time" json:"last_sync_time"`                                     // 最后同步时间
	CreatedAt           primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt           primitive.DateTime `bson:"updated_at" json:"updated_at"`
	ErrorMessage        string             `bson:"error_message,omitempty" json:"error_message,omitempty"` // 错误信息（如果有）
	Progress            float64            `bson:"progress" json:"progress"`                               // 同步进度百分比 (0-100)
	FileSize            int64              `bson:"file_size" json:"file_size"`                             // 文件大小
	FilePath            string             `bson:"file_path" json:"file_path"`                             // 文件路径
	FileName            string             `bson:"file_name" json:"file_name"`                             // 文件名
	Checksum            string             `bson:"checksum" json:"checksum"`                               // 文件校验和
}

// Validate 验证同步记录的有效性
func (m *MediaLibrarySyncRecord) Validate() error {
	if m.MediaFileID.IsZero() {
		return fmt.Errorf("media file ID is required")
	}
	if m.MediaLibraryAudioID.IsZero() {
		return fmt.Errorf("media library audio ID is required")
	}
	if m.SyncStatus == "" {
		return fmt.Errorf("sync status is required")
	}
	if m.SyncType == "" {
		return fmt.Errorf("sync type is required")
	}
	if m.CreatedAt == 0 {
		m.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	if m.UpdatedAt == 0 {
		m.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	return nil
}
