package core

import (
	"context"

	domainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/sync_record/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Upsert 插入或更新实体
func (r *mediaLibrarySyncRecordRepository) Upsert(ctx context.Context, entity *domainCore.MediaLibrarySyncRecord) error {
	entity.SetTimestamps()
	filter := bson.M{"_id": entity.ID}
	update := bson.M{"$set": entity}
	opts := options.Update().SetUpsert(true)
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update, opts)
	return err
}

// CreateMany 批量创建实体
func (r *mediaLibrarySyncRecordRepository) CreateMany(ctx context.Context, entities []*domainCore.MediaLibrarySyncRecord) error {
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

// BulkUpsert 批量创建/更新实体
func (r *mediaLibrarySyncRecordRepository) BulkUpsert(ctx context.Context, entities []*domainCore.MediaLibrarySyncRecord) (int, error) {
	coll := r.db.Collection(r.collection)
	var successCount int

	for _, entity := range entities {
		entity.SetTimestamps()
		filter := bson.M{"_id": entity.ID}
		update := bson.M{"$set": entity}
		opts := options.Update().SetUpsert(true)

		_, err := coll.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			continue
		}
		successCount++
	}

	return successCount, nil
}

// DeleteMany 批量删除实体
func (r *mediaLibrarySyncRecordRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	result, err := r.db.Collection(r.collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// GetAll 获取所有实体
func (r *mediaLibrarySyncRecordRepository) GetAll(ctx context.Context) ([]*domainCore.MediaLibrarySyncRecord, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*domainCore.MediaLibrarySyncRecord
	for cursor.Next(ctx) {
		var entity domainCore.MediaLibrarySyncRecord
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// GetByFilter 根据过滤条件获取实体列表
func (r *mediaLibrarySyncRecordRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*domainCore.MediaLibrarySyncRecord, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*domainCore.MediaLibrarySyncRecord
	for cursor.Next(ctx) {
		var entity domainCore.MediaLibrarySyncRecord
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// GetOneByFilter 根据过滤条件获取单个实体
func (r *mediaLibrarySyncRecordRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*domainCore.MediaLibrarySyncRecord, error) {
	var entity domainCore.MediaLibrarySyncRecord
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Count 统计符合条件的实体数量
func (r *mediaLibrarySyncRecordRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetPaginated 分页获取实体
func (r *mediaLibrarySyncRecordRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*domainCore.MediaLibrarySyncRecord, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.db.Collection(r.collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*domainCore.MediaLibrarySyncRecord
	for cursor.Next(ctx) {
		var entity domainCore.MediaLibrarySyncRecord
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// UpdateMany 批量更新实体
func (r *mediaLibrarySyncRecordRepository) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*driver.UpdateResult, error) {
	return r.db.Collection(r.collection).UpdateMany(ctx, filter, update, opts...)
}

// Exists 检查实体是否存在
func (r *mediaLibrarySyncRecordRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": id}
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByFilter 根据过滤条件检查实体是否存在
func (r *mediaLibrarySyncRecordRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
