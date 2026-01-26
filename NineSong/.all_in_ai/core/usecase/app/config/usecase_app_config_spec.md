# UseCase App Config AI代码生成规范

## 代码结构规范

### UseCase实现规范
- 文件路径: `usecase/usecase_app/usecase_app_config/<config_type>_usercase.go`
- 继承基础UseCase: `usecase.BaseUsecase[<ConfigType>]`
- 使用`base_usecase.go`作为基础实现
- 配置UseCase支持配置验证、版本控制和缓存机制

### 接口实现规范
- 实现`domain.ConfigUsecase[<ConfigType>]`接口
- 支持配置的业务逻辑处理
- 包含配置专用方法如获取有效配置、带验证的更新等

## 代码生成模板

### 新建配置UseCase实现
```
package usecase_app_config

import (
    "context"
    "time"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

type <ConfigType>UsecaseImpl struct {
    usecase.BaseUsecase[domain.<ConfigType>]
    configRepo domain.<ConfigType>Repository
}

func New<ConfigType>UsecaseImpl(configRepo domain.<ConfigType>Repository) domain.<ConfigType>Usecase {
    return &<ConfigType>UsecaseImpl{
        BaseUsecase: usecase.NewBaseUsecase[domain.<ConfigType>](configRepo),
        configRepo:  configRepo,
    }
}

// GetEffectiveConfig 获取有效配置
func (uc *<ConfigType>UsecaseImpl) GetEffectiveConfig(ctx context.Context) (*domain.<ConfigType>, error) {
    // 可以添加缓存逻辑
    config, err := uc.configRepo.GetDefault(ctx)
    if err != nil {
        return nil, err
    }
    
    return config, nil
}

// UpdateConfigWithValidation 带验证的配置更新
func (uc *<ConfigType>UsecaseImpl) UpdateConfigWithValidation(ctx context.Context, config *domain.<ConfigType>) error {
    // 业务逻辑验证
    if err := uc.validateConfig(ctx, config); err != nil {
        return err
    }
    
    // 更新配置
    if err := uc.configRepo.UpdateWithVersion(ctx, config); err != nil {
        return err
    }
    
    return nil
}

// validateConfig 配置验证逻辑
func (uc *<ConfigType>UsecaseImpl) validateConfig(ctx context.Context, config *domain.<ConfigType>) error {
    // 执行配置验证
    if err := config.Validate(); err != nil {
        return err
    }
    
    // 自定义验证逻辑
    // 例如：检查配置值是否在有效范围内
    
    return nil
}
```

### 配置UseCase接口实现
```
// 实现 domain.ConfigUsecase[<ConfigType>] 接口的所有方法
// 通过嵌入 usecase.BaseUsecase[<ConfigType>] 实现基础CRUD业务逻辑

// 自定义配置业务逻辑方法
func (uc *<ConfigType>UsecaseImpl) CreateWithValidation(ctx context.Context, config *domain.<ConfigType>) error {
    // 验证配置
    if err := uc.validateConfig(ctx, config); err != nil {
        return err
    }
    
    // 设置时间戳
    config.SetTimestamps()
    
    // 创建配置
    return uc.configRepo.Create(ctx, config)
}

func (uc *<ConfigType>UsecaseImpl) GetConfigByName(ctx context.Context, name string) (*domain.<ConfigType>, error) {
    // 业务逻辑处理
    config, err := uc.configRepo.GetByName(ctx, name)
    if err != nil {
        return nil, err
    }
    
    // 可以添加权限检查等业务逻辑
    
    return config, nil
}
```

## 代码生成规则

### UseCase结构规则
- 使用结构体嵌入实现基础功能
- 通过`usecase.BaseUsecase[<ConfigType>]`实现通用业务逻辑
- 自定义方法实现配置专用业务逻辑

### 业务逻辑规则
- 配置创建和更新前进行验证
- 实现配置的缓存机制
- 支持配置的版本控制
- 配置变更时触发相关业务逻辑

### 验证操作规则
- 配置值有效性验证
- 配置依赖关系验证
- 配置冲突检测
- 配置安全验证

### 命名规则
- UseCase实现名: `<ConfigType>UsecaseImpl`
- 方法名: 使用驼峰命名法
- 变量名: 使用描述性名称

## 依赖管理规范
- UseCase层依赖Domain层接口定义
- 依赖Repository层实现
- 通过构造函数注入Repository依赖

## 常用代码片段

### 配置验证方法
```go
func (uc *<Config>UsecaseImpl) validateConfig(ctx context.Context, config *domain.<Config>) error {
    // 执行基本验证
    if err := config.Validate(); err != nil {
        return err
    }
    
    // 自定义验证逻辑
    if config.<Field> != "" {
        // 验证字段值是否符合特定规则
        if !isValid<FieldValue>(config.<Field>) {
            return domain.ErrInvalid<Config>Value
        }
    }
    
    return nil
}

func isValid<FieldValue>(value string) bool {
    // 自定义验证逻辑
    return len(value) > 0 && len(value) < 100
}
```

### 配置缓存处理方法
```go
func (uc *<Config>UsecaseImpl) GetConfigWithCache(ctx context.Context, cacheKey string) (*domain.<Config>, error) {
    // 尝试从缓存获取
    // 如果缓存中没有，则从数据库获取并存入缓存
    config, err := uc.configRepo.GetDefault(ctx)
    if err != nil {
        return nil, err
    }
    
    return config, nil
}
```

### 配置变更通知方法
```go
func (uc *<Config>UsecaseImpl) NotifyConfigChange(ctx context.Context, config *domain.<Config>) error {
    // 配置变更后执行通知逻辑
    // 例如：发送事件、更新缓存、通知其他服务等
    
    return nil
}
```

### 配置备份方法
```go
func (uc *<Config>UsecaseImpl) BackupConfig(ctx context.Context, configID string) error {
    config, err := uc.configRepo.GetByID(ctx, configID)
    if err != nil {
        return err
    }
    
    // 执行备份逻辑
    // 例如：保存到备份表、外部存储等
    
    return nil
}
```

### 配置恢复方法
```go
func (uc *<Config>UsecaseImpl) RestoreConfig(ctx context.Context, backupID string) error {
    // 执行恢复逻辑
    // 例如：从备份中获取配置数据并更新
    
    return nil
}
```

### 配置权限检查方法
```go
func (uc *<Config>UsecaseImpl) CheckConfigPermission(ctx context.Context, userID string, config *domain.<Config>) error {
    // 检查用户是否有权限访问或修改配置
    // 例如：检查用户角色、权限等
    
    return nil
}
```

### 配置依赖检查方法
```go
func (uc *<Config>UsecaseImpl) CheckConfigDependencies(ctx context.Context, config *domain.<Config>) error {
    // 检查配置的依赖关系
    // 例如：检查其他配置是否满足当前配置的前置条件
    
    return nil
}
```