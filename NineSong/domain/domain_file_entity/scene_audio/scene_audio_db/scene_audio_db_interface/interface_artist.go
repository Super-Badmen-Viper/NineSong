package scene_audio_db_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ArtistRepository 艺术家领域层接口
type ArtistRepository interface {
	// 创建/更新
	Upsert(ctx context.Context, artist *scene_audio_db_models.ArtistMetadata) error
	BulkUpsert(ctx context.Context, artists []*scene_audio_db_models.ArtistMetadata) (int, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// 删除
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	DeleteByName(ctx context.Context, name string) error
	DeleteAllInvalid(ctx context.Context) (int64, error)
	DeleteAll(ctx context.Context) (int64, error)

	// 查询
	GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.ArtistMetadata, error)
	GetByName(ctx context.Context, name string) (*scene_audio_db_models.ArtistMetadata, error)
	GetAllIDs(ctx context.Context) ([]primitive.ObjectID, error)

	GetAllCounts(ctx context.Context) ([]scene_audio_db_models.ArtistAlbumAndSongCounts, error)
	ResetALLField(ctx context.Context) (int64, error)
	ResetField(ctx context.Context, field string) (int64, error)
	UpdateCounter(ctx context.Context, artistID primitive.ObjectID, field string, increment int) (int64, error)

	GetByMbzID(ctx context.Context, mbzID string) (*scene_audio_db_models.ArtistMetadata, error)

	InspectAlbumCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectGuestAlbumCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectMediaCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectGuestMediaCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectMediaCueCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectGuestMediaCueCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
}
