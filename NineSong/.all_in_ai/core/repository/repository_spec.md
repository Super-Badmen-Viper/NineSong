# Repository层AI代码生成规范

## 代码结构规范

### 基础Repository规范
- 文件路径: `repository/base_mongo_repository.go`
- 泛型实现: `BaseMongoRepository[T any]`
- 接口定义: `domain.BaseRepository[T]`
- 基础CRUD操作通过泛型实现

### 配置Repository规范
- 文件路径: `repository/base_config_repository.go`
- 继承关系: `ConfigMongoRepository[T]` 继承 `BaseMongoRepository[T]`
- 特殊方法: `Get()`, `Upsert()`
- 单例配置模式支持

### 自定义Repository规范
- 文件路径: `repository/repository_<domain>/repository_<subdomain>/<entity_name>_repository.go`
- 实现对应domain的Repository接口
- 可以添加自定义查询方法

## 代码生成模板

### 新建Repository实现
```
package repository_<subdomain>

import (
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_<domain>/domain_<subdomain>"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// <EntityName>Repository is an alias for the generic BaseRepository.
type <EntityName>Repository interface {
    domain.BaseRepository[domain_<subdomain>.<EntityName>]
    // 添加自定义方法
    CustomQuery(ctx context.Context, param string) ([]*domain_<subdomain>.<EntityName>, error)
}

// New<EntityName>Repository creates a new repository for <entity_name> entities.
// It uses the generic BaseMongoRepository implementation.
func New<EntityName>Repository(db mongo.Database, collection string) <EntityName>Repository {
    return repository.NewBaseMongoRepository[domain_<subdomain>.<EntityName>](db, collection)
}
```

### 配置Repository实现模板
```
package repository_<subdomain>

import (
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_<domain>/domain_<subdomain>"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// <EntityName>Repository is an alias for the generic ConfigRepository.
type <EntityName>Repository interface {
    domain.ConfigRepository[domain_<subdomain>.<EntityName>]
    // 添加自定义方法
    CustomConfigMethod(ctx context.Context) (*domain_<subdomain>.<EntityName>, error)
}

// New<EntityName>Repository creates a new repository for <entity_name> configurations.
// It uses the generic ConfigMongoRepository implementation.
func New<EntityName>Repository(db mongo.Database, collection string) <EntityName>Repository {
    return repository.NewConfigMongoRepository[domain_<subdomain>.<EntityName>](db, collection)
}
```

### 自定义Repository实现模板
```
package repository_<subdomain>

import (
    "context"
    "errors"
    "fmt"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_<domain>/domain_<subdomain>"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

type custom<EntityName>Repository struct {
    *repository.BaseMongoRepository[domain_<subdomain>.<EntityName>]
    db         mongo.Database
    collection string
}

// NewCustom<EntityName>Repository creates a new custom repository for <entity_name> entities.
func NewCustom<EntityName>Repository(db mongo.Database, collection string) domain_<subdomain>.<EntityName>Repository {
    baseRepo := repository.NewBaseMongoRepository[domain_<subdomain>.<EntityName>](db, collection)
    return &custom<EntityName>Repository{
        BaseMongoRepository: baseRepo,
        db:                  db,
        collection:          collection,
    }
}

// 实现自定义方法
func (r *custom<EntityName>Repository) CustomQuery(ctx context.Context, param string) ([]*domain_<subdomain>.<EntityName>, error) {
    collection := r.db.Collection(r.collection)
    
    filter := bson.M{"field_name": param} // 根据实际需求修改查询条件
    
    cursor, err := collection.Find(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("failed to execute custom query: %w", err)
    }
    defer cursor.Close(ctx)
    
    var entities []*domain_<subdomain>.<EntityName>
    if err = cursor.All(ctx, &entities); err != nil {
        return nil, fmt.Errorf("failed to decode entities: %w", err)
    }
    
    return entities, nil
}
```

## 代码生成规则

### 方法生成规则
- 基础CRUD方法由泛型Repository自动提供
- 特殊业务方法需要单独实现
- 方法命名遵循Go语言规范
- 所有方法必须接收`context.Context`作为第一个参数
- 查询方法应支持分页、排序和过滤

### 错误处理规则
- 使用`errors.New()`创建新错误
- 使用`fmt.Errorf()`包装错误
- 错误消息使用英文描述
- 使用`%w`动词包装底层错误以便错误追踪

### 返回值规范
- 单个实体: `(*T, error)`
- 多个实体: `([]*T, error)`
- 计数操作: `(int64, error)`
- 布尔操作: `(bool, error)`
- ID操作: `(primitive.ObjectID, error)`

### 查询规范
- 使用bson.M构建查询条件
- 支持复合查询条件
- 合理使用索引字段进行查询
- 对于复杂查询考虑使用聚合管道

## 依赖注入规范
- Repository依赖通过构造函数注入到UseCase
- 不允许在Repository中直接访问其他Repository
- 使用接口类型进行依赖注入
- 避免循环依赖

## 命名规范
- Repository接口: `<EntityName>Repository`
- Repository实现: `New<EntityName>Repository()`
- 变量命名: `repo`, `db`, `collection`, `ctx`, `filter`
- 方法命名: `GetBy<Field>`, `FindBy<Field>`, `UpdateBy<Field>`, `DeleteBy<Field>`

## 常用代码片段

### 自定义查询方法
```go
func (r *<EntityName>Repository) FindBy<Field>(ctx context.Context, fieldVal string) ([]*<Entity>, error) {
    collection := r.db.Collection(r.collection)
    
    filter := bson.M{"field_name": fieldVal}
    cursor, err := collection.Find(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("failed to find entities by field: %w", err)
    }
    defer cursor.Close(ctx)
    
    var entities []*<Entity>
    if err = cursor.All(ctx, &entities); err != nil {
        return nil, fmt.Errorf("failed to decode entities: %w", err)
    }
    
    return entities, nil
}
```

### 批量操作方法
```go
func (r *<EntityName>Repository) BatchCreate(ctx context.Context, entities []*<Entity>) error {
    if len(entities) == 0 {
        return errors.New("entities list cannot be empty")
    }
    
    collection := r.db.Collection(r.collection)
    
    // 转换为interface{}数组
    docs := make([]interface{}, len(entities))
    for i, entity := range entities {
        docs[i] = entity
    }
    
    _, err := collection.InsertMany(ctx, docs)
    if err != nil {
        return fmt.Errorf("failed to batch create entities: %w", err)
    }
    
    return nil
}
```

### 聚合查询方法
```go
func (r *<EntityName>Repository) AggregateQuery(ctx context.Context, matchStage bson.M) ([]*<Entity>, error) {
    collection := r.db.Collection(r.collection)
    
    pipeline := []bson.M{
        {"$match": matchStage},
        // 可以添加更多聚合阶段
        // {"$group": ...},
        // {"$sort": ...},
    }
    
    cursor, err := collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("failed to execute aggregate query: %w", err)
    }
    defer cursor.Close(ctx)
    
    var results []*<Entity>
    if err = cursor.All(ctx, &results); err != nil {
        return nil, fmt.Errorf("failed to decode aggregate results: %w", err)
    }
    
    return results, nil
}
```

### 分页查询方法
```go
func (r *<EntityName>Repository) GetPaginated(ctx context.Context, page, pageSize int) ([]*<Entity>, int64, error) {
    collection := r.db.Collection(r.collection)
    
    // 计算跳过的文档数
    skip := int64((page - 1) * pageSize)
    
    // 查询总数
    total, err := collection.CountDocuments(ctx, bson.M{})
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count documents: %w", err)
    }
    
    // 查询分页数据
    cursor, err := collection.Find(ctx, 
        bson.M{}, 
        options.Find().SetSkip(skip).SetLimit(int64(pageSize)),
    )
    if err != nil {
        return nil, 0, fmt.Errorf("failed to find paginated data: %w", err)
    }
    defer cursor.Close(ctx)
    
    var entities []*<Entity>
    if err = cursor.All(ctx, &entities); err != nil {
        return nil, 0, fmt.Errorf("failed to decode paginated data: %w", err)
    }
    
    return entities, total, nil
}
```

### 复合条件查询
```go
func (r *<EntityName>Repository) FindWithFilters(ctx context.Context, filters map[string]interface{}, sortOptions map[string]int) ([]*<Entity>, error) {
    collection := r.db.Collection(r.collection)
    
    // 构建查询条件
    filter := bson.M{}
    for key, value := range filters {
        filter[key] = value
    }
    
    // 构建排序选项
    sort := bson.D{}
    for field, order := range sortOptions {
        sort = append(sort, bson.E{Key: field, Value: order})
    }
    
    options := options.Find().SetSort(sort)
    cursor, err := collection.Find(ctx, filter, options)
    if err != nil {
        return nil, fmt.Errorf("failed to find entities with filters: %w", err)
    }
    defer cursor.Close(ctx)
    
    var entities []*<Entity>
    if err = cursor.All(ctx, &entities); err != nil {
        return nil, fmt.Errorf("failed to decode entities: %w", err)
    }
    
    return entities, nil
}
```