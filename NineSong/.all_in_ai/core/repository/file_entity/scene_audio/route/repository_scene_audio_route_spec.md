# Repository Scene Audio Route AI代码生成规范

## 代码结构规范

### Repository实现规范
- 文件路径: `repository/repository_file_entity/scene_audio/scene_audio_route_repository/<entity>_repository.go`
- 继承基础Repository: `repository.BaseMongoRepository[<RouteModelEntity>]`
- 使用`base_mongo_repository.go`作为基础实现
- 音频路由Repository支持路由关联数据和复杂查询

### 接口实现规范
- 实现`domain.BaseRepository[<RouteModelEntity>]`接口
- 支持音频路由实体的获取、创建、更新、删除操作
- 包含音频路由专用方法如按路由ID获取、路由数据搜索等

## 代码生成模板

### 新建音频路由Repository实现
```
package scene_audio_route_repository

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
)

type <RouteEntity>RepositoryImpl struct {
    repository.BaseMongoRepository[scene_audio_route_models.<RouteModelEntity>]
}

func New<RouteEntity>RepositoryImpl(db *mongo.Database) domain.<RouteInterfaceEntity>Repository {
    return &<RouteEntity>RepositoryImpl{
        BaseMongoRepository: repository.NewBaseMongoRepository[scene_audio_route_models.<RouteModelEntity>](db, "<route_entity_collection_name>"),
    }
}

// GetBy<FieldName> 根据<field>获取音频路由数据
func (r *<RouteEntity>RepositoryImpl) GetBy<FieldName>(ctx context.Context, <field_name> string) (*scene_audio_route_models.<RouteModelEntity>, error) {
    filter := bson.M{"<field_name>": <field_name>}
    result, err := r.BaseMongoRepository.FindOne(ctx, filter)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, domain.Err<RouteEntity>NotFound
        }
        return nil, err
    }
    return result, nil
}

// GetBy<FieldName>WithPagination 根据<field>分页获取音频路由数据
func (r *<RouteEntity>RepositoryImpl) GetBy<FieldName>WithPagination(ctx context.Context, <field_name> string, page, limit int) ([]*scene_audio_route_models.<RouteModelEntity>, error) {
    filter := bson.M{"<field_name>": <field_name>}
    
    options := r.GetPaginationOptions(page, limit)
    results, err := r.BaseMongoRepository.Find(ctx, filter, options)
    if err != nil {
        return nil, err
    }
    return results, nil
}

// Search<RouteEntity> 搜索音频路由数据
func (r *<RouteEntity>RepositoryImpl) Search<RouteEntity>(ctx context.Context, query string) ([]*scene_audio_route_models.<RouteModelEntity>, error) {
    filter := bson.M{
        "$or": []bson.M{
            {"<field_name>": bson.M{"$regex": query, "$options": "i"}},
            {"<field_name2>": bson.M{"$regex": query, "$options": "i"}},
        },
    }
    
    results, err := r.BaseMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    return results, nil
}

// Get<RouteEntity>ForRoute 根据路由ID获取音频数据
func (r *<RouteEntity>RepositoryImpl) Get<RouteEntity>ForRoute(ctx context.Context, routeId string) ([]*scene_audio_route_models.<RouteModelEntity>, error) {
    objectID, err := primitive.ObjectIDFromHex(routeId)
    if err != nil {
        return nil, err
    }
    
    filter := bson.M{"route_id": objectID}
    results, err := r.BaseMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    return results, nil
}
```

### 音频路由Repository接口实现
```
// 实现 domain.BaseRepository[<RouteModelEntity>] 接口的所有方法
// 通过嵌入 repository.BaseMongoRepository[<RouteModelEntity>] 自动实现基础CRUD方法

// 自定义音频路由相关方法
func (r *<RouteEntity>RepositoryImpl) GetBy<FieldName>Sorted(ctx context.Context, <field_name> string, sortBy string, order int) ([]*scene_audio_route_models.<RouteModelEntity>, error) {
    filter := bson.M{"<field_name>": <field_name>}
    
    options := r.GetSortOptions(sortBy, order)
    results, err := r.BaseMongoRepository.Find(ctx, filter, options)
    if err != nil {
        return nil, err
    }
    return results, nil
}

func (r *<RouteEntity>RepositoryImpl) CountByRoute(ctx context.Context, routeId primitive.ObjectID) (int64, error) {
    filter := bson.M{"route_id": routeId}
    count, err := r.BaseMongoRepository.Collection.CountDocuments(ctx, filter)
    if err != nil {
        return 0, err
    }
    return count, nil
}
```

## 代码生成规则

### Repository结构规则
- 使用结构体嵌入实现基础功能
- 通过`repository.BaseMongoRepository[<RouteModelEntity>]`实现通用操作
- 自定义方法实现音频路由专用逻辑

### 查询操作规则
- 支持按路由ID查询音频数据
- 实现路由级别的分页和排序功能
- 使用bson.M构建复杂查询条件
- 支持跨路由数据搜索

### 更新操作规则
- 音频路由数据更新时设置时间戳
- 验证更新前后的数据一致性
- 支持批量更新操作

### 命名规则
- Repository实现名: `<RouteEntity>RepositoryImpl`
- 集合名称: 使用小写复数形式
- 方法名: 使用驼峰命名法

## 依赖管理规范
- Repository层依赖Domain层接口定义
- 依赖MongoDB驱动和基础Repository实现
- 依赖音频路由模型定义

## 常用代码片段

### 音频路由关联查询方法
```go
func (r *<RouteEntity>RepositoryImpl) GetWithRouteDetails(ctx context.Context, id primitive.ObjectID) (*scene_audio_route_models.<RouteModelEntity>, error) {
    pipeline := []bson.M{
        {"$match": bson.M{"_id": id}},
        // 添加路由详情关联查询管道
        {
            "$lookup": bson.M{
                "from":         "routes", // 假设路由信息存储在routes集合
                "localField":   "route_id",
                "foreignField": "_id",
                "as":           "route_details",
            },
        },
    }
    
    cursor, err := r.BaseMongoRepository.Collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var results []scene_audio_route_models.<RouteModelEntity>
    if err = cursor.All(ctx, &results); err != nil {
        return nil, err
    }
    
    if len(results) == 0 {
        return nil, domain.Err<RouteEntity>NotFound
    }
    
    return &results[0], nil
}
```

### 音频路由数据批量操作方法
```go
func (r *<RouteEntity>RepositoryImpl) UpdateManyByRoute(ctx context.Context, routeId primitive.ObjectID, updateData bson.M) error {
    filter := bson.M{"route_id": routeId}
    
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

### 音频路由数据统计查询方法
```go
func (r *<RouteEntity>RepositoryImpl) Get<RouteEntity>StatsByRoute(ctx context.Context, routeId primitive.ObjectID) (map[string]interface{}, error) {
    pipeline := []bson.M{
        {"$match": bson.M{"route_id": routeId}},
        {
            "$group": bson.M{
                "_id": nil,
                "total": bson.M{"$sum": 1},
                "by<FieldName>": bson.M{
                    "$push": "$<field_name>",
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
            "by<FieldName>": []interface{}{},
        }, nil
    }
    
    return results[0], nil
}
```

### 音频路由数据范围查询方法
```go
func (r *<RouteEntity>RepositoryImpl) Get<RouteEntity>ByRouteAndDateRange(ctx context.Context, routeId primitive.ObjectID, start, end primitive.DateTime) ([]*scene_audio_route_models.<RouteEntity>, error) {
    filter := bson.M{
        "route_id": routeId,
        "created_at": bson.M{
            "$gte": start,
            "$lte": end,
        },
    }
    
    results, err := r.BaseMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    
    return results, nil
}
```

### 音频路由数据事务处理方法
```go
func (r *<RouteEntity>RepositoryImpl) Update<RouteEntity>WithRoute(ctx context.Context, entity *scene_audio_route_models.<RouteEntity>, routeData interface{}) error {
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
        // 更新音频路由数据
        err := r.BaseMongoRepository.UpdateByID(sc, entity.ID, entity)
        if err != nil {
            return err
        }
        
        // 可以在这里添加路由相关数据的操作
        // 例如：更新路由状态、统计信息等
        
        return nil
    })
    
    if err != nil {
        session.AbortTransaction(ctx)
        return err
    }
    
    return session.CommitTransaction(ctx)
}
```

### 音频路由批量查询方法
```go
func (r *<RouteEntity>RepositoryImpl) GetByMultipleRoutes(ctx context.Context, routeIds []primitive.ObjectID) ([]*scene_audio_route_models.<RouteEntity>, error) {
    filter := bson.M{
        "route_id": bson.M{
            "$in": routeIds,
        },
    }
    
    results, err := r.BaseMongoRepository.Find(ctx, filter, nil)
    if err != nil {
        return nil, err
    }
    
    return results, nil
}
```