package core

import (
	"context"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	wordCloudCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/word_cloud/word_cloud/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaFileRepository 媒体文件仓库接口
type MediaFileRepository interface {
	shared.BaseRepository[MediaFileMetadata] // 嵌入基础Repository接口

	// 词云相关
	GetAllGenre(ctx context.Context) ([]wordCloudCore.WordCloudMetadata, error)
	GetHighFrequencyWords(ctx context.Context, limit int) ([]wordCloudCore.WordCloudMetadata, error)
	GetRecommendedByKeywords(ctx context.Context, keywords []string, limit int) ([]wordCloudCore.WordCloudRecommendation, error)
	GetAllCounts(ctx context.Context) ([]MediaFileCounts, error)

	// 创建/更新
	// UpsertWithResult 返回结果的版本（BaseRepository的Upsert不返回结果）
	UpsertWithResult(ctx context.Context, file *MediaFileMetadata) (*MediaFileMetadata, error)
	BulkUpsert(ctx context.Context, files []*MediaFileMetadata) (int, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error)

	// 删除
	DeleteByPath(ctx context.Context, path string) error
	DeleteAllInvalid(ctx context.Context, filePaths []string) (int64, []struct {
		ArtistID primitive.ObjectID
		Count    int64
	}, error)
	DeleteByFolder(ctx context.Context, folderPath string) (int64, error)
	DeleteAll(ctx context.Context) (int64, error)

	// 查询
	GetByPath(ctx context.Context, path string) (*MediaFileMetadata, error)
	GetByFolder(ctx context.Context, folderPath string) ([]string, error)
	GetFilesWithMissingMetadata(ctx context.Context, folderPath string, folderType int) ([]*MediaFileMetadata, error)

	MediaCountByArtist(ctx context.Context, artistID string) (int64, error)
	GuestMediaCountByArtist(ctx context.Context, artistID string) (int64, error)
	MediaCountByAlbum(ctx context.Context, albumID string) (int64, error)

	InspectMediaCountByArtist(ctx context.Context, artistID string, filePaths []string) (int, error)
	InspectGuestMediaCountByArtist(ctx context.Context, artistID string, filePaths []string) (int, error)
	InspectMediaCountByAlbum(ctx context.Context, albumID string, filePaths []string) (int, error)
	InspectGuestMediaCountByAlbum(ctx context.Context, artistID string, filePaths []string) (int, error)
}
