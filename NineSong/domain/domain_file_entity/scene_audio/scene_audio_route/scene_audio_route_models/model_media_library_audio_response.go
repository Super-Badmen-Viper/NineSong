package scene_audio_route_models

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
