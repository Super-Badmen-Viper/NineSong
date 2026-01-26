# Middleware层AI代码生成规范

## 代码结构规范

### 中间件函数规范
- 文件路径: `api/middleware/middleware_<domain>/<middleware_name>.go`
- 函数签名: `func <MiddlewareName>(param string) gin.HandlerFunc`
- 返回类型: `gin.HandlerFunc`
- 必须遵循Gin中间件模式

### 中间件实现规范
- 使用Gin框架的中间件机制
- 通过`c.Next()`调用下一个处理函数
- 错误处理后使用`c.Abort()`中断请求链
- 保证中间件的幂等性和无状态性

### 中间件分类规范
- 认证中间件: 验证用户身份
- 授权中间件: 验证用户权限
- 限流中间件: 控制请求频率
- 日志中间件: 记录请求信息
- 验证中间件: 验证请求参数

## 代码生成模板

### 基础中间件实现
```
package middleware_<domain>

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
)

func <MiddlewareName>(param string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 中间件逻辑
        
        if <condition> {
            c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Not authorized"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### JWT认证中间件实现
```
func JwtAuthMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 令牌获取逻辑
        authToken := c.Query("access_token")
        
        if authToken == "" {
            authHeader := c.GetHeader("Authorization")
            if authHeader == "" {
                c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Not authorized"})
                c.Abort()
                return
            }
            
            tokenParts := strings.Split(authHeader, " ")
            if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
                c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Invalid authorization format"})
                c.Abort()
                return
            }
            authToken = tokenParts[1]
        }
        
        // 令牌验证逻辑
        authorized, err := token_util.IsAuthorized(authToken, secret)
        if !authorized || err != nil {
            c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Invalid token"})
            c.Abort()
            return
        }
        
        // 提取用户信息
        userID, err := token_util.ExtractIDFromToken(authToken, secret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: err.Error()})
            c.Abort()
            return
        }
        
        // 设置上下文信息
        c.Set("x-user-id", userID)
        c.Next()
    }
}
```

### 带配置的中间件实现
```
type MiddlewareConfig struct {
    Secret    string
    Whitelist []string
    Blacklist []string
}

func <ConfigurableMiddleware>(config MiddlewareConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 使用配置进行处理
        
        c.Next()
    }
}
```

## 代码生成规则

### 中间件设计规则
- 中间件函数返回`gin.HandlerFunc`类型
- 使用`c.Abort()`中断请求处理链
- 使用`c.Next()`调用下一个处理函数
- 错误处理后必须中断请求链
- 避免在中间件中存储可变状态

### 错误处理规则
- 认证失败: 返回`401 Unauthorized`
- 格式错误: 返回`400 Bad Request`
- 服务器错误: 返回`500 Internal Server Error`
- 请求过多: 返回`429 Too Many Requests`
- 禁止访问: 返回`403 Forbidden`

### 上下文传递规则
- 通过`c.Set()`设置上下文值
- 通过`c.Get()`获取上下文值
- 常用键名: `"x-user-id"`, `"user-role"`等
- 避免在上下文中存储敏感信息

### 性能优化规则
- 最小化中间件中的计算操作
- 避免阻塞操作
- 合理使用缓存
- 减少不必要的数据库查询

## 命名规范
- 中间件函数: `<Feature>Middleware`
- 中间件参数: `secret`, `config`, `options`等
- 上下文键名: 使用`x-`前缀的命名规范
- 配置结构体: `<MiddlewareName>Config`

## 常用代码片段

### 基础认证中间件
```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 认证逻辑
        c.Next()
    }
}
```

### 权限验证中间件
```go
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("x-user-id")
        
        // 权限验证逻辑
        if !hasPermission(userID, requiredPermission) {
            c.JSON(http.StatusForbidden, domain.ErrorResponse{Message: "Insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 日志记录中间件
```go
func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 记录请求开始
        startTime := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery
        
        c.Next()
        
        // 记录请求结束
        endTime := time.Now()
        latency := endTime.Sub(startTime)
        clientIP := c.ClientIP()
        method := c.Request.Method
        statusCode := c.Writer.Status()
        
        if raw != "" {
            path = path + "?" + raw
        }
        
        log.Printf("[%v] %s %s %d %v %s", 
            endTime.Format("2006/01/02 - 15:04:05"), 
            clientIP, 
            method, 
            statusCode, 
            latency, 
            path,
        )
    }
}
```

### 限流中间件
```go
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
    // 使用令牌桶或其他算法实现限流
    // 这里使用一个简单的计数器实现
    requests := make(map[string]*int64)
    mutex := &sync.RWMutex{}
    
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        mutex.Lock()
        if requests[clientIP] == nil {
            counter := int64(0)
            requests[clientIP] = &counter
        }
        
        counter := requests[clientIP]
        *counter++
        
        current := *counter
        mutex.Unlock()
        
        if current > int64(limit) {
            c.JSON(http.StatusTooManyRequests, domain.ErrorResponse{Message: "Rate limit exceeded"})
            c.Abort()
            return
        }
        
        // 重置计数器的goroutine
        go func() {
            time.Sleep(window)
            mutex.Lock()
            if requests[clientIP] != nil {
                *requests[clientIP] = 0
            }
            mutex.Unlock()
        }()
        
        c.Next()
    }
}
```

### CORS中间件
```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

### 请求大小限制中间件
```go
func MaxSizeMiddleware(maxSize int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查Content-Length头
        contentLength := c.Request.ContentLength
        if contentLength > maxSize {
            c.JSON(http.StatusRequestEntityTooLarge, domain.ErrorResponse{Message: "Request body too large"})
            c.Abort()
            return
        }
        
        // 限制Multipart forms的内存使用
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
        
        c.Next()
    }
}
```

### 请求ID中间件
```go
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := generateRequestID() // 实现生成唯一请求ID的函数
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        
        c.Next()
    }
}

func generateRequestID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return fmt.Sprintf("%x", b)
}
```

### 认证可选中间件
```go
func OptionalAuthMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            // 没有提供认证信息，继续执行但不设置用户ID
            c.Next()
            return
        }
        
        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
            c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Invalid authorization format"})
            c.Abort()
            return
        }
        
        authToken := tokenParts[1]
        
        // 令牌验证逻辑
        authorized, err := token_util.IsAuthorized(authToken, secret)
        if !authorized || err != nil {
            c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Invalid token"})
            c.Abort()
            return
        }
        
        // 提取用户信息
        userID, err := token_util.ExtractIDFromToken(authToken, secret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: err.Error()})
            c.Abort()
            return
        }
        
        // 设置上下文信息
        c.Set("x-user-id", userID)
        c.Next()
    }
}
```

### 多角色授权中间件
```go
func RoleAuthMiddleware(allowedRoles []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("x-user-id")
        
        // 获取用户角色
        userRoles, err := getUserRoles(userID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "Failed to get user roles"})
            c.Abort()
            return
        }
        
        // 检查用户角色是否在允许列表中
        hasPermission := false
        for _, userRole := range userRoles {
            for _, allowedRole := range allowedRoles {
                if userRole == allowedRole {
                    hasPermission = true
                    break
                }
            }
            if hasPermission {
                break
            }
        }
        
        if !hasPermission {
            c.JSON(http.StatusForbidden, domain.ErrorResponse{Message: "Insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```