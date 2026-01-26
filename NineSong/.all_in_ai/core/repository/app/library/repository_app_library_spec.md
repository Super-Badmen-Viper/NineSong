# Repository App Library AI代码生成规范

## 代码结构规范

### Repository实现规范
- 文件路径: `repository/repository_app/repository_app_library/<library_entity>_repository.go`
- 继承基础Repository: `repository.BaseMongoRepository[<LibraryEntity>]`
- 使用`base_mongo_repository.go`作为基础实现
- 媒体库Repository支持路径验证和扫描状态管理

### 接口实现规范
- 实现`domain.BaseRepository[<LibraryEntity>]`接口
- 支持媒体库的获取、创建、更新、删除操作
- 包含媒体库专用方法如按路径获取、更新扫描状态等

## 代码生成模板

### 新建媒体库Repository实现
```
package repository_app_library

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

type <LibraryEntity>RepositoryImpl struct {
    repository.BaseMongoRepository[domain.<LibraryEntity>]
}

func New<LibraryEntity>RepositoryImpl(db *mongo.Database) domain.<LibraryEntity>Repository {
    return &<LibraryEntity>RepositoryImpl{
        BaseMongoRepository: repository.NewBaseMongoRepository[domain.<LibraryEntity>](db, "<library_collection_name>"),
    }
}

// GetByPath 根据路径获取媒体库
func (r *<LibraryEntity>RepositoryImpl) GetByPath(ctx context.Context, path string) (*domain.<LibraryEntity>, error) {
    filter := bson.M{"path": path}
    result, err := r.BaseMongoRepository.FindOne(ctx, filter)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, domain.Err<LibraryEntity>NotFound
        }
        return nil, err
    }
    return result, nil
}

// UpdateScanStatus 更新扫描状态
func (r *<LibraryEntity>RepositoryImpl) UpdateScanStatus(ctx context.Context, id primitive.ObjectID, status domain.LibraryStatus) error {
    filter := bson.M{"_id": id}
    
    update := bson.M{
        "$set": bson.M{
            "status":       status,
            "last_scanned": primitive.NewDateTimeFromTime(time.Now()),
            "updated_at":   primitive.NewDateTimeFromTime(time.Now()),
        },
    }
    
    result, err := r.BaseMongoRepository.Collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }
    
    if result.MatchedCount == 0 {
        return domain.Err<LibraryEntity>NotFound
    }
    
    return nil
}

// GetActiveLibraries 获取活跃的媒体库
func (r *<LibraryEntity>RepositoryImpl) GetActiveLibraries(ctx context.Context) ([]*domain.<LibraryEntity>, error) {
    filter := bson.M{"status": domain.LibraryStatusActive}
    results, err := r.BaseMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    return results, nil
}
```

### 媒体库Repository接口实现
```
// 实现 domain.BaseRepository[<LibraryEntity>] 接口的所有方法
// 通过嵌入 repository.BaseMongoRepository[<LibraryEntity>] 自动实现基础CRUD方法

// 自定义媒体库相关方法
func (r *<LibraryEntity>RepositoryImpl) GetByType(ctx context.Context, libType domain.LibraryType) ([]*domain.<LibraryEntity>, error) {
    filter := bson.M{"type": libType}
    results, err := r.BaseMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    return results, nil
}

func (r *<LibraryEntity>RepositoryImpl) GetByStatus(ctx context.Context, status domain.LibraryStatus) ([]*domain.<LibraryEntity>, error) {
    filter := bson.M{"status": status}
    results, err := r.BaseMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    return results, nil
}
```

## 代码生成规则

### Repository结构规则
- 使用结构体嵌入实现基础功能
- 通过`repository.BaseMongoRepository[<LibraryEntity>]`实现通用操作
- 自定义方法实现媒体库专用逻辑

### 查询操作规则
- 支持按路径、类型、状态等字段查询
- 使用bson.M构建查询条件
- 实现分页和排序功能

### 更新操作规则
- 更新时设置时间戳
- 验证更新前后的数据一致性
- 支持批量更新操作

### 命名规则
- Repository实现名: `<LibraryEntity>RepositoryImpl`
- 集合名称: 使用小写复数形式
- 方法名: 使用驼峰命名法

## 依赖管理规范
- Repository层依赖Domain层接口定义
- 依赖MongoDB驱动和基础Repository实现
- 通过构造函数注入数据库连接

## 常用代码片段

### 媒体库路径验证方法
```go
func (r *<Library>RepositoryImpl) IsPathUnique(ctx context.Context, path string, skipID primitive.ObjectID) (bool, error) {
    filter := bson.M{"path": path}
    if !skipID.IsZero() {
        filter["_id"] = bson.M{"$ne": skipID}
    }
    
    result, err := r.BaseMongoRepository.FindOne(ctx, filter)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return true, nil // 没有找到相同路径，路径唯一
        }
        return false, err
    }
    
    return result == nil, nil
}
```

### 媒体库批量操作方法
```go
func (r *<Library>RepositoryImpl) UpdateManyByType(ctx context.Context, libType domain.LibraryType, updateData bson.M) error {
    filter := bson.M{"type": libType}
    
    update := bson.M{
        "$set": bson.M{
            "updated_at": primitive.NewDateTimeFromTime(time.Now()),
        },
    }
    
    // 合并用户提供的更新数据
    for k, v := range updateData {
        update["$set"].(bson.M)[k] = v
    }
    
    _, err := r.BaseMongoRepository.Collection.UpdateMany(ctx, filter, update)
    return err
}
```

### 媒体库统计查询方法
```go
func (r *<Library>RepositoryImpl) GetLibraryStats(ctx context.Context) (map[string]interface{}, error) {
    pipeline := []bson.M{
        {
            "$group": bson.M{
                "_id": nil,
                "total": bson.M{"$sum": 1},
                "active": bson.M{
                    "$sum": bson.M{
                        "$cond": []interface{}{
                            bson.M{"$eq": []interface{}{"$status", domain.LibraryStatusActive}},
                            1,
                            0,
                        },
                    },
                },
                "byType": bson.M{
                    "$push": bson.M{
                        "type": "$type",
                        "status": "$status",
                    },
                },
            },
        },
    }
    
    cursor, err := r.BaseMongoRepository.Collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var results []map[string]interface{}
    if err = cursor.All(ctx, &results); err != nil {
        return nil, err
    }
    
    if len(results) == 0 {
        return map[string]interface{}{
            "total": 0,
            "active": 0,
            "byType": []interface{}{},
        }, nil
    }
    
    return results[0], nil
}
```

### 媒体库范围查询方法
```go
func (r *<Library>RepositoryImpl) GetLibrariesByPathPrefix(ctx context.Context, pathPrefix string) ([]*domain.<Library>, error) {
    filter := bson.M{
        "path": bson.M{
            "$regex":   "^" + regexp.QuoteMeta(pathPrefix),
            "$options": "i", // 不区分大小写
        },
    }
    
    results, err := r.BaseMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    
    return results, nil
}
```

### 媒体库事务处理方法
```go
func (r *<Library>RepositoryImpl) UpdateLibraryWithFiles(ctx context.Context, library *domain.<Library>, files []interface{}) error {
    session, err := r.BaseMongoRepository.Client.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)
    
    err = session.StartTransaction()
    if err != nil {
        return err
    }
    
    err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
        // 更新媒体库信息
        err := r.BaseMongoRepository.UpdateByID(sc, library.ID, library)
        if err != nil {
            return err
        }
        
        // 可以在这里添加文件相关的操作
        // 例如：批量插入或更新相关文件记录
        
        return nil
    })
    
    if err != nil {
        session.AbortTransaction(ctx)
        return err
    }
    
    return session.CommitTransaction(ctx)
}
```