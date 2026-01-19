package scene_audio_db_interface

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
)

// MediaLibraryAudioRepository 音频文件仓库接口，用于媒体库音频文件的上传下载功能
type MediaLibraryAudioRepository interface {
	domain.BaseRepository[scene_audio_db_models.MediaLibraryAudio]
	
	// 上传下载相关方法
	GetByLibraryID(ctx context.Context, libraryID primitive.ObjectID) ([]*scene_audio_db_models.MediaLibraryAudio, error)
	GetByUploaderID(ctx context.Context, uploaderID primitive.ObjectID) ([]*scene_audio_db_models.MediaLibraryAudio, error)
	GetByStatus(ctx context.Context, status string) ([]*scene_audio_db_models.MediaLibraryAudio, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
	UpdateFilePath(ctx context.Context, id primitive.ObjectID, filePath string) error
	GetByChecksum(ctx context.Context, checksum string) (*scene_audio_db_models.MediaLibraryAudio, error)
	
	// 文件操作相关方法
	GetByFileName(ctx context.Context, fileName string) (*scene_audio_db_models.MediaLibraryAudio, error)
	IncrementDownloadCount(ctx context.Context, id primitive.ObjectID) error
}