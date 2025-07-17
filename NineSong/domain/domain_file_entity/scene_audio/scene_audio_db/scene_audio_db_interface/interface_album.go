package scene_audio_db_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AlbumRepository 专辑领域层接口
type AlbumRepository interface {
	// 创建/更新
	Upsert(ctx context.Context, album *scene_audio_db_models.AlbumMetadata) error
	BulkUpsert(ctx context.Context, albums []*scene_audio_db_models.AlbumMetadata) (int, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// 删除
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	DeleteByName(ctx context.Context, name string) error
	DeleteAll(ctx context.Context) (int64, error)

	// 查询
	GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.AlbumMetadata, error)
	GetByName(ctx context.Context, name string) (*scene_audio_db_models.AlbumMetadata, error)
	GetByArtist(ctx context.Context, artistID string) ([]*scene_audio_db_models.AlbumMetadata, error)
	GetAllIDs(ctx context.Context) ([]primitive.ObjectID, error)

	GetAllCounts(ctx context.Context) ([]scene_audio_db_models.AlbumSongCounts, error)
	ResetALLField(ctx context.Context) (int64, error)
	ResetField(ctx context.Context, field string) (int64, error)
	UpdateCounter(ctx context.Context, albumID primitive.ObjectID, field string, increment int) (int64, error)

	GetByMbzID(ctx context.Context, mbzID string) (*scene_audio_db_models.AlbumMetadata, error)
	GetByFilter(ctx context.Context, filter interface{}) (*scene_audio_db_models.AlbumMetadata, error)

	GetArtistAlbumsMap(ctx context.Context) (map[primitive.ObjectID][]primitive.ObjectID, error)
	GetArtistGuestAlbumsMap(ctx context.Context) (map[primitive.ObjectID][]primitive.ObjectID, error)

	AlbumCountByArtist(ctx context.Context, artistID string) (int64, error)
	GuestAlbumCountByArtist(ctx context.Context, artistID string) (int64, error)

	InspectAlbumMediaCountByAlbum(ctx context.Context, albumID string, operand int) (bool, error)
}

// 查询参数结构
type AlbumFilterParams struct {
	MinYear       int
	MaxYear       int
	ArtistIDs     []string
	AlbumTypes    []string
	MBZAlbumTypes []string
	HasCoverArt   *bool
}

type AlbumSortParams struct {
	Field     string // 支持字段：name, artist, year 等
	Ascending bool
}
