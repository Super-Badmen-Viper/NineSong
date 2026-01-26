# UseCase层AI代码生成规范

## 代码结构规范

### UseCase实现规范
- 文件路径: `usecase/usecase_<domain>/usecase_<subdomain>/<usecase_name>.go`
- 结构体命名: `<EntityName>Usecase struct`
- 构造函数: `New<EntityName>Usecase()`
- 必须包含Repository依赖和超时时间

### 依赖注入规范
- Repository依赖通过构造函数注入
- 超时参数通过构造函数注入
- 不允许在UseCase中直接创建Repository
- 支持多个Repository依赖注入

### 业务逻辑规范
- 业务规则验证在UseCase层实现
- 数据转换在UseCase层进行
- 事务管理在UseCase层控制
- 错误处理和日志记录

## 代码生成模板

### 新建UseCase实现
```
package <subdomain_name>

import (
    "context"
    "time"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_<domain>/domain_<subdomain>"
)

type <entityName>Usecase struct {
    repo    <subdomain>.<EntityName>Repository
    timeout time.Duration
}

func New<entityName>Usecase(repo <subdomain>.<EntityName>Repository, timeout time.Duration) <subdomain>.<EntityName>Usecase {
    return &<entityName>Usecase{
        repo:    repo,
        timeout: timeout,
    }
}
```

### UseCase方法实现模板
```
func (uc *<entityName>Usecase) <MethodName>(c context.Context, param string) (*<EntityName>, error) {
    ctx, cancel := context.WithTimeout(c, uc.timeout)
    defer cancel()
    
    // 业务验证
    if param == "" {
        return nil, errors.New("param is required")
    }
    
    // 业务逻辑实现
    result, err := uc.repo.<RepositoryMethod>(ctx, param)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

### 多Repository依赖UseCase实现
```
package <subdomain_name>

import (
    "context"
    "time"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_<domain>/domain_<subdomain>"
)

type <entityName>Usecase struct {
    repo1   <subdomain>.<EntityName1>Repository
    repo2   <subdomain>.<EntityName2>Repository
    timeout time.Duration
}

func New<entityName>Usecase(
    repo1 <subdomain>.<EntityName1>Repository, 
    repo2 <subdomain>.<EntityName2>Repository, 
    timeout time.Duration,
) <subdomain>.<EntityName>Usecase {
    return &<entityName>Usecase{
        repo1:   repo1,
        repo2:   repo2,
        timeout: timeout,
    }
}
```

## 代码生成规则

### 方法实现规则
- 所有方法使用`context.WithTimeout`控制超时
- 上下文超时时间从结构体字段获取
- 错误处理遵循Go错误处理规范
- 业务验证在UseCase层进行
- 返回结果前进行必要转换

### 依赖管理规则
- Repository依赖通过构造函数注入
- 不允许在方法内部创建新的依赖
- 使用接口类型进行依赖注入
- 支持多个依赖注入

### 业务逻辑规则
- 业务规则验证在UseCase层实现
- 数据转换在UseCase层进行
- 事务管理在UseCase层控制
- 业务逻辑不应包含基础设施细节

### 并发安全规则
- UseCase实例应该是并发安全的
- 避免在UseCase中存储可变状态
- 使用只读依赖确保线程安全

## 命名规范
- UseCase结构体: `<EntityName>Usecase`
- 构造函数: `New<EntityName>Usecase`
- 方法名: 遵循Go导出规范
- 变量命名: `uc` (usecase实例), `ctx` (上下文), `repo` (repository)

## 常用代码片段

### 创建实体方法
```go
func (uc *<entityName>Usecase) Create(ctx context.Context, entity *<EntityName>) error {
    ctx, cancel := context.WithTimeout(ctx, uc.timeout)
    defer cancel()
    
    // 业务验证
    if err := entity.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // 设置时间戳
    entity.SetTimestamps()
    
    // 调用Repository
    return uc.repo.Create(ctx, entity)
}
```

### 获取实体方法
```go
func (uc *<entityName>Usecase) GetByID(ctx context.Context, id string) (*<EntityName>, error) {
    ctx, cancel := context.WithTimeout(ctx, uc.timeout)
    defer cancel()
    
    // 验证ID格式
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, fmt.Errorf("invalid id format: %w", err)
    }
    
    // 调用Repository
    entity, err := uc.repo.GetByID(ctx, objectID.Hex())
    if err != nil {
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }
    
    return entity, nil
}
```

### 更新实体方法
```go
func (uc *<entityName>Usecase) Update(ctx context.Context, id string, updates map[string]interface{}) error {
    ctx, cancel := context.WithTimeout(ctx, uc.timeout)
    defer cancel()
    
    // 验证ID格式
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("invalid id format: %w", err)
    }
    
    // 验证更新内容
    if len(updates) == 0 {
        return errors.New("updates cannot be empty")
    }
    
    // 添加更新时间
    updates["updated_at"] = primitive.NewDateTimeFromTime(time.Now())
    
    // 调用Repository
    return uc.repo.UpdateByID(ctx, objectID.Hex(), updates)
}
```

### 删除实体方法
```go
func (uc *<entityName>Usecase) Delete(ctx context.Context, id string) error {
    ctx, cancel := context.WithTimeout(ctx, uc.timeout)
    defer cancel()
    
    // 验证ID格式
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("invalid id format: %w", err)
    }
    
    // 调用Repository
    return uc.repo.DeleteByID(ctx, objectID.Hex())
}
```

### 业务验证方法
```go
func (uc *<entityName>Usecase) validate<entityName>(ctx context.Context, entity *<EntityName>) error {
    // 业务验证逻辑
    if entity.<Field> == "" {
        return errors.New("<field> is required")
    }
    
    // 检查唯一性约束
    existing, err := uc.repo.FindBy<Field>(ctx, entity.<Field>)
    if err != nil && err != mongo.ErrNoDocuments {
        return fmt.Errorf("failed to check uniqueness: %w", err)
    }
    if existing != nil {
        return errors.New("<field> already exists")
    }
    
    return nil
}
```

### 分页查询方法
```go
func (uc *<entityName>Usecase) GetPaginated(ctx context.Context, page, pageSize int) ([]*<EntityName>, int64, error) {
    ctx, cancel := context.WithTimeout(ctx, uc.timeout)
    defer cancel()
    
    // 验证分页参数
    if page <= 0 {
        page = 1
    }
    if pageSize <= 0 {
        pageSize = 10
    }
    if pageSize > 100 {
        pageSize = 100 // 限制最大页面大小
    }
    
    // 调用Repository
    entities, total, err := uc.repo.GetPaginated(ctx, page, pageSize)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to get paginated entities: %w", err)
    }
    
    return entities, total, nil
}
```

### 复杂业务逻辑方法
```go
func (uc *<entityName>Usecase) ComplexBusinessOperation(ctx context.Context, param *<RequestType>) (*<ResponseType>, error) {
    ctx, cancel := context.WithTimeout(ctx, uc.timeout)
    defer cancel()
    
    // 业务验证
    if err := param.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 获取相关实体
    entities, err := uc.repo.FindByCondition(ctx, param.Condition)
    if err != nil {
        return nil, fmt.Errorf("failed to find entities: %w", err)
    }
    
    // 业务处理
    result := &<ResponseType>{}
    for _, entity := range entities {
        // 处理逻辑
        if entity.Status == "<EntityName>StatusActive" {
            result.Count++
        }
    }
    
    return result, nil
}
```

### 事务处理方法
```go
func (uc *<entityName>Usecase) TransactionalOperation(ctx context.Context, param *<TransactionType>) error {
    ctx, cancel := context.WithTimeout(ctx, uc.timeout)
    defer cancel()
    
    // 开始事务处理逻辑
    // 注意：这里需要根据实际的数据库驱动来实现事务
    // MongoDB示例
    session, err := uc.client.StartSession()
    if err != nil {
        return fmt.Errorf("failed to start session: %w", err)
    }
    defer session.EndSession(ctx)
    
    err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
        // 事务操作1
        if err := uc.repo1.Create(sessCtx, param.Entity1); err != nil {
            return nil, err
        }
        
        // 事务操作2
        if err := uc.repo2.Update(sessCtx, param.Entity2); err != nil {
            return nil, err
        }
        
        return nil, nil
    })
    
    if err != nil {
        return fmt.Errorf("transaction failed: %w", err)
    }
    
    return nil
}
```

### 事件驱动方法
```go
func (uc *<entityName>Usecase) ProcessWithEvent(ctx context.Context, entity *<EntityName>) error {
    ctx, cancel := context.WithTimeout(ctx, uc.timeout)
    defer cancel()
    
    // 业务逻辑处理
    if err := uc.repo.Create(ctx, entity); err != nil {
        return fmt.Errorf("failed to create entity: %w", err)
    }
    
    // 触发事件（例如：发送通知、更新缓存等）
    go func() {
        // 异步处理事件
        // 注意：在实际实现中需要考虑错误处理和重试机制
        _ = uc.handleEntityCreatedEvent(entity)
    }()
    
    return nil
}

func (uc *<entityName>Usecase) handleEntityCreatedEvent(entity *<EntityName>) error {
    // 事件处理逻辑
    return nil
}
```