package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseUsecase 通用Usecase接口
type BaseUsecase[T any] interface {
	// 基础CRUD操作
	Create(ctx context.Context, entity *T) (*T, error)
	GetByID(ctx context.Context, id string) (*T, error)
	Update(ctx context.Context, entity *T) error
	UpdateByID(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error

	// 批量操作
	CreateMany(ctx context.Context, entities []*T) error
	DeleteByIDs(ctx context.Context, ids []string) error

	// 查询操作
	GetAll(ctx context.Context) ([]*T, error)
	GetPaginated(ctx context.Context, page, pageSize int) ([]*T, int64, error)
	Count(ctx context.Context) (int64, error)

	// 验证操作
	Exists(ctx context.Context, id string) (bool, error)
}

// ConfigUsecase 配置类Usecase接口
type ConfigUsecase[T any] interface {
	Get(ctx context.Context) (*T, error)
	Update(ctx context.Context, config *T) error
	GetAll(ctx context.Context) ([]*T, error)
	ReplaceAll(ctx context.Context, configs []*T) error
}

// SearchableUsecase 可搜索Usecase接口
type SearchableUsecase[T any] interface {
	BaseUsecase[T]
	GetByName(ctx context.Context, name string) (*T, error)
	Search(ctx context.Context, keyword string) ([]*T, error)
	GetByNamePattern(ctx context.Context, pattern string) ([]*T, error)
}

// BaseUsecaseImpl 通用Usecase实现
type BaseUsecaseImpl[T any] struct {
	repo    domain.BaseRepository[T]
	timeout time.Duration
}

// NewBaseUsecase 创建通用Usecase实例
func NewBaseUsecase[T any](repo domain.BaseRepository[T], timeout time.Duration) BaseUsecase[T] {
	return &BaseUsecaseImpl[T]{
		repo:    repo,
		timeout: timeout,
	}
}

// Create 创建实体
func (uc *BaseUsecaseImpl[T]) Create(ctx context.Context, entity *T) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if entity == nil {
		return nil, errors.New("entity cannot be nil")
	}

	if err := uc.repo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return entity, nil
}

// GetByID 根据ID获取实体
func (uc *BaseUsecaseImpl[T]) GetByID(ctx context.Context, id string) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %w", err)
	}

	entity, err := uc.repo.GetByID(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return entity, nil
}

// Update 更新实体
func (uc *BaseUsecaseImpl[T]) Update(ctx context.Context, entity *T) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if entity == nil {
		return errors.New("entity cannot be nil")
	}

	if err := uc.repo.Update(ctx, entity); err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	return nil
}

// UpdateByID 根据ID更新指定字段
func (uc *BaseUsecaseImpl[T]) UpdateByID(ctx context.Context, id string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if id == "" {
		return errors.New("id cannot be empty")
	}

	if len(updates) == 0 {
		return errors.New("updates cannot be empty")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id format: %w", err)
	}

	updateDoc := bson.M{"$set": updates}
	_, err = uc.repo.UpdateByID(ctx, objID, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	return nil
}

// Delete 删除实体
func (uc *BaseUsecaseImpl[T]) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if id == "" {
		return errors.New("id cannot be empty")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id format: %w", err)
	}

	if err := uc.repo.Delete(ctx, objID); err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	return nil
}

// CreateMany 批量创建实体
func (uc *BaseUsecaseImpl[T]) CreateMany(ctx context.Context, entities []*T) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if len(entities) == 0 {
		return nil
	}

	for _, entity := range entities {
		if entity == nil {
			return errors.New("entity cannot be nil")
		}
	}

	if err := uc.repo.CreateMany(ctx, entities); err != nil {
		return fmt.Errorf("failed to create entities: %w", err)
	}

	return nil
}

// DeleteByIDs 根据ID列表批量删除
func (uc *BaseUsecaseImpl[T]) DeleteByIDs(ctx context.Context, ids []string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if len(ids) == 0 {
		return nil
	}

	objIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			return errors.New("id cannot be empty")
		}
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid id format: %s", id)
		}
		objIDs = append(objIDs, objID)
	}

	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	if _, err := uc.repo.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("failed to delete entities: %w", err)
	}

	return nil
}

// GetAll 获取所有实体
func (uc *BaseUsecaseImpl[T]) GetAll(ctx context.Context) ([]*T, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	entities, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	return entities, nil
}

// GetPaginated 分页获取实体
func (uc *BaseUsecaseImpl[T]) GetPaginated(ctx context.Context, page, pageSize int) ([]*T, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	entities, err := uc.repo.GetPaginated(ctx, bson.M{}, skip, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get paginated entities: %w", err)
	}

	total, err := uc.repo.Count(ctx, bson.M{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count entities: %w", err)
	}

	return entities, total, nil
}

// Count 统计实体数量
func (uc *BaseUsecaseImpl[T]) Count(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	count, err := uc.repo.Count(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}

	return count, nil
}

// Exists 检查实体是否存在
func (uc *BaseUsecaseImpl[T]) Exists(ctx context.Context, id string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if id == "" {
		return false, errors.New("id cannot be empty")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, fmt.Errorf("invalid id format: %w", err)
	}

	exists, err := uc.repo.Exists(ctx, objID)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists, nil
}

// ConfigUsecaseImpl 配置类Usecase实现
type ConfigUsecaseImpl[T any] struct {
	repo    domain.ConfigRepository[T]
	timeout time.Duration
}

// NewConfigUsecase 创建配置Usecase实例
func NewConfigUsecase[T any](repo domain.ConfigRepository[T], timeout time.Duration) ConfigUsecase[T] {
	return &ConfigUsecaseImpl[T]{
		repo:    repo,
		timeout: timeout,
	}
}

// Get 获取配置
func (uc *ConfigUsecaseImpl[T]) Get(ctx context.Context) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	config, err := uc.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return config, nil
}

// Update 更新配置
func (uc *ConfigUsecaseImpl[T]) Update(ctx context.Context, config *T) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if config == nil {
		return errors.New("config cannot be nil")
	}

	if err := uc.repo.Update(ctx, config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}

// GetAll 获取所有配置
func (uc *ConfigUsecaseImpl[T]) GetAll(ctx context.Context) ([]*T, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	configs, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs: %w", err)
	}

	return configs, nil
}

// ReplaceAll 替换所有配置
func (uc *ConfigUsecaseImpl[T]) ReplaceAll(ctx context.Context, configs []*T) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if configs == nil {
		return errors.New("configs cannot be nil")
	}

	if err := uc.repo.ReplaceAll(ctx, configs); err != nil {
		return fmt.Errorf("failed to replace configs: %w", err)
	}

	return nil
}

// SearchableUsecaseImpl 可搜索Usecase实现
type SearchableUsecaseImpl[T any] struct {
	*BaseUsecaseImpl[T]
	searchRepo domain.SearchableRepository[T]
}

// NewSearchableUsecase 创建可搜索Usecase实例
func NewSearchableUsecase[T any](repo domain.SearchableRepository[T], timeout time.Duration) SearchableUsecase[T] {
	baseUsecase := &BaseUsecaseImpl[T]{
		repo:    repo,
		timeout: timeout,
	}
	return &SearchableUsecaseImpl[T]{
		BaseUsecaseImpl: baseUsecase,
		searchRepo:      repo,
	}
}

// GetByName 根据名称获取实体
func (uc *SearchableUsecaseImpl[T]) GetByName(ctx context.Context, name string) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	entity, err := uc.searchRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity by name: %w", err)
	}

	return entity, nil
}

// Search 搜索实体（通过名称模式）
func (uc *SearchableUsecaseImpl[T]) Search(ctx context.Context, keyword string) ([]*T, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if keyword == "" {
		return nil, errors.New("keyword cannot be empty")
	}

	entities, err := uc.searchRepo.GetByNamePattern(ctx, keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}

	return entities, nil
}

// GetByNamePattern 根据名称模式获取实体
func (uc *SearchableUsecaseImpl[T]) GetByNamePattern(ctx context.Context, pattern string) ([]*T, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if pattern == "" {
		return nil, errors.New("pattern cannot be empty")
	}

	entities, err := uc.searchRepo.GetByNamePattern(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities by pattern: %w", err)
	}

	return entities, nil
}
