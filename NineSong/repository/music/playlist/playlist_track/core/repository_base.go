package core

import (
	"context"
	"fmt"

	playlistTrackCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/playlist/playlist_track/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create 创建播放列表曲目
func (r *playlistTrackRepository) Create(ctx context.Context, entity *playlistTrackCore.PlaylistTrackMetadata) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.InsertOne(ctx, entity)
	return err
}

// GetByID 根据ID获取播放列表曲目
func (r *playlistTrackRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*playlistTrackCore.PlaylistTrackMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"_id": id})

	var track playlistTrackCore.PlaylistTrackMetadata
	if err := result.Decode(&track); err != nil {
		if shared.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get playlist track by ID failed: %w", err)
	}
	return &track, nil
}

// Upsert 插入或更新播放列表曲目
func (r *playlistTrackRepository) Upsert(ctx context.Context, entity *playlistTrackCore.PlaylistTrackMetadata) error {
	coll := r.db.Collection(r.collection)

	// 序列化结构体为BSON
	var doc bson.M
	data, err := bson.Marshal(entity)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	if err := bson.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	// 构造更新文档
	update := bson.M{
		"$set":         make(bson.M),
		"$setOnInsert": bson.M{},
	}

	// 动态构建$set内容，排除特定字段
	for k, v := range doc {
		if k == "_id" {
			continue // 跳过ID
		}
		update["$set"].(bson.M)[k] = v
	}

	// 执行更新操作
	filter := bson.M{"_id": entity.ID}
	opts := options.Update().SetUpsert(true)
	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("playlist track upsert failed: %w", err)
	}

	// 处理插入后的ID同步
	if result.UpsertedID != nil {
		if oid, ok := result.UpsertedID.(primitive.ObjectID); ok {
			entity.ID = oid
		}
	}

	return nil
}

// UpdateByID 根据ID更新播放列表曲目
func (r *playlistTrackRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	coll := r.db.Collection(r.collection)

	setUpdate := bson.M{}

	hasOperator := false
	for key := range update {
		if key[0] == '$' {
			hasOperator = true
			break
		}
	}

	if hasOperator {
		setUpdate = update
	} else {
		for k, v := range update {
			setUpdate[k] = v
		}
	}

	finalUpdate := bson.M{"$set": setUpdate}

	result, err := coll.UpdateOne(ctx, bson.M{"_id": id}, finalUpdate)
	if err != nil {
		return false, err
	}

	return result.MatchedCount > 0, nil
}

// DeleteByID 根据ID删除播放列表曲目
func (r *playlistTrackRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete playlist track by ID failed: %w", err)
	}
	return nil
}

// CreateMany 批量创建播放列表曲目
func (r *playlistTrackRepository) CreateMany(ctx context.Context, entities []*playlistTrackCore.PlaylistTrackMetadata) error {
	var docs []interface{}
	for _, entity := range entities {
		docs = append(docs, entity)
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := r.db.Collection(r.collection).InsertMany(ctx, docs)
	return err
}

// BulkUpsert 批量创建/更新播放列表曲目
func (r *playlistTrackRepository) BulkUpsert(ctx context.Context, entities []*playlistTrackCore.PlaylistTrackMetadata) (int, error) {
	coll := r.db.Collection(r.collection)
	var successCount int

	for _, entity := range entities {
		filter := bson.M{"_id": entity.ID}
		update := bson.M{"$set": entity}

		_, err := coll.UpdateOne(
			ctx,
			filter,
			update,
			options.Update().SetUpsert(true),
		)

		if err != nil {
			return successCount, fmt.Errorf("bulk upsert failed at index %d: %w", successCount, err)
		}
		successCount++
	}
	return successCount, nil
}

// DeleteMany 批量删除播放列表曲目
func (r *playlistTrackRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	coll := r.db.Collection(r.collection)
	return coll.DeleteMany(ctx, filter)
}

// GetAll 获取所有播放列表曲目
func (r *playlistTrackRepository) GetAll(ctx context.Context) ([]*playlistTrackCore.PlaylistTrackMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tracks []*playlistTrackCore.PlaylistTrackMetadata
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, err
	}
	return tracks, nil
}

// GetByFilter 根据过滤条件获取播放列表曲目列表
func (r *playlistTrackRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*playlistTrackCore.PlaylistTrackMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tracks []*playlistTrackCore.PlaylistTrackMetadata
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, err
	}
	return tracks, nil
}

// GetOneByFilter 根据过滤条件获取单个播放列表曲目
func (r *playlistTrackRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*playlistTrackCore.PlaylistTrackMetadata, error) {
	coll := r.db.Collection(r.collection)
	var track playlistTrackCore.PlaylistTrackMetadata
	err := coll.FindOne(ctx, filter).Decode(&track)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &track, nil
}

// Count 统计符合条件的播放列表曲目数量
func (r *playlistTrackRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	coll := r.db.Collection(r.collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetPaginated 分页获取播放列表曲目
func (r *playlistTrackRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*playlistTrackCore.PlaylistTrackMetadata, error) {
	coll := r.db.Collection(r.collection)
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "index", Value: 1}})
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tracks []*playlistTrackCore.PlaylistTrackMetadata
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, err
	}
	return tracks, nil
}

// UpdateMany 批量更新播放列表曲目
func (r *playlistTrackRepository) UpdateMany(
	ctx context.Context,
	filter interface{},
	update interface{},
	opts ...*options.UpdateOptions,
) (*driver.UpdateResult, error) {
	coll := r.db.Collection(r.collection)
	return coll.UpdateMany(ctx, filter, update, opts...)
}

// Exists 检查播放列表曲目是否存在
func (r *playlistTrackRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{"_id": id})
	return count > 0, err
}

// ExistsByFilter 根据过滤器检查播放列表曲目是否存在
func (r *playlistTrackRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	return count > 0, err
}
