# UseCase App Library AI代码生成规范

## 代码结构规范

### UseCase实现规范
- 文件路径: `usecase/usecase_app/usecase_app_library/<library_entity>_usercase.go`
- 继承基础UseCase: `usecase.BaseUsecase[<LibraryEntity>]`
- 使用`base_usecase.go`作为基础实现
- 媒体库UseCase支持路径验证、扫描状态管理和文件同步

### 接口实现规范
- 实现`domain.BaseUsecase[<LibraryEntity>]`接口
- 支持媒体库的业务逻辑处理
- 包含媒体库专用方法如扫描库、同步库、验证路径等

## 代码生成模板

### 新建媒体库UseCase实现
```
package usecase_app_library

import (
    "context"
    "time"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

type <LibraryEntity>UsecaseImpl struct {
    usecase.BaseUsecase[domain.<LibraryEntity>]
    libraryRepo domain.<LibraryEntity>Repository
}

func New<LibraryEntity>UsecaseImpl(libraryRepo domain.<LibraryEntity>Repository) domain.<LibraryEntity>Usecase {
    return &<LibraryEntity>UsecaseImpl{
        BaseUsecase: usecase.NewBaseUsecase[domain.<LibraryEntity>](libraryRepo),
        libraryRepo: libraryRepo,
    }
}

// ScanLibrary 扫描媒体库
func (uc *<LibraryEntity>UsecaseImpl) ScanLibrary(ctx context.Context, id string) error {
    // 获取媒体库信息
    library, err := uc.libraryRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }
    
    // 更新扫描状态
    if err := uc.libraryRepo.UpdateScanStatus(ctx, library.ID, domain.LibraryStatusScanning); err != nil {
        return err
    }
    
    // 执行扫描逻辑
    if err := uc.performLibraryScan(ctx, library); err != nil {
        // 扫描失败，更新状态
        uc.libraryRepo.UpdateScanStatus(ctx, library.ID, domain.LibraryStatusError)
        return err
    }
    
    // 扫描成功，更新状态
    if err := uc.libraryRepo.UpdateScanStatus(ctx, library.ID, domain.LibraryStatusActive); err != nil {
        return err
    }
    
    return nil
}

// SyncLibrary 同步媒体库
func (uc *<LibraryEntity>UsecaseImpl) SyncLibrary(ctx context.Context, id string) error {
    // 获取媒体库信息
    library, err := uc.libraryRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }
    
    // 执行同步逻辑
    if err := uc.performLibrarySync(ctx, library); err != nil {
        return err
    }
    
    return nil
}

// ValidateLibraryPath 验证媒体库路径
func (uc *<LibraryEntity>UsecaseImpl) ValidateLibraryPath(ctx context.Context, path string) error {
    // 验证路径格式
    if err := uc.validateLibraryPathFormat(path); err != nil {
        return err
    }
    
    // 检查路径是否已存在
    if err := uc.checkPathUniqueness(ctx, path); err != nil {
        return err
    }
    
    return nil
}

// performLibraryScan 执行媒体库扫描
func (uc *<LibraryEntity>UsecaseImpl) performLibraryScan(ctx context.Context, library *domain.<LibraryEntity>) error {
    // 实现具体的扫描逻辑
    // 例如：遍历目录、解析媒体文件、更新数据库等
    
    return nil
}

// performLibrarySync 执行媒体库同步
func (uc *<LibraryEntity>UsecaseImpl) performLibrarySync(ctx context.Context, library *domain.<LibraryEntity>) error {
    // 实现具体的同步逻辑
    // 例如：比较文件列表、同步新增/修改/删除的文件等
    
    return nil
}

// validateLibraryPathFormat 验证媒体库路径格式
func (uc *<LibraryEntity>UsecaseImpl) validateLibraryPathFormat(path string) error {
    if path == "" {
        return domain.ErrInvalidPath
    }
    
    // 检查路径是否为绝对路径
    if !filepath.IsAbs(path) {
        return domain.ErrInvalidPath
    }
    
    return nil
}

// checkPathUniqueness 检查路径唯一性
func (uc *<LibraryEntity>UsecaseImpl) checkPathUniqueness(ctx context.Context, path string) error {
    isUnique, err := uc.libraryRepo.IsPathUnique(ctx, path, primitive.NilObjectID)
    if err != nil {
        return err
    }
    
    if !isUnique {
        return domain.ErrPathAlreadyExists
    }
    
    return nil
}
```

### 媒体库UseCase接口实现
```
// 实现 domain.BaseUsecase[<LibraryEntity>] 接口的所有方法
// 通过嵌入 usecase.BaseUsecase[<LibraryEntity>] 实现基础CRUD业务逻辑

// 自定义媒体库业务逻辑方法
func (uc *<LibraryEntity>UsecaseImpl) CreateWithValidation(ctx context.Context, library *domain.<LibraryEntity>) error {
    // 验证媒体库
    if err := uc.validateLibrary(ctx, library); err != nil {
        return err
    }
    
    // 设置时间戳
    library.SetTimestamps()
    
    // 创建媒体库
    return uc.libraryRepo.Create(ctx, library)
}

func (uc *<LibraryEntity>UsecaseImpl) GetActiveLibraries(ctx context.Context) ([]*domain.<LibraryEntity>, error) {
    // 获取活跃的媒体库
    libraries, err := uc.libraryRepo.GetActiveLibraries(ctx)
    if err != nil {
        return nil, err
    }
    
    // 可以添加额外的业务逻辑处理
    
    return libraries, nil
}
```

## 代码生成规则

### UseCase结构规则
- 使用结构体嵌入实现基础功能
- 通过`usecase.BaseUsecase[<LibraryEntity>]`实现通用业务逻辑
- 自定义方法实现媒体库专用业务逻辑

### 业务逻辑规则
- 媒体库创建和更新前进行路径验证
- 实现媒体库扫描和同步逻辑
- 支持扫描状态管理
- 媒体库操作时进行权限检查

### 验证操作规则
- 媒体库路径格式验证
- 路径唯一性验证
- 媒体库类型验证
- 媒体库状态验证

### 命名规则
- UseCase实现名: `<LibraryEntity>UsecaseImpl`
- 方法名: 使用驼峰命名法
- 变量名: 使用描述性名称

## 依赖管理规范
- UseCase层依赖Domain层接口定义
- 依赖Repository层实现
- 通过构造函数注入Repository依赖

## 常用代码片段

### 媒体库验证方法
```go
func (uc *<Library>UsecaseImpl) validateLibrary(ctx context.Context, library *domain.<Library>) error {
    // 执行基本验证
    if err := library.Validate(); err != nil {
        return err
    }
    
    // 验证路径格式
    if err := uc.validateLibraryPathFormat(library.Path); err != nil {
        return err
    }
    
    // 检查路径唯一性
    isUnique, err := uc.libraryRepo.IsPathUnique(ctx, library.Path, library.ID)
    if err != nil {
        return err
    }
    
    if !isUnique {
        return domain.ErrPathAlreadyExists
    }
    
    return nil
}
```

### 媒体库权限检查方法
```go
func (uc *<Library>UsecaseImpl) checkLibraryPermission(ctx context.Context, userID string, library *domain.<Library>) error {
    // 检查用户是否有权限访问或修改媒体库
    // 例如：检查用户角色、权限等
    
    return nil
}
```

### 媒体库文件统计方法
```go
func (uc *<Library>UsecaseImpl) GetLibraryStats(ctx context.Context, id string) (map[string]interface{}, error) {
    // 获取媒体库统计信息
    stats, err := uc.libraryRepo.GetLibraryStats(ctx, id)
    if err != nil {
        return nil, err
    }
    
    return stats, nil
}
```

### 媒体库异步扫描方法
```go
func (uc *<Library>UsecaseImpl) ScanLibraryAsync(ctx context.Context, id string) error {
    go func() {
        // 在goroutine中执行扫描操作
        if err := uc.ScanLibrary(ctx, id); err != nil {
            // 处理扫描错误
            log.Printf("Error scanning library %s: %v", id, err)
        }
    }()
    
    return nil
}
```

### 媒体库批量操作方法
```go
func (uc *<Library>UsecaseImpl) UpdateMultipleLibraries(ctx context.Context, updates []map[string]interface{}) error {
    for _, update := range updates {
        // 获取库ID
        idStr, ok := update["id"].(string)
        if !ok {
            continue
        }
        
        // 执行更新操作
        // 这里可以根据需要实现具体逻辑
    }
    
    return nil
}
```

### 媒体库健康检查方法
```go
func (uc *<Library>UsecaseImpl) CheckLibraryHealth(ctx context.Context, id string) (bool, error) {
    // 检查媒体库路径是否存在且可访问
    library, err := uc.libraryRepo.GetByID(ctx, id)
    if err != nil {
        return false, err
    }
    
    // 执行健康检查逻辑
    // 例如：检查路径是否可访问、磁盘空间等
    
    return true, nil
}
```

### 媒体库文件操作方法
```go
func (uc *<Library>UsecaseImpl) ProcessLibraryFiles(ctx context.Context, libraryID string, processFunc func(string) error) error {
    // 获取媒体库信息
    library, err := uc.libraryRepo.GetByID(ctx, libraryID)
    if err != nil {
        return err
    }
    
    // 遍历库中的文件并执行处理函数
    // 这里可以根据需要实现具体逻辑
    
    return nil
}
```