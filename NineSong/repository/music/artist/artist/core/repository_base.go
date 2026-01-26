package core

import (
	"context"
	"time"

	artistCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/artist/artist/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create 创建艺术家
func (r *artistRepository) Create(ctx context.Context, entity *artistCore.ArtistMetadata) error {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	coll := r.db.Collection(r.collection)
	_, err := coll.InsertOne(ctx, entity)
	return err
}

// CreateMany 批量创建艺术家
func (r *artistRepository) CreateMany(ctx context.Context, entities []*artistCore.ArtistMetadata) error {
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

// DeleteMany 批量删除艺术家
func (r *artistRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	result, err := r.db.Collection(r.collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// GetAll 获取所有艺术家
func (r *artistRepository) GetAll(ctx context.Context) ([]*artistCore.ArtistMetadata, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*artistCore.ArtistMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetByFilter 根据过滤器获取艺术家列表
func (r *artistRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*artistCore.ArtistMetadata, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*artistCore.ArtistMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetOneByFilter 根据过滤器获取单个艺术家
func (r *artistRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*artistCore.ArtistMetadata, error) {
	var result artistCore.ArtistMetadata
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

// Count 统计艺术家数量
func (r *artistRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	return r.db.Collection(r.collection).CountDocuments(ctx, filter)
}

// GetPaginated 分页获取艺术家
func (r *artistRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*artistCore.ArtistMetadata, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.db.Collection(r.collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*artistCore.ArtistMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Exists 检查艺术家是否存在
func (r *artistRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{"_id": id})
	return count > 0, err
}

// ExistsByFilter 根据过滤器检查艺术家是否存在
func (r *artistRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	return count > 0, err
}
