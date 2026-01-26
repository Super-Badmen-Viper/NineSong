package core

import (
	"context"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AlbumRepository 专辑领域层接口
type AlbumRepository interface {
	shared.BaseRepository[AlbumMetadata] // 嵌入基础Repository接口

	// 创建/更新
	Upsert(ctx context.Context, album *AlbumMetadata) error
	BulkUpsert(ctx context.Context, albums []*AlbumMetadata) (int, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// 删除
	DeleteByName(ctx context.Context, name string) error
	DeleteAll(ctx context.Context) (int64, error)

	// 查询
	GetByName(ctx context.Context, name string) (*AlbumMetadata, error)
	GetByArtist(ctx context.Context, artistID string) ([]*AlbumMetadata, error)
	GetAllIDs(ctx context.Context) ([]primitive.ObjectID, error)

	GetAllCounts(ctx context.Context) ([]AlbumSongCounts, error)
	ResetALLField(ctx context.Context) (int64, error)
	ResetField(ctx context.Context, field string) (int64, error)
	UpdateCounter(ctx context.Context, albumID primitive.ObjectID, field string, increment int) (int64, error)

	GetByMbzID(ctx context.Context, mbzID string) (*AlbumMetadata, error)
	// GetByFilterSingle 根据过滤器获取单个专辑（与BaseRepository的GetByFilter区分，BaseRepository返回列表）
	GetByFilterSingle(ctx context.Context, filter interface{}) (*AlbumMetadata, error)

	GetArtistAlbumsMap(ctx context.Context) (map[primitive.ObjectID][]primitive.ObjectID, error)
	GetArtistGuestAlbumsMap(ctx context.Context) (map[primitive.ObjectID][]primitive.ObjectID, error)

	AlbumCountByArtist(ctx context.Context, artistID string) (int64, error)
	GuestAlbumCountByArtist(ctx context.Context, artistID string) (int64, error)

	InspectAlbumMediaCountByAlbum(ctx context.Context, albumID string, operand int) (bool, error)
}
