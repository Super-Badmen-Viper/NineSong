package scene_audio_db_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChunkedUploadSession 分块上传会话模型
// 用于跟踪分块上传的状态和进度
type ChunkedUploadSession struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UploadID       string             `bson:"upload_id" json:"upload_id" validate:"required"`     // 上传会话唯一标识
	FileName       string             `bson:"file_name" json:"file_name" validate:"required"`     // 文件名
	FileSize       int64              `bson:"file_size" json:"file_size"`                         // 总文件大小
	TotalChunks    int                `bson:"total_chunks" json:"total_chunks"`                   // 总分块数
	UploadedChunks int                `bson:"uploaded_chunks" json:"uploaded_chunks"`             // 已上传分块数
	UploadedBytes  int64              `bson:"uploaded_bytes" json:"uploaded_bytes"`               // 已上传字节数
	Status         string             `bson:"status" json:"status" validate:"required"`           // 状态: pending, uploading, completed, failed
	LibraryID      primitive.ObjectID `bson:"library_id" json:"library_id" validate:"required"`   // 关联的媒体库ID
	UploaderID     primitive.ObjectID `bson:"uploader_id" json:"uploader_id" validate:"required"` // 上传者ID
	FilePath       string             `bson:"file_path" json:"file_path"`                         // 目标文件路径
	Checksum       string             `bson:"checksum" json:"checksum"`                           // 文件校验和
	CreatedAt      primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt      primitive.DateTime `bson:"updated_at" json:"updated_at"`
}

// Validate 验证分块上传会话实体
func (c *ChunkedUploadSession) Validate() error {
	if c.UploadID == "" {
		return ValidationError{"upload_id": "Upload ID cannot be empty"}
	}
	if c.FileName == "" {
		return ValidationError{"file_name": "File name cannot be empty"}
	}
	if c.Status == "" {
		return ValidationError{"status": "Status cannot be empty"}
	}
	if c.LibraryID.IsZero() {
		return ValidationError{"library_id": "Library ID cannot be empty"}
	}
	if c.UploaderID.IsZero() {
		return ValidationError{"uploader_id": "Uploader ID cannot be empty"}
	}
	if c.Status != "pending" && c.Status != "uploading" && c.Status != "completed" && c.Status != "failed" {
		return ValidationError{"status": "invalid status: " + c.Status}
	}
	return nil
}

// SetTimestamps 设置时间戳
func (c *ChunkedUploadSession) SetTimestamps() {
	now := primitive.NewDateTimeFromTime(time.Now())
	c.CreatedAt = now
	c.UpdatedAt = now
}

// ValidationError 自定义验证错误类型
type ValidationError map[string]string

func (e ValidationError) Error() string {
	for _, msg := range e {
		return msg
	}
	return "Validation error occurred"
}
