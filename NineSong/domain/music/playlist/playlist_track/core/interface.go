package core

import (
	"context"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlaylistTrackRepository 播放列表曲目仓库接口
type PlaylistTrackRepository interface {
	shared.BaseRepository[PlaylistTrackMetadata] // 嵌入基础Repository接口

	// 特定方法
	GetByPlaylistID(ctx context.Context, playlistID primitive.ObjectID) ([]*PlaylistTrackMetadata, error)
	GetByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) ([]*PlaylistTrackMetadata, error)
	DeleteByPlaylistID(ctx context.Context, playlistID primitive.ObjectID) error
	DeleteByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) error
}
