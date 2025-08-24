package repository

import (
	"context"
	"errors"
	"fmt"
	driver "go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BaseMongoRepository MongoDB通用Repository实现
type BaseMongoRepository[T any] struct {
	db         mongo.Database
	collection string
}

// NewBaseMongoRepository 创建新的MongoDB Repository实例
func NewBaseMongoRepository[T any](db mongo.Database, collection string) domain.BaseRepository[T] {
	return &BaseMongoRepository[T]{
		db:         db,
		collection: collection,
	}
}

// Create 创建新实体
func (r *BaseMongoRepository[T]) Create(ctx context.Context, entity *T) error {
	if entity == nil {
		return errors.New("entity cannot be nil")
	}

	// 设置创建时间（如果实体有相关字段）
	r.setTimestamps(entity, true)

	coll := r.db.Collection(r.collection)
	resultID, err := coll.InsertOne(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	// 设置生成的ID
	if oid, ok := resultID.(primitive.ObjectID); ok {
		r.setEntityID(entity, oid)
	}

	return nil
}

// GetByID 根据ID获取实体
func (r *BaseMongoRepository[T]) GetByID(ctx context.Context, id primitive.ObjectID) (*T, error) {
	if id.IsZero() {
		return nil, errors.New("id cannot be empty")
	}

	coll := r.db.Collection(r.collection)
	var entity T
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&entity)
	if err != nil {
		if errors.Is(err, driver.ErrNoDocuments) {
			return nil, fmt.Errorf("entity not found with id: %s", id.Hex())
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return &entity, nil
}

// Update 更新实体（支持 Upsert）
func (r *BaseMongoRepository[T]) Update(ctx context.Context, entity *T) error {
	if entity == nil {
		return errors.New("entity cannot be nil")
	}

	id := r.getEntityID(entity)
	if id.IsZero() {
		return errors.New("entity ID cannot be empty")
	}

	// 设置更新时间戳
	r.setTimestamps(entity, false)

	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": entity}

	// 关键修改：启用 Upsert 选项
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update or insert entity: %w", err)
	}

	return nil
}

// UpdateByID 根据ID更新指定字段
func (r *BaseMongoRepository[T]) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	if id.IsZero() {
		return false, errors.New("id cannot be empty")
	}

	// 添加更新时间
	if update["$set"] != nil {
		if setUpdate, ok := update["$set"].(bson.M); ok {
			setUpdate["updated_at"] = primitive.NewDateTimeFromTime(time.Now())
		}
	} else {
		update["$set"] = bson.M{"updated_at": primitive.NewDateTimeFromTime(time.Now())}
	}

	coll := r.db.Collection(r.collection)
	result, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return false, fmt.Errorf("failed to update entity: %w", err)
	}

	return result.ModifiedCount > 0, nil
}

// Delete 删除实体
func (r *BaseMongoRepository[T]) Delete(ctx context.Context, id primitive.ObjectID) error {
	if id.IsZero() {
		return errors.New("id cannot be empty")
	}

	coll := r.db.Collection(r.collection)
	deletedCount, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	if deletedCount == 0 {
		return fmt.Errorf("entity not found with id: %s", id.Hex())
	}

	return nil
}

// CreateMany 批量创建实体
func (r *BaseMongoRepository[T]) CreateMany(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}

	// 设置时间戳
	for _, entity := range entities {
		r.setTimestamps(entity, true)
	}

	docs := make([]interface{}, len(entities))
	for i, entity := range entities {
		docs[i] = entity
	}

	coll := r.db.Collection(r.collection)
	resultIDs, err := coll.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to create entities: %w", err)
	}

	// 设置生成的ID
	for i, id := range resultIDs {
		if oid, ok := id.(primitive.ObjectID); ok && i < len(entities) {
			r.setEntityID(entities[i], oid)
		}
	}

	return nil
}

// BulkUpsert 批量插入或更新
func (r *BaseMongoRepository[T]) BulkUpsert(ctx context.Context, entities []*T) (int, error) {
	if len(entities) == 0 {
		return 0, nil
	}

	coll := r.db.Collection(r.collection)
	bulk := coll.BulkWrite()

	for _, entity := range entities {
		r.setTimestamps(entity, false) // 对于upsert，我们设置为更新模式

		id := r.getEntityID(entity)
		filter := bson.M{"_id": id}
		update := bson.M{"$set": entity}

		model := driver.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)

		bulk.AddModel(model)
	}

	result, err := bulk.Execute(ctx)
	if err != nil {
		return 0, fmt.Errorf("bulk upsert failed: %w", err)
	}

	return int(result.UpsertedCount() + result.ModifiedCount()), nil
}

// DeleteMany 批量删除
func (r *BaseMongoRepository[T]) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	coll := r.db.Collection(r.collection)
	deletedCount, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete entities: %w", err)
	}

	return deletedCount, nil
}

// GetAll 获取所有实体
func (r *BaseMongoRepository[T]) GetAll(ctx context.Context) ([]*T, error) {
	return r.GetByFilter(ctx, bson.M{})
}

// GetByFilter 根据过滤条件获取实体
func (r *BaseMongoRepository[T]) GetByFilter(ctx context.Context, filter interface{}) ([]*T, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find entities: %w", err)
	}
	defer cursor.Close(ctx)

	var entities []*T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, fmt.Errorf("failed to decode entity: %w", err)
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// GetOneByFilter 根据过滤条件获取单个实体
func (r *BaseMongoRepository[T]) GetOneByFilter(ctx context.Context, filter interface{}) (*T, error) {
	coll := r.db.Collection(r.collection)
	var entity T
	err := coll.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if errors.Is(err, driver.ErrNoDocuments) {
			return nil, nil // 没找到返回nil，不是错误
		}
		return nil, fmt.Errorf("failed to find entity: %w", err)
	}

	return &entity, nil
}

// Count 统计数量
func (r *BaseMongoRepository[T]) Count(ctx context.Context, filter interface{}) (int64, error) {
	coll := r.db.Collection(r.collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}

	return count, nil
}

// GetPaginated 分页查询
func (r *BaseMongoRepository[T]) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*T, error) {
	coll := r.db.Collection(r.collection)
	opts := options.Find().SetSkip(skip).SetLimit(limit)

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find entities: %w", err)
	}
	defer cursor.Close(ctx)

	var entities []*T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, fmt.Errorf("failed to decode entity: %w", err)
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// UpdateMany 批量更新
func (r *BaseMongoRepository[T]) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*driver.UpdateResult, error) {
	coll := r.db.Collection(r.collection)
	return coll.UpdateMany(ctx, filter, update, opts...)
}

// FindOneAndUpdate 查找并更新
//func (r *BaseMongoRepository[T]) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *driver.SingleResult {
//	coll := r.db.Collection(r.collection)
//	return coll.FindOneAndUpdate(ctx, filter, update, opts...)
//}

// Exists 检查实体是否存在
func (r *BaseMongoRepository[T]) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	if id.IsZero() {
		return false, errors.New("id cannot be empty")
	}

	count, err := r.Count(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByFilter 根据过滤条件检查实体是否存在
func (r *BaseMongoRepository[T]) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.Count(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// 辅助方法：设置时间戳
func (r *BaseMongoRepository[T]) setTimestamps(entity *T, isCreate bool) {
	val := reflect.ValueOf(entity).Elem()
	typ := val.Type()

	now := primitive.NewDateTimeFromTime(time.Now())

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		fieldName := fieldType.Tag.Get("bson")
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		// 设置创建时间
		if isCreate && (fieldName == "created_at" || fieldName == "CreatedAt") && field.Type() == reflect.TypeOf(now) {
			field.Set(reflect.ValueOf(now))
		}

		// 设置更新时间
		if (fieldName == "updated_at" || fieldName == "UpdatedAt") && field.Type() == reflect.TypeOf(now) {
			field.Set(reflect.ValueOf(now))
		}
	}
}

// 获取实体ID
func (r *BaseMongoRepository[T]) getEntityID(entity *T) primitive.ObjectID {
	if entity == nil {
		return primitive.NilObjectID
	}
	val := reflect.ValueOf(entity).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 跳过非导出字段（getEntityID也需要检查）
		if !field.CanInterface() { // 或 field.CanAddr()+field.CanSet()
			continue
		}

		// 解析bson标签（兼容带选项的标签）
		tag := fieldType.Tag.Get("bson")
		fieldName, _, _ := strings.Cut(tag, ",") // 分割标签和选项
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		// 匹配ID字段名，并检查类型兼容性（支持指针和非指针）
		if matchesIDField(fieldName) && isObjectIDType(field.Type()) {
			if field.Kind() == reflect.Ptr {
				if !field.IsNil() {
					return field.Elem().Interface().(primitive.ObjectID)
				}
				return primitive.NilObjectID
			}
			return field.Interface().(primitive.ObjectID)
		}
	}
	return primitive.NilObjectID
}

// 设置实体ID
func (r *BaseMongoRepository[T]) setEntityID(entity *T, id primitive.ObjectID) {
	if entity == nil {
		return
	}
	val := reflect.ValueOf(entity).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() { // 跳过不可修改字段
			continue
		}

		// 统一解析标签
		tag := fieldType.Tag.Get("bson")
		fieldName, _, _ := strings.Cut(tag, ",")
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		// 类型检查兼容指针和非指针
		if matchesIDField(fieldName) && isObjectIDType(field.Type()) {
			if field.Kind() == reflect.Ptr {
				newID := id // 避免取地址临时变量
				field.Set(reflect.ValueOf(&newID))
			} else {
				field.Set(reflect.ValueOf(id))
			}
			return
		}
	}
}

// 辅助函数：检查字段名是否匹配ID
func matchesIDField(name string) bool {
	return name == "_id" || name == "ID"
}

// 辅助函数：检查类型是否为primitive.ObjectID或其指针
func isObjectIDType(t reflect.Type) bool {
	return t == reflect.TypeOf(primitive.ObjectID{}) ||
		t == reflect.TypeOf(&primitive.ObjectID{})
}
