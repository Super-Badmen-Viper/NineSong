package scene_audio_db_models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaLibraryAudio 音频文件实体模型，用于媒体库音频文件的上传下载功能
type MediaLibraryAudio struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt primitive.DateTime `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at"`

	// 基本文件信息
	FileName string `bson:"file_name" json:"file_name"` // 文件名
	FilePath string `bson:"file_path" json:"file_path"` // 文件存储路径
	FileSize int64  `bson:"file_size" json:"file_size"` // 文件大小（字节）
	FileType string `bson:"file_type" json:"file_type"` // 文件类型 (audio/mp3, audio/flac等)
	MimeType string `bson:"mime_type" json:"mime_type"` // MIME类型
	Checksum string `bson:"checksum" json:"checksum"`   // 文件校验和

	// 音频元数据
	Title      string  `bson:"title,omitempty" json:"title,omitempty"`             // 音频标题
	Artist     string  `bson:"artist,omitempty" json:"artist,omitempty"`           // 艺术家
	Album      string  `bson:"album,omitempty" json:"album,omitempty"`             // 专辑
	Duration   float64 `bson:"duration,omitempty" json:"duration,omitempty"`       // 持续时间（秒）
	BitRate    int     `bson:"bit_rate,omitempty" json:"bit_rate,omitempty"`       // 比特率
	SampleRate int     `bson:"sample_rate,omitempty" json:"sample_rate,omitempty"` // 采样率
	Channels   int     `bson:"channels,omitempty" json:"channels,omitempty"`       // 声道数

	// 媒体库信息
	LibraryID  primitive.ObjectID `bson:"library_id,omitempty" json:"library_id,omitempty"`   // 所属媒体库ID
	UploaderID primitive.ObjectID `bson:"uploader_id,omitempty" json:"uploader_id,omitempty"` // 上传者ID
	Status     string             `bson:"status" json:"status"`                               // 文件状态 (uploading, processing, ready, error)

	// 存储位置
	StoragePath   string             `bson:"storage_path" json:"storage_path"`                         // 存储路径
	IsUploaded    bool               `bson:"is_uploaded" json:"is_uploaded"`                           // 是否已上传完成
	UploadTime    primitive.DateTime `bson:"upload_time,omitempty" json:"upload_time,omitempty"`       // 上传时间
	DownloadCount int                `bson:"download_count,omitempty" json:"download_count,omitempty"` // 下载次数
}

// Validate 验证音频实体字段的有效性
func (e *MediaLibraryAudio) Validate() error {
	if e.FileName == "" {
		return errors.New("file name cannot be empty")
	}
	if e.FilePath == "" {
		return errors.New("file path cannot be empty")
	}
	if e.FileSize <= 0 {
		return errors.New("file size must be greater than 0")
	}
	return nil
}

// SetTimestamps 设置时间戳
func (e *MediaLibraryAudio) SetTimestamps() {
	now := primitive.NewDateTimeFromTime(time.Now())
	if e.ID.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	if e.UploadTime.Time().IsZero() && e.IsUploaded {
		e.UploadTime = now
	}
}
