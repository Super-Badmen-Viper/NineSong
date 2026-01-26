package core

import (
	"context"
	"fmt"

	playlistTrackCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/playlist/playlist_track/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type playlistTrackRepository struct {
	db         mongo.Database
	collection string
}

func NewPlaylistTrackRepository(db mongo.Database, collection string) playlistTrackCore.PlaylistTrackRepository {
	return &playlistTrackRepository{
		db:         db,
		collection: collection,
	}
}

// GetByPlaylistID 根据播放列表ID获取曲目列表
func (r *playlistTrackRepository) GetByPlaylistID(ctx context.Context, playlistID primitive.ObjectID) ([]*playlistTrackCore.PlaylistTrackMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, bson.M{"playlist_id": playlistID})
	if err != nil {
		return nil, fmt.Errorf("get playlist tracks by playlist ID failed: %w", err)
	}
	defer cursor.Close(ctx)

	var tracks []*playlistTrackCore.PlaylistTrackMetadata
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, fmt.Errorf("decode playlist tracks failed: %w", err)
	}
	return tracks, nil
}

// GetByMediaFileID 根据媒体文件ID获取曲目列表
func (r *playlistTrackRepository) GetByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) ([]*playlistTrackCore.PlaylistTrackMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, bson.M{"media_file_id": mediaFileID})
	if err != nil {
		return nil, fmt.Errorf("get playlist tracks by media file ID failed: %w", err)
	}
	defer cursor.Close(ctx)

	var tracks []*playlistTrackCore.PlaylistTrackMetadata
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, fmt.Errorf("decode playlist tracks failed: %w", err)
	}
	return tracks, nil
}

// DeleteByPlaylistID 根据播放列表ID删除曲目
func (r *playlistTrackRepository) DeleteByPlaylistID(ctx context.Context, playlistID primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteMany(ctx, bson.M{"playlist_id": playlistID})
	if err != nil {
		return fmt.Errorf("delete playlist tracks by playlist ID failed: %w", err)
	}
	return nil
}

// DeleteByMediaFileID 根据媒体文件ID删除曲目
func (r *playlistTrackRepository) DeleteByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteMany(ctx, bson.M{"media_file_id": mediaFileID})
	if err != nil {
		return fmt.Errorf("delete playlist tracks by media file ID failed: %w", err)
	}
	return nil
}
