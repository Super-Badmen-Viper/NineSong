package core

import (
	"context"
	"fmt"
	"time"

	annotationCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/annotation/annotation/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create 创建注释
func (r *annotationRepository) Create(ctx context.Context, entity *annotationCore.AnnotationMetadata) error {
	entity.SetTimestamps()
	coll := r.db.Collection(r.collection)
	_, err := coll.InsertOne(ctx, entity)
	return err
}

// GetByID 根据ID获取注释
func (r *annotationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*annotationCore.AnnotationMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"_id": id})

	var annotation annotationCore.AnnotationMetadata
	if err := result.Decode(&annotation); err != nil {
		if shared.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get annotation by ID failed: %w", err)
	}
	return &annotation, nil
}

// Upsert 插入或更新注释
func (r *annotationRepository) Upsert(ctx context.Context, entity *annotationCore.AnnotationMetadata) error {
	coll := r.db.Collection(r.collection)
	now := time.Now()

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

	// 强制设置更新时间
	update["$set"].(bson.M)["updated_at"] = now

	// 执行更新操作
	filter := bson.M{"_id": entity.ID}
	opts := options.Update().SetUpsert(true)
	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("annotation upsert failed: %w", err)
	}

	// 处理插入后的ID同步
	if result.UpsertedID != nil {
		if oid, ok := result.UpsertedID.(primitive.ObjectID); ok {
			entity.ID = oid
		}
	}
	entity.UpdatedAt = now

	return nil
}

// UpdateByID 根据ID更新注释
func (r *annotationRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	coll := r.db.Collection(r.collection)

	now := time.Now()
	setUpdate := bson.M{}

	hasOperator := false
	for key := range update {
		if key[0] == '$' {
			hasOperator = true
			break
		}
	}

	if hasOperator {
		for op, opValue := range update {
			if op == "$set" {
				if setValues, ok := opValue.(bson.M); ok {
					setValues["updated_at"] = now
					setUpdate[op] = setValues
				} else {
					setUpdate[op] = opValue
				}
			} else {
				setUpdate[op] = opValue
			}
		}
		if _, exists := setUpdate["$set"]; !exists {
			setUpdate["$set"] = bson.M{"updated_at": now}
		}
	} else {
		for k, v := range update {
			setUpdate[k] = v
		}
		setUpdate["updated_at"] = now
	}

	finalUpdate := bson.M{"$set": setUpdate}

	result, err := coll.UpdateOne(ctx, bson.M{"_id": id}, finalUpdate)
	if err != nil {
		return false, err
	}

	return result.MatchedCount > 0, nil
}

// DeleteByID 根据ID删除注释
func (r *annotationRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete annotation by ID failed: %w", err)
	}
	return nil
}

// CreateMany 批量创建注释
func (r *annotationRepository) CreateMany(ctx context.Context, entities []*annotationCore.AnnotationMetadata) error {
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

// BulkUpsert 批量创建/更新注释
func (r *annotationRepository) BulkUpsert(ctx context.Context, entities []*annotationCore.AnnotationMetadata) (int, error) {
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

// DeleteMany 批量删除注释
func (r *annotationRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	coll := r.db.Collection(r.collection)
	return coll.DeleteMany(ctx, filter)
}

// GetAll 获取所有注释
func (r *annotationRepository) GetAll(ctx context.Context) ([]*annotationCore.AnnotationMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var annotations []*annotationCore.AnnotationMetadata
	if err := cursor.All(ctx, &annotations); err != nil {
		return nil, err
	}
	return annotations, nil
}

// GetByFilter 根据过滤条件获取注释列表
func (r *annotationRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*annotationCore.AnnotationMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var annotations []*annotationCore.AnnotationMetadata
	if err := cursor.All(ctx, &annotations); err != nil {
		return nil, err
	}
	return annotations, nil
}

// GetOneByFilter 根据过滤条件获取单个注释
func (r *annotationRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*annotationCore.AnnotationMetadata, error) {
	coll := r.db.Collection(r.collection)
	var annotation annotationCore.AnnotationMetadata
	err := coll.FindOne(ctx, filter).Decode(&annotation)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &annotation, nil
}

// Count 统计符合条件的注释数量
func (r *annotationRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	coll := r.db.Collection(r.collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetPaginated 分页获取注释
func (r *annotationRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*annotationCore.AnnotationMetadata, error) {
	coll := r.db.Collection(r.collection)
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var annotations []*annotationCore.AnnotationMetadata
	if err := cursor.All(ctx, &annotations); err != nil {
		return nil, err
	}
	return annotations, nil
}

// UpdateMany 批量更新注释
func (r *annotationRepository) UpdateMany(
	ctx context.Context,
	filter interface{},
	update interface{},
	opts ...*options.UpdateOptions,
) (*driver.UpdateResult, error) {
	coll := r.db.Collection(r.collection)
	return coll.UpdateMany(ctx, filter, update, opts...)
}

// Exists 检查注释是否存在
func (r *annotationRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{"_id": id})
	return count > 0, err
}

// ExistsByFilter 根据过滤器检查注释是否存在
func (r *annotationRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	return count > 0, err
}
