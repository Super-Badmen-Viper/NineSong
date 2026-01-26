package core

import (
	"context"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlaylistRepository 播放列表仓库接口
type PlaylistRepository interface {
	shared.BaseRepository[PlaylistMetadata] // 嵌入基础Repository接口

	// 特定方法
	GetByPath(ctx context.Context, path string) (*PlaylistMetadata, error)
	DeleteByPath(ctx context.Context, path string) error
}
