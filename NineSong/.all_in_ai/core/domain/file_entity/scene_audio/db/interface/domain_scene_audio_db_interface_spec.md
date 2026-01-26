# Domain Scene Audio DB Interface AI代码生成规范

## 代码结构规范

### 接口定义规范
- 文件路径: `domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface/<interface_entity>.go`
- 数据库接口定义: 定义音频场景数据库相关接口
- 使用泛型定义通用数据库操作接口
- 遵循领域驱动设计原则

### 实体定义规范
- 文件路径: `domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models/<model_entity>.go`
- 数据库实体模型: 定义音频场景数据库实体
- 使用`primitive.DateTime`作为时间字段类型
- ID字段使用`primitive.ObjectID`类型

## 代码生成模板

### 新建音频数据库接口定义
```
package scene_audio_db_interface

import (
    "context"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
)

type <InterfaceEntity>Repository interface {
    domain.BaseRepository[<ModelEntity>]
    // 添加音频数据库专用方法
    GetBy<FieldName>(ctx context.Context, <field_name> string) (*<ModelEntity>, error)
    GetBy<FieldName>WithPagination(ctx context.Context, <field_name> string, page, limit int) ([]*<ModelEntity>, error)
    Search<InterfaceEntity>(ctx context.Context, query string) ([]*<ModelEntity>, error)
}
```

### 新建音频数据库实体模型
```
package scene_audio_db_models

import "go.mongodb.org/mongo-driver/bson/primitive"

type <ModelEntity> struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    CreatedAt   primitive.DateTime `bson:"created_at,omitempty" json:"created_at"`
    UpdatedAt   primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at"`
    // 音频实体字段
    <Field>     <FieldType>        `bson:"<field>" json:"<field>"`
    <Field2>    <FieldType2>       `bson:"<field2>" json:"<field2>"`
}

// Validate 验证音频实体字段的有效性
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

### 音频数据库UseCase接口定义
```
package scene_audio_db_interface

import (
    "context"
    "time"
)

type <InterfaceEntity>Usecase interface {
    domain.BaseUsecase[<ModelEntity>]
    // 添加音频数据库业务方法
    Process<InterfaceEntity>Data(ctx context.Context, data *<ModelEntity>) error
    Validate<InterfaceEntity>Data(ctx context.Context, data *<ModelEntity>) error
    Get<InterfaceEntity>WithDetails(ctx context.Context, id string) (*<ModelEntity>, error)
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
- 音频相关字段需考虑大小写敏感性

### 音频操作规则
- 支持按多种条件查询音频数据
- 实现分页功能
- 支持音频数据搜索
- 数据验证在UseCase层进行

### 命名规则
- 接口实体名: `<InterfaceEntity>` (如Album, Artist, MediaFile)
- 数据库模型: `<ModelEntity>` (如AlbumModel, ArtistModel)
- 接口名: `<InterfaceEntity>Repository` 或 `<InterfaceEntity>Usecase`

## 依赖管理规范
- Domain层不依赖其他业务层
- 只依赖基础库和类型定义
- 音频数据库接口用于解耦业务层

## 常用代码片段

### 音频实体验证方法
```go
func (e *<ModelEntity>) Validate() error {
    if e.<Field> == "" {
        return errors.New("<field> is required")
    }
    
    if e.<Field2> == "" {
        return errors.New("<field2> is required")
    }
    
    return nil
}
```

### 音频实体安全方法
```go
func (e *<ModelEntity>) Sanitize() *<ModelEntity> {
    // 创建副本并清理敏感信息
    sanitized := *e
    sanitized.PrivateField = "" // 移除敏感字段
    return &sanitized
}
```

### 音频实体转换方法
```go
type <ModelEntity>DTO struct {
    ID        string `json:"id"`
    <Field>   string `json:"<field>"`
    <Field2>  string `json:"<field2>"`
}

func (e *<ModelEntity>) ToDTO() <ModelEntity>DTO {
    return <ModelEntity>DTO{
        ID:       e.ID.Hex(),
        <Field>:  e.<Field>,
        <Field2>: e.<Field2>,
    }
}
```

### 音频实体比较方法
```go
func (e *<ModelEntity>) Equals(other *<ModelEntity>) bool {
    if e == nil || other == nil {
        return e == other
    }
    
    return e.<Field> == other.<Field> &&
           e.<Field2> == other.<Field2>
}
```

### 音频实体辅助方法
```go
func (e *<ModelEntity>) IsNewerThan(other *<ModelEntity>) bool {
    if e == nil || other == nil {
        return false
    }
    return e.UpdatedAt > other.UpdatedAt
}
```

### 音频数据处理方法
```go
func (e *<ModelEntity>) ProcessData() error {
    // 处理音频数据的逻辑
    e.<Field> = strings.TrimSpace(e.<Field>)
    return nil
}
```