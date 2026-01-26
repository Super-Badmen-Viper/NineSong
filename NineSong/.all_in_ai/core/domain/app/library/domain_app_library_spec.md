# Domain App Library AI代码生成规范

## 代码结构规范

### 实体定义规范
- 文件路径: `domain/domain_app/domain_app_library/<library_entity>.go`
- 媒体库实体结构: 包含库信息、路径、类型、状态等字段
- 使用`primitive.DateTime`作为时间字段类型
- ID字段使用`primitive.ObjectID`类型

### 接口定义规范
- Repository接口: `domain.BaseRepository[<LibraryEntity>]`
- UseCase接口: `domain.BaseUsecase[<LibraryEntity>]`
- 媒体库专用接口，支持扫描、同步等操作

## 代码生成模板

### 新建媒体库实体定义
```
package domain_app_library

import "go.mongodb.org/mongo-driver/bson/primitive"

type <LibraryEntity> struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    CreatedAt   primitive.DateTime `bson:"created_at,omitempty" json:"created_at"`
    UpdatedAt   primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at"`
    // 媒体库字段
    Name        string             `bson:"name" json:"name"`
    Path        string             `bson:"path" json:"path"`
    Type        <LibraryType>      `bson:"type" json:"type"`
    Status      <LibraryStatus>    `bson:"status" json:"status"`
    LastScanned primitive.DateTime `bson:"last_scanned,omitempty" json:"last_scanned,omitempty"`
}

// Validate 验证媒体库实体字段的有效性
func (l *<LibraryEntity>) Validate() error {
    // 验证逻辑
    return nil
}

// SetTimestamps 设置时间戳
func (l *<LibraryEntity>) SetTimestamps() {
    now := primitive.NewDateTimeFromTime(time.Now())
    if l.ID.IsZero() {
        l.CreatedAt = now
    }
    l.UpdatedAt = now
}
```

### 媒体库Repository接口定义
```
package domain_app_library

import (
    "context"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
)

type <LibraryEntity>Repository interface {
    domain.BaseRepository[<LibraryEntity>]
    // 添加媒体库专用方法
    GetByPath(ctx context.Context, path string) (*<LibraryEntity>, error)
    UpdateScanStatus(ctx context.Context, id primitive.ObjectID, status <LibraryStatus>) error
    GetActiveLibraries(ctx context.Context) ([]*<LibraryEntity>, error)
}
```

### 媒体库UseCase接口定义
```
package domain_app_library

import (
    "context"
    "time"
)

type <LibraryEntity>Usecase interface {
    domain.BaseUsecase[<LibraryEntity>]
    // 添加媒体库业务方法
    ScanLibrary(ctx context.Context, id string) error
    SyncLibrary(ctx context.Context, id string) error
    ValidateLibraryPath(ctx context.Context, path string) error
}
```

## 代码生成规则

### 实体字段规则
- ID字段必须为`primitive.ObjectID`类型
- 时间字段使用`primitive.DateTime`类型
- 路径字段验证文件系统路径有效性
- 状态字段使用枚举类型

### 媒体库操作规则
- 路径验证确保文件系统访问权限
- 扫描操作支持异步处理
- 库状态管理确保数据一致性
- 支持多种媒体库类型

### 命名规则
- 媒体库实体名: `<LibraryEntity>` (如MediaFileLibrary)
- 媒体库字段: 使用描述性名称
- 接口名: `<LibraryEntity>Repository` 或 `<LibraryEntity>Usecase`

## 依赖管理规范
- Domain层不依赖其他业务层
- 只依赖基础库和类型定义
- 媒体库接口用于解耦业务层

## 常用代码片段

### 媒体库验证方法
```go
func (l *<Library>) Validate() error {
    if l.Name == "" {
        return errors.New("library name is required")
    }
    
    if l.Path == "" {
        return errors.New("library path is required")
    }
    
    // 验证路径是否有效
    if !filepath.IsAbs(l.Path) {
        return errors.New("library path must be absolute")
    }
    
    return nil
}
```

### 媒体库状态枚举
```go
type LibraryStatus string

const (
    LibraryStatusActive   LibraryStatus = "active"
    LibraryStatusInactive LibraryStatus = "inactive"
    LibraryStatusScanning LibraryStatus = "scanning"
    LibraryStatusError    LibraryStatus = "error"
)

func (s LibraryStatus) IsValid() bool {
    switch s {
    case LibraryStatusActive, LibraryStatusInactive, LibraryStatusScanning, LibraryStatusError:
        return true
    default:
        return false
    }
}
```

### 媒体库类型枚举
```go
type LibraryType string

const (
    LibraryTypeAudio  LibraryType = "audio"
    LibraryTypeVideo  LibraryType = "video"
    LibraryTypeImage  LibraryType = "image"
    LibraryTypeOther  LibraryType = "other"
)

func (t LibraryType) IsValid() bool {
    switch t {
    case LibraryTypeAudio, LibraryTypeVideo, LibraryTypeImage, LibraryTypeOther:
        return true
    default:
        return false
    }
}
```

### 媒体库安全方法
```go
func (l *<Library>) Sanitize() *<LibraryEntity> {
    // 创建副本并清理敏感信息
    sanitized := *l
    // 不返回路径或其他敏感信息
    return &sanitized
}
```

### 媒体库辅助方法
```go
func (l *<Library>) IsActive() bool {
    return l.Status == LibraryStatusActive
}

func (l *<Library>) IsScanning() bool {
    return l.Status == LibraryStatusScanning
}
```

### 媒体库转换方法
```go
type <LibraryEntity>DTO struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Path        string        `json:"path"`
    Type        LibraryType   `json:"type"`
    Status      LibraryStatus `json:"status"`
    LastScanned int64         `json:"last_scanned"`
}

func (l *<Library>) ToDTO() <LibraryEntity>DTO {
    return <LibraryEntity>DTO{
        ID:          l.ID.Hex(),
        Name:        l.Name,
        Path:        l.Path,
        Type:        l.Type,
        Status:      l.Status,
        LastScanned: int64(l.LastScanned),
    }
}
```