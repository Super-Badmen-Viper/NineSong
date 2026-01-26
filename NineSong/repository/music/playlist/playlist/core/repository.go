package core

import (
	"context"
	"fmt"
	"time"

	playlistCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/playlist/playlist/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type playlistRepository struct {
	db         mongo.Database
	collection string
}

func NewPlaylistRepository(db mongo.Database, collection string) playlistCore.PlaylistRepository {
	return &playlistRepository{
		db:         db,
		collection: collection,
	}
}

// GetByPath 根据路径获取播放列表
func (r *playlistRepository) GetByPath(ctx context.Context, path string) (*playlistCore.PlaylistMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"path": path})

	var playlist playlistCore.PlaylistMetadata
	if err := result.Decode(&playlist); err != nil {
		if shared.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get playlist by path failed: %w", err)
	}
	return &playlist, nil
}

// DeleteByPath 根据路径删除播放列表
func (r *playlistRepository) DeleteByPath(ctx context.Context, path string) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"path": path})
	if err != nil {
		return fmt.Errorf("delete playlist by path failed: %w", err)
	}
	return nil
}
