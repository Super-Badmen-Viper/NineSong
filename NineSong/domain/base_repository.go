package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BaseRepository 通用Repository接口，提供标准CRUD操作
// T: 实体类型，必须包含ID字段
type BaseRepository[T any] interface {
	// 基础CRUD操作
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*T, error)
	Update(ctx context.Context, entity *T) error
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error)
	Delete(ctx context.Context, id primitive.ObjectID) error

	// 批量操作
	CreateMany(ctx context.Context, entities []*T) error
	BulkUpsert(ctx context.Context, entities []*T) (int, error)
	DeleteMany(ctx context.Context, filter interface{}) (int64, error)

	// 查询操作
	GetAll(ctx context.Context) ([]*T, error)
	GetByFilter(ctx context.Context, filter interface{}) ([]*T, error)
	GetOneByFilter(ctx context.Context, filter interface{}) (*T, error)
	Count(ctx context.Context, filter interface{}) (int64, error)

	// 分页查询
	GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*T, error)

	// MongoDB原生操作支持
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// 验证和检查
	Exists(ctx context.Context, id primitive.ObjectID) (bool, error)
	ExistsByFilter(ctx context.Context, filter interface{}) (bool, error)
}

// ConfigRepository 配置类Repository接口，适用于系统配置等场景
// T: 配置实体类型
type ConfigRepository[T any] interface {
	// 配置操作
	Get(ctx context.Context) (*T, error)
	Update(ctx context.Context, config *T) error

	// 批量配置操作（适用于多配置项场景）
	GetAll(ctx context.Context) ([]*T, error)
	ReplaceAll(ctx context.Context, configs []*T) error
}

// SearchableRepository 可搜索Repository接口，适用于需要复杂查询的场景
// T: 实体类型
type SearchableRepository[T any] interface {
	BaseRepository[T]

	// 名称查询（常用模式）
	GetByName(ctx context.Context, name string) (*T, error)
	GetByNamePattern(ctx context.Context, pattern string) ([]*T, error)

	// 路径查询（文件系统相关）
	GetByPath(ctx context.Context, path string) (*T, error)
	GetByFolder(ctx context.Context, folderPath string) ([]*T, error)

	// 排序查询
	GetAllSorted(ctx context.Context, sortField string, ascending bool) ([]*T, error)
	GetPaginatedSorted(ctx context.Context, filter interface{}, skip, limit int64, sortField string, ascending bool) ([]*T, error)
}

// AuditableRepository 可审计Repository接口，自动处理创建/更新时间
// T: 实体类型，应包含CreatedAt、UpdatedAt字段
type AuditableRepository[T any] interface {
	BaseRepository[T]

	// 审计查询
	GetCreatedAfter(ctx context.Context, after primitive.DateTime) ([]*T, error)
	GetUpdatedAfter(ctx context.Context, after primitive.DateTime) ([]*T, error)
	GetCreatedBetween(ctx context.Context, start, end primitive.DateTime) ([]*T, error)
}
