package core

import "go.mongodb.org/mongo-driver/bson/primitive"

// MediaLibraryAudioResponse represents the response model for media library audio operations
type MediaLibraryAudioResponse struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FileName    string             `json:"file_name" bson:"file_name"`
	FilePath    string             `json:"file_path" bson:"file_path"`
	StoragePath string             `json:"storage_path" bson:"storage_path"`
	FileSize    int64              `json:"file_size" bson:"file_size"`
	FileType    string             `json:"file_type" bson:"file_type"`
	Checksum    string             `json:"checksum" bson:"checksum"`
	LibraryID   primitive.ObjectID `json:"library_id" bson:"library_id"`
	UploaderID  primitive.ObjectID `json:"uploader_id" bson:"uploader_id"`
	UploadTime  primitive.DateTime `json:"upload_time" bson:"upload_time"`
	CreatedAt   primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt   primitive.DateTime `json:"updated_at" bson:"updated_at"`
	Status      string             `json:"status,omitempty" bson:"status,omitempty"`     // Upload/download status
	Progress    float64            `json:"progress,omitempty" bson:"progress,omitempty"` // Upload/download progress percentage
}

// FileInfo 文件信息结构
type FileInfo struct {
	ID         string `json:"id"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	FileType   string `json:"file_type"`
	MimeType   string `json:"mime_type"`
	Format     string `json:"format"`
	Duration   float64 `json:"duration"`
	Bitrate    int    `json:"bitrate"`
	SampleRate int    `json:"sample_rate"`
	Channels   int    `json:"channels"`
	Encoding   string `json:"encoding"`
	Artist     string `json:"artist"`
	Title      string `json:"title"`
	Album      string `json:"album"`
	Year       int    `json:"year"`
	Genre      string `json:"genre"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
