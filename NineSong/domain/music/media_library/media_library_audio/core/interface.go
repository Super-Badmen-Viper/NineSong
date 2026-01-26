package core

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
)

// MediaLibraryAudioRepository 音频文件仓库接口，用于媒体库音频文件的上传下载功能
type MediaLibraryAudioRepository interface {
	shared.BaseRepository[MediaLibraryAudio]
	
	// 上传下载相关方法
	GetByLibraryID(ctx context.Context, libraryID primitive.ObjectID) ([]*MediaLibraryAudio, error)
	GetByUploaderID(ctx context.Context, uploaderID primitive.ObjectID) ([]*MediaLibraryAudio, error)
	GetByStatus(ctx context.Context, status string) ([]*MediaLibraryAudio, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
	UpdateFilePath(ctx context.Context, id primitive.ObjectID, filePath string) error
	GetByChecksum(ctx context.Context, checksum string) (*MediaLibraryAudio, error)
	
	// 文件操作相关方法
	GetByFileName(ctx context.Context, fileName string) (*MediaLibraryAudio, error)
	IncrementDownloadCount(ctx context.Context, id primitive.ObjectID) error
}
