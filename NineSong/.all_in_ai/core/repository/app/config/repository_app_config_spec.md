# Repository App Config AI代码生成规范

## 代码结构规范

### Repository实现规范
- 文件路径: `repository/repository_app/repository_app_config/<config_type>_repository.go`
- 继承基础配置Repository: `repository.ConfigMongoRepository[<ConfigType>]`
- 使用`base_config_repository.go`作为基础实现
- 配置Repository支持单例模式和版本控制

### 接口实现规范
- 实现`domain.ConfigRepository[<ConfigType>]`接口
- 支持配置的获取、更新、删除操作
- 包含配置专用方法如获取默认配置、版本控制等

## 代码生成模板

### 新建配置Repository实现
```
package repository_app_config

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

type <ConfigType>RepositoryImpl struct {
    repository.ConfigMongoRepository[domain.<ConfigType>]
}

func New<ConfigType>RepositoryImpl(db *mongo.Database) domain.<ConfigType>Repository {
    return &<ConfigType>RepositoryImpl{
        ConfigMongoRepository: repository.NewConfigMongoRepository[domain.<ConfigType>](db, "<config_collection_name>"),
    }
}

// GetDefault 获取默认配置
func (r *<ConfigType>RepositoryImpl) GetDefault(ctx context.Context) (*domain.<ConfigType>, error) {
    filter := bson.M{} // 根据需要添加过滤条件
    result, err := r.ConfigMongoRepository.FindOne(ctx, filter)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, domain.ErrConfigNotFound
        }
        return nil, err
    }
    return result, nil
}

// UpdateWithVersion 带版本控制的更新
func (r *<ConfigType>RepositoryImpl) UpdateWithVersion(ctx context.Context, config *domain.<ConfigType>) error {
    filter := bson.M{
        "_id":     config.ID,
        "version": config.Version - 1, // 乐观锁版本控制
    }
    
    config.Version++ // 增加版本号
    config.SetTimestamps() // 设置时间戳
    
    update := bson.M{
        "$set": bson.M{
            "updated_at": config.UpdatedAt,
            "version":    config.Version,
            // 添加其他需要更新的字段
        },
    }
    
    result, err := r.ConfigMongoRepository.Collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }
    
    if result.MatchedCount == 0 {
        return domain.ErrVersionConflict
    }
    
    return nil
}
```

### 配置Repository接口实现
```
// 实现 domain.ConfigRepository[<ConfigType>] 接口的所有方法
// 通过嵌入 repository.ConfigMongoRepository[<ConfigType>] 自动实现基础CRUD方法

// 自定义配置相关方法
func (r *<ConfigType>RepositoryImpl) GetByName(ctx context.Context, name string) (*domain.<ConfigType>, error) {
    filter := bson.M{"name": name}
    result, err := r.ConfigMongoRepository.FindOne(ctx, filter)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, domain.ErrConfigNotFound
        }
        return nil, err
    }
    return result, nil
}

func (r *<ConfigType>RepositoryImpl) GetActiveConfigs(ctx context.Context) ([]*domain.<ConfigType>, error) {
    filter := bson.M{"status": "active"}
    results, err := r.ConfigMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    return results, nil
}
```

## 代码生成规则

### Repository结构规则
- 使用结构体嵌入实现基础功能
- 通过`repository.ConfigMongoRepository[<ConfigType>]`实现通用操作
- 自定义方法实现配置专用逻辑

### 查询操作规则
- 配置查询通常返回单例数据
- 支持按名称、状态等字段查询
- 使用bson.M构建查询条件
- 实现版本控制机制

### 更新操作规则
- 配置更新需要版本控制
- 使用乐观锁防止并发更新冲突
- 更新时设置时间戳
- 验证更新前后的数据一致性

### 命名规则
- Repository实现名: `<ConfigType>RepositoryImpl`
- 集合名称: 使用小写复数形式
- 方法名: 使用驼峰命名法

## 依赖管理规范
- Repository层依赖Domain层接口定义
- 依赖MongoDB驱动和基础Repository实现
- 通过构造函数注入数据库连接

## 常用代码片段

### 配置查询方法
```go
func (r *<Config>RepositoryImpl) GetBy<FieldName>(ctx context.Context, <field_name> string) (*domain.<Config>, error) {
    filter := bson.M{"<field_name>": <field_name>}
    result, err := r.ConfigMongoRepository.FindOne(ctx, filter)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, domain.Err<Config>NotFound
        }
        return nil, err
    }
    return result, nil
}
```

### 配置批量操作方法
```go
func (r *<Config>RepositoryImpl) UpdateManyBy<FieldName>(ctx context.Context, <field_name> string, updateData *domain.<Config>) error {
    filter := bson.M{"<field_name>": <field_name>}
    
    update := bson.M{
        "$set": bson.M{
            "updated_at": primitive.NewDateTimeFromTime(time.Now()),
            // 添加其他需要更新的字段
        },
    }
    
    _, err := r.ConfigMongoRepository.Collection.UpdateMany(ctx, filter, update)
    return err
}
```

### 配置验证方法
```go
func (r *<Config>RepositoryImpl) ValidateBeforeCreate(ctx context.Context, config *domain.<Config>) error {
    // 检查配置是否已存在
    existing, err := r.GetBy<FieldName>(ctx, config.<Field>)
    if err != nil && err != domain.Err<Config>NotFound {
        return err
    }
    
    if existing != nil {
        return domain.Err<Config>AlreadyExists
    }
    
    return nil
}
```

### 配置聚合查询方法
```go
func (r *<Config>RepositoryImpl) Get<Config>Stats(ctx context.Context) (map[string]interface{}, error) {
    pipeline := []bson.M{
        {
            "$group": bson.M{
                "_id": nil,
                "total": bson.M{"$sum": 1},
                "active": bson.M{
                    "$sum": bson.M{
                        "$cond": []interface{}{
                            bson.M{"$eq": []interface{}{"$status", "active"}},
                            1,
                            0,
                        },
                    },
                },
            },
        },
    }
    
    cursor, err := r.ConfigMongoRepository.Collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var results []map[string]interface{}
    if err = cursor.All(ctx, &results); err != nil {
        return nil, err
    }
    
    if len(results) == 0 {
        return map[string]interface{}{"total": 0, "active": 0}, nil
    }
    
    return results[0], nil
}
```

### 配置事务处理方法
```go
func (r *<Config>RepositoryImpl) UpdateConfigTransaction(ctx context.Context, config *domain.<Config>) error {
    session, err := r.ConfigMongoRepository.Client.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)
    
    err = session.StartTransaction()
    if err != nil {
        return err
    }
    
    err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
        // 执行配置更新操作
        err := r.UpdateWithVersion(sc, config)
        if err != nil {
            return err
        }
        
        // 可以在这里添加其他相关操作
        // 例如：记录配置变更日志
        
        return nil
    })
    
    if err != nil {
        session.AbortTransaction(ctx)
        return err
    }
    
    return session.CommitTransaction(ctx)
}
```