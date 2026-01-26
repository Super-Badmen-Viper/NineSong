# Domain Scene Audio Route Interface AI代码生成规范

## 代码结构规范

### 接口定义规范
- 文件路径: `domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface/<interface_entity>.go`
- 路由接口定义: 定义音频场景路由相关接口
- 使用泛型定义通用路由操作接口
- 遵循领域驱动设计原则

### 实体定义规范
- 文件路径: `domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models/<model_entity>.go`
- 路由实体模型: 定义音频场景路由实体
- 使用`primitive.DateTime`作为时间字段类型
- ID字段使用`primitive.ObjectID`类型

## 代码生成模板

### 新建音频路由接口定义
```
package scene_audio_route_interface

import (
    "context"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
)

type <InterfaceEntity>Repository interface {
    domain.BaseRepository[<ModelEntity>]
    // 添加音频路由专用方法
    GetBy<FieldName>(ctx context.Context, <field_name> string) (*<ModelEntity>, error)
    GetBy<FieldName>WithPagination(ctx context.Context, <field_name> string, page, limit int) ([]*<ModelEntity>, error)
    Search<InterfaceEntity>(ctx context.Context, query string) ([]*<ModelEntity>, error)
    Get<InterfaceEntity>ForRoute(ctx context.Context, routeId string) ([]*<ModelEntity>, error)
}
```

### 新建音频路由实体模型
```
package scene_audio_route_models

import "go.mongodb.org/mongo-driver/bson/primitive"

type <ModelEntity> struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    CreatedAt   primitive.DateTime `bson:"created_at,omitempty" json:"created_at"`
    UpdatedAt   primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at"`
    // 音频路由实体字段
    <Field>     <FieldType>        `bson:"<field>" json:"<field>"`
    <Field2>    <FieldType2>       `bson:"<field2>" json:"<field2>"`
    RouteID     primitive.ObjectID `bson:"route_id" json:"route_id"`
}

// Validate 验证音频路由实体字段的有效性
func (e *<ModelEntity>) Validate() error {
    // 验证逻辑
    return nil
}

// SetTimestamps 设置时间戳
func (e *<ModelEntity>) SetTimestamps() {
    now := primitive.NewDateTimeFromTime(time.Now())
    if e.ID.IsZero() {
        e.CreatedAt = now
    }
    e.UpdatedAt = now
}
```

### 音频路由UseCase接口定义
```
package scene_audio_route_interface

import (
    "context"
    "time"
)

type <InterfaceEntity>Usecase interface {
    domain.BaseUsecase[<ModelEntity>]
    // 添加音频路由业务方法
    Process<InterfaceEntity>DataForRoute(ctx context.Context, routeId string, data *<ModelEntity>) error
    Validate<InterfaceEntity>DataForRoute(ctx context.Context, routeId string, data *<ModelEntity>) error
    Get<InterfaceEntity>WithRouteDetails(ctx context.Context, id string) (*<ModelEntity>, error)
    Get<InterfaceEntity>ByRoute(ctx context.Context, routeId string) ([]*<ModelEntity>, error)
}
```

## 代码生成规则

### 接口定义规则
- Repository接口继承`domain.BaseRepository[<ModelEntity>]`
- UseCase接口继承`domain.BaseUsecase[<ModelEntity>]`
- 使用泛型参数指定具体实体类型
- 接口方法名使用驼峰命名法

### 实体字段规则
- ID字段必须为`primitive.ObjectID`类型
- 时间字段使用`primitive.DateTime`类型
- 字段标签使用`bson`和`json`标签
- 路由相关字段需关联路由ID

### 音频路由操作规则
- 支持按路由ID查询音频数据
- 实现路由级别的分页功能
- 支持路由数据搜索
- 数据验证在UseCase层进行

### 命名规则
- 接口实体名: `<InterfaceEntity>` (如Album, Artist, MediaFile)
- 路由模型: `<ModelEntity>` (如AlbumRouteModel, ArtistRouteModel)
- 接口名: `<InterfaceEntity>Repository` 或 `<InterfaceEntity>Usecase`

## 依赖管理规范
- Domain层不依赖其他业务层
- 只依赖基础库和类型定义
- 音频路由接口用于解耦业务层

## 常用代码片段

### 音频路由实体验证方法
```go
func (e *<ModelEntity>) Validate() error {
    if e.<Field> == "" {
        return errors.New("<field> is required")
    }
    
    if e.RouteID.IsZero() {
        return errors.New("route id is required")
    }
    
    return nil
}
```

### 音频路由实体安全方法
```go
func (e *<ModelEntity>) Sanitize() *<ModelEntity> {
    // 创建副本并清理敏感信息
    sanitized := *e
    sanitized.PrivateField = "" // 移除敏感字段
    return &sanitized
}
```

### 音频路由实体转换方法
```go
type <ModelEntity>DTO struct {
    ID        string `json:"id"`
    <Field>   string `json:"<field>"`
    <Field2>  string `json:"<field2>"`
    RouteID   string `json:"route_id"`
}

func (e *<ModelEntity>) ToDTO() <ModelEntity>DTO {
    return <ModelEntity>DTO{
        ID:      e.ID.Hex(),
        <Field>: e.<Field>,
        <Field2>: e.<Field2>,
        RouteID: e.RouteID.Hex(),
    }
}
```

### 音频路由实体比较方法
```go
func (e *<ModelEntity>) Equals(other *<ModelEntity>) bool {
    if e == nil || other == nil {
        return e == other
    }
    
    return e.<Field> == other.<Field> &&
           e.RouteID == other.RouteID
}
```

### 音频路由实体辅助方法
```go
func (e *<ModelEntity>) BelongsToRoute(routeId primitive.ObjectID) bool {
    return e.RouteID == routeId
}

func (e *<ModelEntity>) IsNewerThan(other *<ModelEntity>) bool {
    if e == nil || other == nil {
        return false
    }
    return e.UpdatedAt > other.UpdatedAt
}
```

### 音频路由数据处理方法
```go
func (e *<ModelEntity>) ProcessRouteData() error {
    // 处理音频路由数据的逻辑
    e.<Field> = strings.TrimSpace(e.<Field>)
    return nil
}
```