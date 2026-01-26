# UseCase Scene Audio Route AI代码生成规范

## 代码结构规范

### UseCase实现规范
- 文件路径: `usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase/<entity>_usecase.go`
- 继承基础UseCase: `usecase.BaseUsecase[<RouteModelEntity>]`
- 使用`base_usecase.go`作为基础实现
- 音频路由UseCase支持路由关联数据处理和复杂业务逻辑

### 接口实现规范
- 实现`domain.<RouteInterfaceEntity>Usecase`接口
- 支持音频路由的业务逻辑处理
- 包含音频路由专用方法如处理路由数据、验证路由数据等

## 代码生成模板

### 新建音频路由UseCase实现
```
package scene_audio_route_usecase

import (
    "context"
    "time"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
)

type <RouteEntity>UsecaseImpl struct {
    usecase.BaseUsecase[scene_audio_route_models.<RouteModelEntity>]
    routeRepo domain.<RouteInterfaceEntity>Repository
}

func New<RouteEntity>UsecaseImpl(routeRepo domain.<RouteInterfaceEntity>Repository) domain.<RouteInterfaceEntity>Usecase {
    return &<RouteEntity>UsecaseImpl{
        BaseUsecase: usecase.NewBaseUsecase[scene_audio_route_models.<RouteModelEntity>](routeRepo),
        routeRepo:   routeRepo,
    }
}

// Process<RouteEntity>DataForRoute 处理路由音频数据
func (uc *<RouteEntity>UsecaseImpl) Process<RouteEntity>DataForRoute(ctx context.Context, routeId string, data *scene_audio_route_models.<RouteModelEntity>) error {
    // 验证路由ID
    if !isValidObjectId(routeId) {
        return domain.ErrInvalidRouteId
    }
    
    // 验证数据
    if err := uc.validateRouteData(ctx, routeId, data); err != nil {
        return err
    }
    
    // 设置路由ID
    objectID, err := primitive.ObjectIDFromHex(routeId)
    if err != nil {
        return err
    }
    data.RouteID = objectID
    
    // 设置时间戳
    data.SetTimestamps()
    
    // 保存数据
    return uc.routeRepo.Create(ctx, data)
}

// Validate<RouteEntity>DataForRoute 验证路由音频数据
func (uc *<RouteEntity>UsecaseImpl) Validate<RouteEntity>DataForRoute(ctx context.Context, routeId string, data *scene_audio_route_models.<RouteModelEntity>) error {
    // 验证路由ID
    if !isValidObjectId(routeId) {
        return domain.ErrInvalidRouteId
    }
    
    // 验证数据
    if err := data.Validate(); err != nil {
        return err
    }
    
    // 自定义验证逻辑
    if err := uc.validateRouteData(ctx, routeId, data); err != nil {
        return err
    }
    
    return nil
}

// Get<RouteEntity>WithRouteDetails 获取带路由详情的音频数据
func (uc *<RouteEntity>UsecaseImpl) Get<RouteEntity>WithRouteDetails(ctx context.Context, id string) (*scene_audio_route_models.<RouteModelEntity>, error) {
    // 验证ID
    if !isValidObjectId(id) {
        return nil, domain.ErrInvalidId
    }
    
    // 获取数据
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    
    data, err := uc.routeRepo.GetByID(ctx, objectID.Hex())
    if err != nil {
        return nil, err
    }
    
    // 获取路由详情
    // 这里可以根据需要添加路由详情获取逻辑
    
    return data, nil
}

// Get<RouteEntity>ByRoute 根据路由获取音频数据
func (uc *<RouteEntity>UsecaseImpl) Get<RouteEntity>ByRoute(ctx context.Context, routeId string) ([]*scene_audio_route_models.<RouteModelEntity>, error) {
    // 验证路由ID
    if !isValidObjectId(routeId) {
        return nil, domain.ErrInvalidRouteId
    }
    
    // 转换ID
    objectID, err := primitive.ObjectIDFromHex(routeId)
    if err != nil {
        return nil, err
    }
    
    // 获取数据
    data, err := uc.routeRepo.Get<RouteEntity>ForRoute(ctx, objectID.Hex())
    if err != nil {
        return nil, err
    }
    
    return data, nil
}

// validateRouteData 验证路由数据
func (uc *<RouteEntity>UsecaseImpl) validateRouteData(ctx context.Context, routeId string, data *scene_audio_route_models.<RouteModelEntity>) error {
    // 自定义验证逻辑
    // 例如：检查路由是否存在、数据是否与路由匹配等
    
    return nil
}

// isValidObjectId 验证ObjectID格式
func isValidObjectId(id string) bool {
    return primitive.IsValidObjectID(id)
}
```

### 音频路由UseCase接口实现
```
// 实现 domain.<RouteInterfaceEntity>Usecase 接口的所有方法
// 通过嵌入 usecase.BaseUsecase[<RouteModelEntity>] 实现基础CRUD业务逻辑

// 自定义音频路由业务逻辑方法
func (uc *<RouteEntity>UsecaseImpl) CreateWithRouteValidation(ctx context.Context, routeId string, data *scene_audio_route_models.<RouteModelEntity>) error {
    // 验证路由数据
    if err := uc.Validate<RouteEntity>DataForRoute(ctx, routeId, data); err != nil {
        return err
    }
    
    // 设置路由ID
    objectID, err := primitive.ObjectIDFromHex(routeId)
    if err != nil {
        return err
    }
    data.RouteID = objectID
    
    // 设置时间戳
    data.SetTimestamps()
    
    // 创建数据
    return uc.routeRepo.Create(ctx, data)
}

func (uc *<RouteEntity>UsecaseImpl) UpdateWithRouteValidation(ctx context.Context, routeId string, data *scene_audio_route_models.<RouteModelEntity>) error {
    // 验证路由数据
    if err := uc.Validate<RouteEntity>DataForRoute(ctx, routeId, data); err != nil {
        return err
    }
    
    // 更新数据
    return uc.routeRepo.UpdateByID(ctx, data.ID.Hex(), data)
}
```

## 代码生成规则

### UseCase结构规则
- 使用结构体嵌入实现基础功能
- 通过`usecase.BaseUsecase[<RouteModelEntity>]`实现通用业务逻辑
- 自定义方法实现音频路由专用业务逻辑

### 业务逻辑规则
- 音频路由数据创建和更新前进行验证
- 实现路由ID格式验证
- 支持路由关联数据处理
- 音频路由操作时进行权限检查

### 验证操作规则
- 路由ID格式验证
- 路由数据有效性验证
- 路由关联关系验证
- 音频路由数据格式验证

### 命名规则
- UseCase实现名: `<RouteEntity>UsecaseImpl`
- 方法名: 使用驼峰命名法
- 变量名: 使用描述性名称

## 依赖管理规范
- UseCase层依赖Domain层接口定义
- 依赖Repository层实现
- 依赖音频路由模型定义
- 通过构造函数注入Repository依赖

## 常用代码片段

### 音频路由数据验证方法
```go
func (uc *<Route>UsecaseImpl) validateRouteData(ctx context.Context, routeId string, data *scene_audio_route_models.<RouteModelEntity>) error {
    // 执行基本验证
    if err := data.Validate(); err != nil {
        return err
    }
    
    // 验证路由ID格式
    if !isValidObjectId(routeId) {
        return domain.ErrInvalidRouteId
    }
    
    // 自定义验证逻辑
    if data.<Field> == "" {
        return domain.ErrInvalid<FieldName>
    }
    
    return nil
}
```

### 音频路由权限检查方法
```go
func (uc *<Route>UsecaseImpl) checkRoutePermission(ctx context.Context, userID, routeId string) error {
    // 检查用户是否有权限访问或修改路由数据
    // 例如：检查用户角色、权限等
    
    return nil
}
```

### 音频路由数据统计方法
```go
func (uc *<Route>UsecaseImpl) GetRouteStats(ctx context.Context, routeId string) (map[string]interface{}, error) {
    // 获取路由统计信息
    stats, err := uc.routeRepo.Get<RouteEntity>StatsByRoute(ctx, routeId)
    if err != nil {
        return nil, err
    }
    
    return stats, nil
}
```

### 音频路由批量处理方法
```go
func (uc *<Route>UsecaseImpl) ProcessMultipleRouteData(ctx context.Context, routeId string, dataList []*scene_audio_route_models.<RouteModelEntity>) error {
    for _, data := range dataList {
        if err := uc.Process<RouteEntity>DataForRoute(ctx, routeId, data); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 音频路由数据同步方法
```go
func (uc *<Route>UsecaseImpl) SyncRouteData(ctx context.Context, routeId string) error {
    // 同步路由数据
    // 例如：同步外部数据源、更新缓存等
    
    return nil
}
```

### 音频路由健康检查方法
```go
func (uc *<Route>UsecaseImpl) CheckRouteHealth(ctx context.Context, routeId string) (bool, error) {
    // 检查路由是否健康
    // 例如：检查路由数据是否完整、访问权限等
    
    return true, nil
}
```

### 音频路由数据转换方法
```go
func (uc *<Route>UsecaseImpl) TransformRouteData(ctx context.Context, routeId string, transformFunc func(*scene_audio_route_models.<RouteModelEntity>) *scene_audio_route_models.<RouteModelEntity>) error {
    // 获取路由数据
    dataList, err := uc.Get<RouteEntity>ByRoute(ctx, routeId)
    if err != nil {
        return err
    }
    
    // 转换数据
    for _, data := range dataList {
        transformedData := transformFunc(data)
        if err := uc.routeRepo.UpdateByID(ctx, data.ID.Hex(), transformedData); err != nil {
            return err
        }
    }
    
    return nil
}
```