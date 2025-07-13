package scene_audio_db_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaFileRepository 基础CRUD接口
type MediaFileRepository interface {
	GetHighFrequencyWords(ctx context.Context, limit int) ([]scene_audio_db_models.WordCloudMetadata, error)
	GetRecommendedByKeywords(ctx context.Context, keywords []string, limit int) ([]scene_audio_db_models.Recommendation, error)

	// 创建/更新
	Upsert(ctx context.Context, file *scene_audio_db_models.MediaFileMetadata) (*scene_audio_db_models.MediaFileMetadata, error)
	BulkUpsert(ctx context.Context, files []*scene_audio_db_models.MediaFileMetadata) (int, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error)

	// 删除
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	DeleteByPath(ctx context.Context, path string) error
	DeleteAllInvalid(ctx context.Context, filePaths []string) (int64, []struct {
		ArtistID primitive.ObjectID
		Count    int64
	}, error)
	DeleteByFolder(ctx context.Context, folderPath string) (int64, error)

	// 查询
	GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.MediaFileMetadata, error)
	GetByPath(ctx context.Context, path string) (*scene_audio_db_models.MediaFileMetadata, error)
	GetByFolder(ctx context.Context, folderPath string) ([]string, error)

	MediaCountByArtist(ctx context.Context, artistID string) (int64, error)
	GuestMediaCountByArtist(ctx context.Context, artistID string) (int64, error)
	MediaCountByAlbum(ctx context.Context, albumID string) (int64, error)

	InspectMediaCountByArtist(ctx context.Context, artistID string, filePaths []string) (int, error)
	InspectGuestMediaCountByArtist(ctx context.Context, artistID string, filePaths []string) (int, error)
	InspectMediaCountByAlbum(ctx context.Context, albumID string, filePaths []string) (int, error)
	InspectGuestMediaCountByAlbum(ctx context.Context, artistID string, filePaths []string) (int, error)
}
