package core

import (
	"context"
	"fmt"
	"strconv"

	albumCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/album/album/core"
	artistCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/artist/artist/core"
	homeCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/home/home/core"
	mediaFileCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file/media_file/core"
	mediaFileCueCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file_cue/media_file_cue/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type homeRepository struct {
	db mongo.Database
}

func NewHomeRepository(db mongo.Database) homeCore.HomeRepository {
	return &homeRepository{db: db}
}

// GetRandomArtistList 获取随机艺术家列表
func (r *homeRepository) GetRandomArtistList(
	ctx context.Context,
	start, end string,
) ([]*artistCore.ArtistMetadata, error) {
	collection := r.db.Collection(shared.CollectionFileEntityAudioSceneArtist)

	// 转换分页参数
	skip, _ := strconv.Atoi(start)
	limit, _ := strconv.Atoi(end)
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	// 构建随机查询
	pipeline := []bson.M{
		{"$sample": bson.M{"size": limit + skip}},
		{"$skip": skip},
		{"$limit": limit},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("random query failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*artistCore.ArtistMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

// GetRandomAlbumList 获取随机专辑列表
func (r *homeRepository) GetRandomAlbumList(
	ctx context.Context,
	start, end string,
) ([]*albumCore.AlbumMetadata, error) {
	collection := r.db.Collection(shared.CollectionFileEntityAudioSceneAlbum)

	skip, _ := strconv.Atoi(start)
	limit, _ := strconv.Atoi(end)
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	pipeline := []bson.M{
		{"$sample": bson.M{"size": limit + skip}},
		{"$skip": skip},
		{"$limit": limit},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("random query failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*albumCore.AlbumMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

// GetRandomMediaFileList 获取随机媒体文件列表
func (r *homeRepository) GetRandomMediaFileList(
	ctx context.Context,
	start, end string,
) ([]*mediaFileCore.MediaFileMetadata, error) {
	collection := r.db.Collection(shared.CollectionFileEntityAudioSceneMediaFile)

	skip, _ := strconv.Atoi(start)
	limit, _ := strconv.Atoi(end)
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	pipeline := []bson.M{
		{"$sample": bson.M{"size": limit + skip}},
		{"$skip": skip},
		{"$limit": limit},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("random query failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*mediaFileCore.MediaFileMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

// GetRandomMediaCueList 获取随机CUE文件列表
func (r *homeRepository) GetRandomMediaCueList(
	ctx context.Context,
	start, end string,
) ([]*mediaFileCueCore.MediaFileCueMetadata, error) {
	collection := r.db.Collection(shared.CollectionFileEntityAudioSceneMediaFileCue)

	skip, _ := strconv.Atoi(start)
	limit, _ := strconv.Atoi(end)
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	pipeline := []bson.M{
		{"$sample": bson.M{"size": limit + skip}},
		{"$skip": skip},
		{"$limit": limit},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("random query failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*mediaFileCueCore.MediaFileCueMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}
