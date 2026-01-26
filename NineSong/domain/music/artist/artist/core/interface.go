package core

import (
	"context"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ArtistRepository 艺术家领域层接口
type ArtistRepository interface {
	shared.BaseRepository[ArtistMetadata] // 嵌入基础Repository接口

	// 创建/更新
	Upsert(ctx context.Context, artist *ArtistMetadata) error
	BulkUpsert(ctx context.Context, artists []*ArtistMetadata) (int, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// 删除
	DeleteByName(ctx context.Context, name string) error
	DeleteAllInvalid(ctx context.Context) (int64, error)
	DeleteAll(ctx context.Context) (int64, error)

	// 艺术家-专辑关联管理
	UpdateArtistAlbums(ctx context.Context, artistID primitive.ObjectID, albumIDs []ArtistIDPair) error
	UpdateArtistGuestAlbums(ctx context.Context, artistID primitive.ObjectID, albumIDs []ArtistIDPair) error

	// 查询
	GetByName(ctx context.Context, name string) (*ArtistMetadata, error)
	GetAllIDs(ctx context.Context) ([]primitive.ObjectID, error)

	GetAllCounts(ctx context.Context) ([]ArtistAlbumAndSongCounts, error)
	ResetALLField(ctx context.Context) (int64, error)
	ResetField(ctx context.Context, field string) (int64, error)
	UpdateCounter(ctx context.Context, artistID primitive.ObjectID, field string, increment int) (int64, error)
	SetCounter(ctx context.Context, artistID primitive.ObjectID, field string, value int) error

	GetByMbzID(ctx context.Context, mbzID string) (*ArtistMetadata, error)

	InspectAlbumCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectGuestAlbumCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectMediaCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectGuestMediaCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectMediaCueCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
	InspectGuestMediaCueCountByArtist(ctx context.Context, artistID string, operand int) (int, error)
}
