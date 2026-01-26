# UseCase Auth AI代码生成规范

## 代码结构规范

### UseCase实现规范
- 文件路径: `usecase/usecase_auth/<auth_feature>_usecase.go`
- 继承基础UseCase: `usecase.BaseUsecase[<AuthEntity>]`
- 使用`base_usecase.go`作为基础实现
- 认证UseCase支持用户验证、JWT令牌生成和权限管理

### 接口实现规范
- 实现`domain.<AuthFeature>Usecase`接口
- 支持认证的业务逻辑处理
- 包含认证专用方法如创建访问令牌、验证用户等

## 代码生成模板

### 新建登录UseCase实现
```
package usecase_auth

import (
    "context"
    "time"
    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v4"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

type LoginUsecaseImpl struct {
    userRepo domain.UserRepository
}

func NewLoginUsecaseImpl(userRepo domain.UserRepository) domain.LoginUsecase {
    return &LoginUsecaseImpl{
        userRepo: userRepo,
    }
}

// GetUserByEmail 根据邮箱获取用户
func (uc *LoginUsecaseImpl) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
    user, err := uc.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return user, err
    }
    
    return user, nil
}

// CreateAccessToken 创建访问令牌
func (uc *LoginUsecaseImpl) CreateAccessToken(user *domain.User, secret string, expiry int) (accessToken string, err error) {
    claims := &domain.JwtCustomClaims{
        ID:    user.ID.Hex(),
        Name:  user.Name,
        Email: user.Email,
        Admin: user.Admin,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Second)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "NineSong",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    accessToken, err = token.SignedString([]byte(secret))
    if err != nil {
        return "", err
    }

    return accessToken, nil
}

// CreateRefreshToken 创建刷新令牌
func (uc *LoginUsecaseImpl) CreateRefreshToken(user *domain.User, secret string, expiry int) (refreshToken string, err error) {
    claims := &domain.JwtCustomClaims{
        ID:    user.ID.Hex(),
        Name:  user.Name,
        Email: user.Email,
        Admin: user.Admin,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Second)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "NineSong",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    refreshToken, err = token.SignedString([]byte(secret))
    if err != nil {
        return "", err
    }

    return refreshToken, nil
}
```

### 新建注册UseCase实现
```
package usecase_auth

import (
    "context"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "golang.org/x/crypto/bcrypt"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
)

type SignupUsecaseImpl struct {
    userRepo domain.UserRepository
}

func NewSignupUsecaseImpl(userRepo domain.UserRepository) domain.SignupUsecase {
    return &SignupUsecaseImpl{
        userRepo: userRepo,
    }
}

// Create 创建用户
func (uc *SignupUsecaseImpl) Create(ctx context.Context, user *domain.User) error {
    // 验证用户
    if err := uc.validateUser(ctx, user); err != nil {
        return err
    }
    
    // 检查邮箱是否已存在
    isEmailTaken, err := uc.userRepo.IsEmailTaken(ctx, user.Email, primitive.NilObjectID)
    if err != nil {
        return err
    }
    
    if isEmailTaken {
        return domain.ErrEmailAlreadyExists
    }
    
    // 加密密码
    if err := user.HashPassword(); err != nil {
        return err
    }
    
    // 设置时间戳
    user.SetTimestamps()
    
    // 创建用户
    return uc.userRepo.Create(ctx, user)
}

// GetUserByEmail 根据邮箱获取用户
func (uc *SignupUsecaseImpl) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
    user, err := uc.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return user, err
    }
    
    return user, nil
}

// CreateAccessToken 创建访问令牌
func (uc *SignupUsecaseImpl) CreateAccessToken(user *domain.User, secret string, expiry int) (accessToken string, err error) {
    // 实现访问令牌创建逻辑
    claims := &domain.JwtCustomClaims{
        ID:    user.ID.Hex(),
        Name:  user.Name,
        Email: user.Email,
        Admin: user.Admin,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Second)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "NineSong",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    accessToken, err = token.SignedString([]byte(secret))
    if err != nil {
        return "", err
    }

    return accessToken, nil
}

// CreateRefreshToken 创建刷新令牌
func (uc *SignupUsecaseImpl) CreateRefreshToken(user *domain.User, secret string, expiry int) (refreshToken string, err error) {
    // 实现刷新令牌创建逻辑
    claims := &domain.JwtCustomClaims{
        ID:    user.ID.Hex(),
        Name:  user.Name,
        Email: user.Email,
        Admin: user.Admin,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Second)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "NineSong",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    refreshToken, err = token.SignedString([]byte(secret))
    if err != nil {
        return "", err
    }

    return refreshToken, nil
}

// validateUser 验证用户
func (uc *SignupUsecaseImpl) validateUser(ctx context.Context, user *domain.User) error {
    // 执行基本验证
    if err := user.Validate(); err != nil {
        return err
    }
    
    // 自定义验证逻辑
    // 例如：检查密码强度、邮箱格式等
    
    return nil
}
```

### 认证UseCase接口实现
```
// 实现 domain.LoginUsecase 和 domain.SignupUsecase 接口的所有方法
// 包含用户认证相关的业务逻辑

// 自定义认证业务逻辑方法
func (uc *LoginUsecaseImpl) AuthenticateUser(ctx context.Context, email, password string) (domain.User, string, string, error) {
    // 获取用户
    user, err := uc.GetUserByEmail(ctx, email)
    if err != nil {
        return user, "", "", err
    }
    
    // 验证密码
    if !user.CheckPassword(password) {
        return user, "", "", domain.ErrInvalidCredentials
    }
    
    // 生成令牌
    accessToken, err := uc.CreateAccessToken(&user, "access_secret", 3600)
    if err != nil {
        return user, "", "", err
    }
    
    refreshToken, err := uc.CreateRefreshToken(&user, "refresh_secret", 86400)
    if err != nil {
        return user, "", "", err
    }
    
    return user, accessToken, refreshToken, nil
}
```

## 代码生成规则

### UseCase结构规则
- 使用结构体嵌入实现基础功能
- 通过`usecase.BaseUsecase[<AuthEntity>]`实现通用业务逻辑
- 自定义方法实现认证专用业务逻辑

### 业务逻辑规则
- 用户创建和登录前进行验证
- 实现密码加密和验证
- 支持JWT令牌生成和验证
- 用户操作时进行权限检查

### 验证操作规则
- 用户邮箱格式验证
- 密码强度验证
- 用户权限验证
- 令牌有效性验证

### 命名规则
- UseCase实现名: `<AuthFeature>UsecaseImpl`
- 方法名: 使用驼峰命名法
- 变量名: 使用描述性名称

## 依赖管理规范
- UseCase层依赖Domain层接口定义
- 依赖Repository层实现
- 依赖JWT和密码加密库
- 通过构造函数注入Repository依赖

## 常用代码片段

### 用户验证方法
```go
func (uc *<Auth>UsecaseImpl) validateUser(ctx context.Context, user *domain.User) error {
    // 执行基本验证
    if err := user.Validate(); err != nil {
        return err
    }
    
    // 自定义验证逻辑
    if len(user.Password) < 6 {
        return domain.ErrWeakPassword
    }
    
    return nil
}
```

### 密码加密方法
```go
func (uc *<Auth>UsecaseImpl) encryptPassword(password string) (string, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    
    return string(hashedPassword), nil
}
```

### 令牌验证方法
```go
func (uc *<Auth>UsecaseImpl) ValidateToken(ctx context.Context, tokenString, secret string) (*domain.JwtCustomClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &domain.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*domain.JwtCustomClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, domain.ErrInvalidToken
}
```

### 用户权限检查方法
```go
func (uc *<Auth>UsecaseImpl) CheckUserPermission(ctx context.Context, userID string, requiredPermission string) error {
    // 检查用户是否有权限执行特定操作
    user, err := uc.userRepo.GetByID(ctx, userID)
    if err != nil {
        return err
    }
    
    // 检查用户权限
    if !user.IsAdmin() {
        return domain.ErrUnauthorized
    }
    
    return nil
}
```

### 用户活动记录方法
```go
func (uc *<Auth>UsecaseImpl) RecordUserActivity(ctx context.Context, userID string, activity string) error {
    // 记录用户活动
    // 例如：登录时间、操作记录等
    
    return nil
}
```

### 用户状态更新方法
```go
func (uc *<Auth>UsecaseImpl) UpdateUserStatus(ctx context.Context, userID string, status string) error {
    // 更新用户状态
    user, err := uc.userRepo.GetByID(ctx, userID)
    if err != nil {
        return err
    }
    
    // 更新用户状态
    // 这里可以根据需要实现具体逻辑
    
    return nil
}
```

### 用户安全验证方法
```go
func (uc *<Auth>UsecaseImpl) PerformSecurityCheck(ctx context.Context, email string) error {
    // 执行安全检查
    // 例如：检查是否在黑名单中、IP限制等
    
    return nil
}
```