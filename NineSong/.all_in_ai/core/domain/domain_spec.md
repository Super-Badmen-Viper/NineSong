# Domain层AI代码生成规范

## 代码结构规范

### 实体定义规范
- 文件路径: `domain/domain_<domain>/domain_<subdomain>/<entity_name>.go`
- 实体结构: 使用`primitive.DateTime`作为时间字段类型
- ID字段: 使用`primitive.ObjectID`作为ID类型
- 字段标签: 必须包含bson和json标签

### 接口定义规范
- Repository接口: `domain.BaseRepository[T]` 或 `domain.ConfigRepository[T]`
- UseCase接口: `domain.BaseUsecase[T]` 或 `domain.ConfigUsecase[T]`
- 自定义接口: 根据业务需求定义特定方法

### 枚举类型规范
- 文件路径: `domain/domain_<domain>/domain_<subdomain>/types_<entity_name>.go`
- 使用常量定义枚举值
- 提供验证函数确保值的有效性

## 代码生成模板

### 新建实体定义
```
package domain_<subdomain>

import "go.mongodb.org/mongo-driver/bson/primitive"

type <EntityName> struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    CreatedAt   primitive.DateTime `bson:"created_at,omitempty" json:"created_at"`
    UpdatedAt   primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at"`
    // 添加其他字段
    <FieldName> <FieldType> `bson:"<field_name>" json:"<field_name>"`
    <PrivateField> <FieldType> `bson:"<private_field_name>,omitempty" json:"<private_field_name>,omitempty"`
}

// Validate 验证实体字段的有效性
func (e *<EntityName>) Validate() error {
    // 验证逻辑
    return nil
}

// SetTimestamps 设置时间戳
func (e *<EntityName>) SetTimestamps() {
    now := primitive.NewDateTimeFromTime(time.Now())
    if e.ID.IsZero() {
        e.CreatedAt = now
    }
    e.UpdatedAt = now
}
```

### Repository接口定义
```
package domain_<subdomain>

import (
    "context"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
)

type <EntityName>Repository interface {
    domain.BaseRepository[<EntityName>]
    // 添加自定义方法
    CustomMethod(ctx context.Context, param string) (*<EntityName>, error)
    FindBy<Field>(ctx context.Context, fieldVal string) ([]*<EntityName>, error)
}
```

### UseCase接口定义
```
package domain_<subdomain>

import (
    "context"
    "time"
)

type <EntityName>Usecase interface {
    domain.BaseUsecase[<EntityName>]
    // 添加自定义业务方法
    CustomBusinessMethod(ctx context.Context, param string) error
    BusinessMethodWithValidation(ctx context.Context, entity *<EntityName>) error
}
```

### 枚举类型定义
```
package domain_<subdomain>

// <EntityName>Status 定义<EntityName>状态枚举
type <EntityName>Status string

const (
    <EntityName>StatusActive   <EntityName>Status = "active"
    <EntityName>StatusInactive <EntityName>Status = "inactive"
    <EntityName>StatusPending  <EntityName>Status = "pending"
)

// IsValid 验证状态值是否有效
func (s <EntityName>Status) IsValid() bool {
    switch s {
    case <EntityName>StatusActive, <EntityName>StatusInactive, <EntityName>StatusPending:
        return true
    default:
        return false
    }
}
```

## 代码生成规则

### 实体字段规则
- ID字段必须为`primitive.ObjectID`类型
- 时间字段使用`primitive.DateTime`类型
- 字段标签使用`bson`和`json`标签
- 敏感字段添加`json:"-"`标签
- 可选字段使用`omitempty`标签

### 接口方法规则
- 所有方法必须接收`context.Context`作为第一个参数
- 错误处理使用`error`作为最后一个返回值
- 返回实体时使用指针类型
- 方法名应清晰表达其功能

### 命名规则
- 实体名: 首字母大写的驼峰命名
- 字段名: 首字母大写的驼峰命名
- 接口名: `<EntityName>Repository` 或 `<EntityName>Usecase`
- 方法名: 首字母小写的驼峰命名，动词+名词形式

### 验证规则
- 提供Validate方法验证实体字段
- 对于枚举类型提供IsValid方法
- 在UseCase层进行业务逻辑验证

## 依赖管理规范
- Domain层不依赖其他业务层
- 只依赖基础库和类型定义
- 接口定义用于解耦业务层
- 避免导入实现细节

## 常用代码片段

### 实体验证方法
```go
func (e *<Entity>) Validate() error {
    if e.<FieldName> == "" {
        return errors.New("<field_name> is required")
    }
    
    if e.<NumberField> <= 0 {
        return errors.New("<number_field> must be greater than 0")
    }
    
    return nil
}
```

### 实体辅助方法
```go
func (e *<Entity>) SetTimestamps() {
    now := primitive.NewDateTimeFromTime(time.Now())
    if e.ID.IsZero() {
        e.CreatedAt = now
    }
    e.UpdatedAt = now
}

func (e *<Entity>) IsNew() bool {
    return e.ID.IsZero()
}
```

### 接口默认实现
```go
type base<EntityName>Usecase struct {
    repo <EntityName>Repository
    timeout time.Duration
}

func NewBase<EntityName>Usecase(repo <EntityName>Repository, timeout time.Duration) *base<EntityName>Usecase {
    return &base<EntityName>Usecase{
        repo:    repo,
        timeout: timeout,
    }
}
```

### 实体比较方法
```go
func (e *<Entity>) Equals(other *<Entity>) bool {
    if e == nil || other == nil {
        return e == other
    }
    
    return e.ID == other.ID &&
           e.<FieldName> == other.<FieldName> &&
           e.<OtherField> == other.<OtherField>
}
```

### 实体转换方法
```go
type <EntityName>DTO struct {
    ID        string `json:"id"`
    FieldName string `json:"field_name"`
}

func (e *<Entity>) ToDTO() <EntityName>DTO {
    return <EntityName>DTO{
        ID:        e.ID.Hex(),
        FieldName: e.FieldName,
    }
}
```

### 实体克隆方法
```go
func (e *<Entity>) Clone() *<Entity> {
    if e == nil {
        return nil
    }
    
    clone := *e
    return &clone
}
```

### 实体安全方法
```go
func (e *<Entity>) Sanitize() {
    // 移除或清理敏感信息
    e.PrivateField = ""
}
```