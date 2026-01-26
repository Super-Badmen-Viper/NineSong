package core

import (
	"context"
	"time"

	mediaFileCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file/media_file/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create 创建媒体文件
func (r *mediaFileRepository) Create(ctx context.Context, entity *mediaFileCore.MediaFileMetadata) error {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	coll := r.db.Collection(r.collection)
	_, err := coll.InsertOne(ctx, entity)
	return err
}

// CreateMany 批量创建媒体文件
func (r *mediaFileRepository) CreateMany(ctx context.Context, entities []*mediaFileCore.MediaFileMetadata) error {
	var docs []interface{}
	for _, entity := range entities {
		entity.CreatedAt = time.Now()
		entity.UpdatedAt = time.Now()
		docs = append(docs, entity)
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := r.db.Collection(r.collection).InsertMany(ctx, docs)
	return err
}

// DeleteMany 批量删除媒体文件
func (r *mediaFileRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	result, err := r.db.Collection(r.collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// GetByFilter 根据过滤器获取媒体文件列表
func (r *mediaFileRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*mediaFileCore.MediaFileMetadata, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*mediaFileCore.MediaFileMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetOneByFilter 根据过滤器获取单个媒体文件
func (r *mediaFileRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*mediaFileCore.MediaFileMetadata, error) {
	var result mediaFileCore.MediaFileMetadata
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

// Count 统计媒体文件数量
func (r *mediaFileRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	return r.db.Collection(r.collection).CountDocuments(ctx, filter)
}

// GetPaginated 分页获取媒体文件
func (r *mediaFileRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*mediaFileCore.MediaFileMetadata, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.db.Collection(r.collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*mediaFileCore.MediaFileMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// UpdateMany 批量更新媒体文件
func (r *mediaFileRepository) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*driver.UpdateResult, error) {
	return r.db.Collection(r.collection).UpdateMany(ctx, filter, update, opts...)
}

// Exists 检查媒体文件是否存在
func (r *mediaFileRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{"_id": id})
	return count > 0, err
}

// ExistsByFilter 根据过滤器检查媒体文件是否存在
func (r *mediaFileRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	return count > 0, err
}

// Upsert 实现BaseRepository接口（不返回结果）
// 注意：由于MediaFileRepository接口中的Upsert返回结果，而BaseRepository要求不返回结果，
// 这里通过调用返回结果的Upsert方法来实现BaseRepository的版本
func (r *mediaFileRepository) UpsertBase(ctx context.Context, entity *mediaFileCore.MediaFileMetadata) error {
	_, err := r.Upsert(ctx, entity)
	return err
}
