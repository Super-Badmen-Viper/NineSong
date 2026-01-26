package core

import (
	"context"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaFileCueRepository 媒体文件CUE仓库接口
type MediaFileCueRepository interface {
	shared.BaseRepository[MediaFileCueMetadata] // 嵌入基础Repository接口

	// 特定方法（Upsert返回结果版本）
	UpsertWithResult(ctx context.Context, file *MediaFileCueMetadata) (*MediaFileCueMetadata, error)
	GetByPath(ctx context.Context, path string) (*MediaFileCueMetadata, error)
	DeleteByPath(ctx context.Context, path string) error
	DeleteAllInvalid(ctx context.Context, filePaths []string) (int64, []struct {
		ArtistID primitive.ObjectID
		Count    int64
	}, error)
	DeleteByFolder(ctx context.Context, folderPath string) (int64, error)
	DeleteAll(ctx context.Context) (int64, error)
	GetByFolder(ctx context.Context, folderPath string) ([]string, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error)

	MediaCueCountByArtist(ctx context.Context, artistID string) (int64, error)
	GuestMediaCueCountByArtist(ctx context.Context, artistID string) (int64, error)
	MediaCountByAlbum(ctx context.Context, albumID string) (int64, error)

	InspectMediaCueCountByArtist(ctx context.Context, artistID string, filePaths []string) (int, error)
	InspectGuestMediaCueCountByArtist(ctx context.Context, artistID string, filePaths []string) (int, error)
}
