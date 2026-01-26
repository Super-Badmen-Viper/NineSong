# Domain App Config AI代码生成规范

## 代码结构规范

### 实体定义规范
- 文件路径: `domain/domain_app/domain_app_config/<config_type>.go`
- 配置实体结构: 包含配置项、版本、时间戳等字段
- 使用`primitive.DateTime`作为时间字段类型
- ID字段使用`primitive.ObjectID`类型

### 接口定义规范
- Repository接口: `domain.ConfigRepository[<ConfigType>]`
- UseCase接口: `domain.ConfigUsecase[<ConfigType>]`
- 配置专用接口，支持获取单例和批量操作

## 代码生成模板

### 新建配置实体定义
```
package domain_app_config

import "go.mongodb.org/mongo-driver/bson/primitive"

type <ConfigType> struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    CreatedAt   primitive.DateTime `bson:"created_at,omitempty" json:"created_at"`
    UpdatedAt   primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at"`
    // 配置项字段
    <ConfigField> <FieldType> `bson:"<config_field>" json:"<config_field>"`
    Version     int            `bson:"version" json:"version"`
}

// Validate 验证配置实体字段的有效性
func (c *<ConfigType>) Validate() error {
    // 验证逻辑
    return nil
}

// SetTimestamps 设置时间戳
func (c *<ConfigType>) SetTimestamps() {
    now := primitive.NewDateTimeFromTime(time.Now())
    if c.ID.IsZero() {
        c.CreatedAt = now
    }
    c.UpdatedAt = now
}
```

### 配置Repository接口定义
```
package domain_app_config

import (
    "context"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
)

type <ConfigType>Repository interface {
    domain.ConfigRepository[<ConfigType>]
    // 添加配置专用方法
    GetDefault(ctx context.Context) (*<ConfigType>, error)
    UpdateWithVersion(ctx context.Context, config *<ConfigType>) error
}
```

### 配置UseCase接口定义
```
package domain_app_config

import (
    "context"
    "time"
)

type <ConfigType>Usecase interface {
    domain.ConfigUsecase[<ConfigType>]
    // 添加配置业务方法
    GetEffectiveConfig(ctx context.Context) (*<ConfigType>, error)
    UpdateConfigWithValidation(ctx context.Context, config *<ConfigType>) error
}
```

## 代码生成规则

### 实体字段规则
- ID字段必须为`primitive.ObjectID`类型
- 时间字段使用`primitive.DateTime`类型
- 配置版本字段用于控制配置更新
- 字段标签使用`bson`和`json`标签

### 配置操作规则
- 配置获取通常返回单例数据
- 配置更新需要版本控制
- 配置验证在UseCase层进行
- 支持配置的备份和恢复

### 命名规则
- 配置实体名: `<ConfigType>` (如AppConfig, ServerConfig)
- 配置字段: 使用描述性名称
- 接口名: `<ConfigType>Repository` 或 `<ConfigType>Usecase`

## 依赖管理规范
- Domain层不依赖其他业务层
- 只依赖基础库和类型定义
- 配置接口用于解耦业务层

## 常用代码片段

### 配置验证方法
```go
func (c *<Config>) Validate() error {
    if c.<Field> == "" {
        return errors.New("<field> is required")
    }
    
    if c.<NumberField> <= 0 {
        return errors.New("<number_field> must be greater than 0")
    }
    
    return nil
}
```

### 配置比较方法
```go
func (c *<Config>) Equals(other *<Config>) bool {
    if c == nil || other == nil {
        return c == other
    }
    
    return c.<Field> == other.<Field> &&
           c.Version == other.Version
}
```

### 配置默认值设置
```go
func NewDefault<ConfigType>() *<ConfigType> {
    return &<ConfigType>{
        <Field>: <DefaultValue>,
        Version: 1,
    }
}
```

### 配置安全方法
```go
func (c *<Config>) Sanitize() *<ConfigType> {
    // 创建副本并清理敏感信息
    sanitized := *c
    sanitized.PrivateField = "" // 移除敏感字段
    return &sanitized
}
```

### 配置转换方法
```go
type <ConfigType>DTO struct {
    Field    <FieldType> `json:"field"`
    Version  int         `json:"version"`
}

func (c *<Config>) ToDTO() <ConfigType>DTO {
    return <ConfigType>DTO{
        Field:   c.Field,
        Version: c.Version,
    }
}
```