package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConfigMongoRepository 配置类MongoDB Repository实现
type ConfigMongoRepository[T any] struct {
	*BaseMongoRepository[T]
}

// NewConfigMongoRepository 创建新的配置Repository实例
func NewConfigMongoRepository[T any](db mongo.Database, collection string) domain.ConfigRepository[T] {
	baseRepo := &BaseMongoRepository[T]{
		db:         db,
		collection: collection,
	}
	return &ConfigMongoRepository[T]{
		BaseMongoRepository: baseRepo,
	}
}

// Get 获取配置（单例模式，获取第一个配置）
func (r *ConfigMongoRepository[T]) Get(ctx context.Context) (*T, error) {
	coll := r.db.Collection(r.collection)
	var config T
	err := coll.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("configuration not found: %w", err)
	}
	return &config, nil
}

// Upsert 更新配置
func (r *ConfigMongoRepository[T]) Upsert(ctx context.Context, config *T) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	id := r.getEntityID(config)
	if id.IsZero() {
		return errors.New("config ID cannot be empty")
	}

	// 设置更新时间
	r.setTimestamps(config, false)

	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": config}
	opts := options.Update().SetUpsert(true)

	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}

// GetAll 获取所有配置
func (r *ConfigMongoRepository[T]) GetAll(ctx context.Context) ([]*T, error) {
	return r.BaseMongoRepository.GetAll(ctx)
}

// ReplaceAll 替换所有配置（非事务模式）
func (r *ConfigMongoRepository[T]) ReplaceAll(ctx context.Context, configs []*T) error {
	// 获取集合句柄
	coll := r.db.Collection(r.collection)

	// 如果没有新配置
	if len(configs) == 0 {
		// 删除所有现有配置
		if _, err := coll.DeleteMany(ctx, bson.M{}); err != nil {
			return fmt.Errorf("failed to delete existing configs: %w", err)
		}
		return nil
	}

	// 1. 删除所有现有配置
	if _, err := coll.DeleteMany(ctx, bson.M{}); err != nil {
		return fmt.Errorf("failed to delete existing configs: %w", err)
	}

	// 2. 准备插入文档
	docs := make([]interface{}, len(configs))
	for i, config := range configs {
		r.setTimestamps(config, true) // 设置时间戳
		docs[i] = config
	}

	// 3. 插入新配置
	if _, err := coll.InsertMany(ctx, docs); err != nil {
		return fmt.Errorf("failed to insert new configs: %w", err)
	}

	return nil
}

// deleteAll 删除所有配置
func (r *ConfigMongoRepository[T]) deleteAll(ctx context.Context) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to delete all configs: %w", err)
	}
	return nil
}

// SearchableMongoRepository 可搜索MongoDB Repository实现
type SearchableMongoRepository[T any] struct {
	*BaseMongoRepository[T]
}

// NewSearchableMongoRepository 创建新的可搜索Repository实例
func NewSearchableMongoRepository[T any](db mongo.Database, collection string) domain.SearchableRepository[T] {
	baseRepo := &BaseMongoRepository[T]{
		db:         db,
		collection: collection,
	}
	return &SearchableMongoRepository[T]{
		BaseMongoRepository: baseRepo,
	}
}

// GetByName 根据名称获取实体
func (r *SearchableMongoRepository[T]) GetByName(ctx context.Context, name string) (*T, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	filter := bson.M{"name": name}
	return r.GetOneByFilter(ctx, filter)
}

// GetByNamePattern 根据名称模式获取实体
func (r *SearchableMongoRepository[T]) GetByNamePattern(ctx context.Context, pattern string) ([]*T, error) {
	if pattern == "" {
		return nil, errors.New("pattern cannot be empty")
	}

	filter := bson.M{"name": bson.M{"$regex": pattern, "$options": "i"}}
	return r.GetByFilter(ctx, filter)
}

// GetByPath 根据路径获取实体
func (r *SearchableMongoRepository[T]) GetByPath(ctx context.Context, path string) (*T, error) {
	if path == "" {
		return nil, errors.New("path cannot be empty")
	}

	filter := bson.M{"path": path}
	return r.GetOneByFilter(ctx, filter)
}

// GetByFolder 根据文件夹路径获取实体
func (r *SearchableMongoRepository[T]) GetByFolder(ctx context.Context, folderPath string) ([]*T, error) {
	if folderPath == "" {
		return nil, errors.New("folder path cannot be empty")
	}

	filter := bson.M{"folder_path": folderPath}
	return r.GetByFilter(ctx, filter)
}

// GetAllSorted 获取所有实体并排序
func (r *SearchableMongoRepository[T]) GetAllSorted(ctx context.Context, sortField string, ascending bool) ([]*T, error) {
	coll := r.db.Collection(r.collection)

	sortOrder := 1
	if !ascending {
		sortOrder = -1
	}

	opts := options.Find().SetSort(bson.M{sortField: sortOrder})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
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

// GetPaginatedSorted 分页查询并排序
func (r *SearchableMongoRepository[T]) GetPaginatedSorted(ctx context.Context, filter interface{}, skip, limit int64, sortField string, ascending bool) ([]*T, error) {
	coll := r.db.Collection(r.collection)

	sortOrder := 1
	if !ascending {
		sortOrder = -1
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{sortField: sortOrder})

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

// AuditableMongoRepository 可审计MongoDB Repository实现
type AuditableMongoRepository[T any] struct {
	*BaseMongoRepository[T]
}

// NewAuditableMongoRepository 创建新的可审计Repository实例
func NewAuditableMongoRepository[T any](db mongo.Database, collection string) domain.AuditableRepository[T] {
	baseRepo := &BaseMongoRepository[T]{
		db:         db,
		collection: collection,
	}
	return &AuditableMongoRepository[T]{
		BaseMongoRepository: baseRepo,
	}
}

// GetCreatedAfter 获取指定时间后创建的实体
func (r *AuditableMongoRepository[T]) GetCreatedAfter(ctx context.Context, after primitive.DateTime) ([]*T, error) {
	filter := bson.M{"created_at": bson.M{"$gt": after}}
	return r.GetByFilter(ctx, filter)
}

// GetUpdatedAfter 获取指定时间后更新的实体
func (r *AuditableMongoRepository[T]) GetUpdatedAfter(ctx context.Context, after primitive.DateTime) ([]*T, error) {
	filter := bson.M{"updated_at": bson.M{"$gt": after}}
	return r.GetByFilter(ctx, filter)
}

// GetCreatedBetween 获取指定时间范围内创建的实体
func (r *AuditableMongoRepository[T]) GetCreatedBetween(ctx context.Context, start, end primitive.DateTime) ([]*T, error) {
	filter := bson.M{
		"created_at": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	return r.GetByFilter(ctx, filter)
}
