package core

import (
	"context"
	"time"

	albumCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/album/album/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create 创建专辑
func (r *albumRepository) Create(ctx context.Context, entity *albumCore.AlbumMetadata) error {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	coll := r.db.Collection(r.collection)
	_, err := coll.InsertOne(ctx, entity)
	return err
}

// CreateMany 批量创建专辑
func (r *albumRepository) CreateMany(ctx context.Context, entities []*albumCore.AlbumMetadata) error {
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

// DeleteMany 批量删除专辑
func (r *albumRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	result, err := r.db.Collection(r.collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// GetAll 获取所有专辑
func (r *albumRepository) GetAll(ctx context.Context) ([]*albumCore.AlbumMetadata, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*albumCore.AlbumMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetByFilter 根据过滤器获取专辑列表
func (r *albumRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*albumCore.AlbumMetadata, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*albumCore.AlbumMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetOneByFilter 根据过滤器获取单个专辑
func (r *albumRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*albumCore.AlbumMetadata, error) {
	var result albumCore.AlbumMetadata
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

// Count 统计专辑数量
func (r *albumRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	return r.db.Collection(r.collection).CountDocuments(ctx, filter)
}

// GetPaginated 分页获取专辑
func (r *albumRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*albumCore.AlbumMetadata, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.db.Collection(r.collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*albumCore.AlbumMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Exists 检查专辑是否存在
func (r *albumRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{"_id": id})
	return count > 0, err
}

// ExistsByFilter 根据过滤器检查专辑是否存在
func (r *albumRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	return count > 0, err
}
