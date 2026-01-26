# Domain Auth AI代码生成规范

## 代码结构规范

### 实体定义规范
- 文件路径: `domain/domain_auth/<auth_entity>.go`
- 认证实体结构: 包含用户、任务、登录等认证相关实体
- 使用`primitive.DateTime`作为时间字段类型
- ID字段使用`primitive.ObjectID`类型

### 接口定义规范
- Repository接口: `domain.BaseRepository[<AuthEntity>]`
- UseCase接口: `domain.BaseUsecase[<AuthEntity>]`
- 认证专用接口，支持登录、注册、权限等操作

## 代码生成模板

### 新建用户实体定义
```
package domain_auth

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    CreatedAt primitive.DateTime `bson:"created_at,omitempty" json:"created_at"`
    UpdatedAt primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at"`
    // 用户字段
    Name      string             `bson:"name" json:"name"`
    Email     string             `bson:"email" json:"email"`
    Password  string             `bson:"password" json:"password"`
    Admin     bool               `bson:"admin" json:"admin"`
}

// Validate 验证用户实体字段的有效性
func (u *User) Validate() error {
    // 验证逻辑
    return nil
}

// SetTimestamps 设置时间戳
func (u *User) SetTimestamps() {
    now := primitive.NewDateTimeFromTime(time.Now())
    if u.ID.IsZero() {
        u.CreatedAt = now
    }
    u.UpdatedAt = now
}
```

### 登录请求/响应结构
```
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

type SignupRequest struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type SignupResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}
```

### 认证Repository接口定义
```
package domain_auth

import (
    "context"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
)

type UserRepository interface {
    domain.BaseRepository[User]
    // 添加认证专用方法
    GetByEmail(ctx context.Context, email string) (User, error)
    IsEmailTaken(ctx context.Context, email string, skipID primitive.ObjectID) (bool, error)
}
```

### 认证UseCase接口定义
```
package domain_auth

import (
    "context"
    "time"
)

type LoginUsecase interface {
    GetUserByEmail(c context.Context, email string) (User, error)
    CreateAccessToken(user *User, secret string, expiry int) (accessToken string, err error)
    CreateRefreshToken(user *User, secret string, expiry int) (refreshToken string, err error)
}

type SignupUsecase interface {
    Create(c context.Context, user *User) error
    GetUserByEmail(c context.Context, email string) (User, error)
    CreateAccessToken(user *User, secret string, expiry int) (accessToken string, err error)
    CreateRefreshToken(user *User, secret string, expiry int) (refreshToken string, err error)
}
```

## 代码生成规则

### 实体字段规则
- ID字段必须为`primitive.ObjectID`类型
- 时间字段使用`primitive.DateTime`类型
- 密码字段需要加密处理
- 邮箱字段需要格式验证

### 认证操作规则
- 用户邮箱唯一性验证
- 密码加密存储
- JWT令牌生成和验证
- 用户权限管理

### 命名规则
- 认证实体名: `User`, `Task`, `Profile`
- 认证字段: 使用描述性名称
- 接口名: `<AuthFeature>Usecase` (如LoginUsecase, SignupUsecase)

## 依赖管理规范
- Domain层不依赖其他业务层
- 只依赖基础库和类型定义
- 认证接口用于解耦业务层

## 常用代码片段

### 用户验证方法
```go
func (u *User) Validate() error {
    if u.Name == "" {
        return errors.New("user name is required")
    }
    
    if u.Email == "" {
        return errors.New("user email is required")
    }
    
    if !isValidEmail(u.Email) {
        return errors.New("invalid email format")
    }
    
    if u.Password == "" {
        return errors.New("password is required")
    }
    
    if len(u.Password) < 6 {
        return errors.New("password must be at least 6 characters")
    }
    
    return nil
}

func isValidEmail(email string) bool {
    // 简单的邮箱格式验证
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}
```

### 用户安全方法
```go
func (u *User) Sanitize() *User {
    // 创建副本并移除敏感信息
    sanitized := *u
    sanitized.Password = "" // 不返回密码
    return &sanitized
}
```

### 密码处理方法
```go
func (u *User) HashPassword() error {
    if u.Password == "" {
        return errors.New("password is empty")
    }
    
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    
    u.Password = string(hashedPassword)
    return nil
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
```

### 用户权限方法
```go
func (u *User) IsAdmin() bool {
    return u.Admin
}

func (u *User) CanAccessResource(resource string) bool {
    // 根据用户权限判断是否可以访问资源
    return u.Admin // 管理员可以访问所有资源
}
```

### 认证错误类型
```go
var (
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserNotFound       = errors.New("user not found")
    ErrUnauthorized       = errors.New("unauthorized")
)
```

### Profile结构
```go
type Profile struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```