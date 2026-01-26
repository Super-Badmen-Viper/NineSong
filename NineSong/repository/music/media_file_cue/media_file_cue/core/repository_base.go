package core

import (
	"context"
	"fmt"
	"time"

	mediaFileCueCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file_cue/media_file_cue/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create 创建CUE文件
func (r *mediaFileCueRepository) Create(ctx context.Context, entity *mediaFileCueCore.MediaFileCueMetadata) error {
	entity.SetTimestamps()
	coll := r.db.Collection(r.collection)
	_, err := coll.InsertOne(ctx, entity)
	return err
}

// GetByID 根据ID获取CUE文件
func (r *mediaFileCueRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*mediaFileCueCore.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"_id": id})

	var file mediaFileCueCore.MediaFileCueMetadata
	if err := result.Decode(&file); err != nil {
		if shared.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("获取CUE文件失败: %w", err)
	}
	return &file, nil
}

// Upsert 实现BaseRepository接口（不返回结果）
// 注意：MediaFileCueRepository接口中的Upsert返回结果，这里实现BaseRepository的版本
// 由于Go不允许同名方法有不同签名，我们需要通过类型转换来处理
// 实际上，由于接口中已经定义了Upsert返回结果，BaseRepository的Upsert会被覆盖
// 但为了满足BaseRepository接口，我们需要提供一个不返回结果的版本
// 这里我们通过调用返回结果的版本并忽略结果来实现
func (r *mediaFileCueRepository) Upsert(ctx context.Context, entity *mediaFileCueCore.MediaFileCueMetadata) error {
	_, err := r.UpsertWithResult(ctx, entity)
	return err
}

// UpdateByID 已在repository.go中实现

// DeleteByID 根据ID删除CUE文件
func (r *mediaFileCueRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("删除CUE文件失败: %w", err)
	}
	return nil
}

// CreateMany 批量创建CUE文件
func (r *mediaFileCueRepository) CreateMany(ctx context.Context, entities []*mediaFileCueCore.MediaFileCueMetadata) error {
	var docs []interface{}
	for _, entity := range entities {
		entity.SetTimestamps()
		docs = append(docs, entity)
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := r.db.Collection(r.collection).InsertMany(ctx, docs)
	return err
}

// BulkUpsert 批量创建/更新CUE文件
func (r *mediaFileCueRepository) BulkUpsert(ctx context.Context, entities []*mediaFileCueCore.MediaFileCueMetadata) (int, error) {
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
			return successCount, fmt.Errorf("bulk upsert失败于索引%d: %w", successCount, err)
		}
		successCount++
	}
	return successCount, nil
}

// DeleteMany 批量删除CUE文件
func (r *mediaFileCueRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	coll := r.db.Collection(r.collection)
	return coll.DeleteMany(ctx, filter)
}

// GetAll 获取所有CUE文件
func (r *mediaFileCueRepository) GetAll(ctx context.Context) ([]*mediaFileCueCore.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var files []*mediaFileCueCore.MediaFileCueMetadata
	if err := cursor.All(ctx, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// GetByFilter 根据过滤条件获取CUE文件列表
func (r *mediaFileCueRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*mediaFileCueCore.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var files []*mediaFileCueCore.MediaFileCueMetadata
	if err := cursor.All(ctx, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// GetOneByFilter 根据过滤条件获取单个CUE文件
func (r *mediaFileCueRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*mediaFileCueCore.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	var file mediaFileCueCore.MediaFileCueMetadata
	err := coll.FindOne(ctx, filter).Decode(&file)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// Count 统计符合条件的CUE文件数量
func (r *mediaFileCueRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	coll := r.db.Collection(r.collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetPaginated 分页获取CUE文件
func (r *mediaFileCueRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*mediaFileCueCore.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var files []*mediaFileCueCore.MediaFileCueMetadata
	if err := cursor.All(ctx, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// UpdateMany 批量更新CUE文件
func (r *mediaFileCueRepository) UpdateMany(
	ctx context.Context,
	filter interface{},
	update interface{},
	opts ...*options.UpdateOptions,
) (*driver.UpdateResult, error) {
	coll := r.db.Collection(r.collection)
	return coll.UpdateMany(ctx, filter, update, opts...)
}

// Exists 检查CUE文件是否存在
func (r *mediaFileCueRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{"_id": id})
	return count > 0, err
}

// ExistsByFilter 根据过滤器检查CUE文件是否存在
func (r *mediaFileCueRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	return count > 0, err
}
