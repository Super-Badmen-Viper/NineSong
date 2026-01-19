package scene_audio_db_interface

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
)

// ChunkedUploadSessionRepository 分块上传会话仓库接口
type ChunkedUploadSessionRepository interface {
	domain.BaseRepository[scene_audio_db_models.ChunkedUploadSession]

	// 通过上传ID获取会话
	GetByUploadID(ctx context.Context, uploadID string) (*scene_audio_db_models.ChunkedUploadSession, error)

	// 更新会话状态
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error

	// 更新已上传分块数
	UpdateUploadedChunks(ctx context.Context, id primitive.ObjectID, uploadedChunks int) error

	// 更新文件路径
	UpdateFilePath(ctx context.Context, id primitive.ObjectID, filePath string) error

	// 更新校验和
	UpdateChecksum(ctx context.Context, id primitive.ObjectID, checksum string) error

	// 更新进度
	UpdateProgress(ctx context.Context, id primitive.ObjectID, progress float64) error

	// 更新已上传字节数
	UpdateUploadedBytes(ctx context.Context, id primitive.ObjectID, uploadedBytes int64) error
}