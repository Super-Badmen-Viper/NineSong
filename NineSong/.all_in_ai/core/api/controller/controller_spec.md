# Controller层AI代码生成规范

## 代码结构规范

### Controller结构定义
- 文件路径: `api/controller/controller_<domain>/controller_<subdomain>/<controller_name>.go`
- 结构体命名: `<EntityName>Controller struct`
- 依赖注入: UseCase依赖通过字段注入
- 必须包含对应的UseCase接口

### HTTP响应规范
- 成功响应: 使用`domain.SuccessResponse`
- 错误响应: 使用`domain.ErrorResponse`
- 状态码: 遵循HTTP标准状态码
- 内容类型: 统一使用JSON格式

### 请求处理规范
- 参数验证: 在方法开始时进行验证
- 错误处理: 统一错误响应格式
- 认证信息: 从上下文获取用户信息

## 代码生成模板

### 新建Controller实现
```
package <subdomain_name>

import (
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_<domain>/domain_<subdomain>"
    "github.com/gin-gonic/gin"
)

type <EntityName>Controller struct {
    <entityName>Usecase <subdomain>.<EntityName>Usecase
}

func New<EntityName>Controller(uc <subdomain>.<EntityName>Usecase) *<EntityName>Controller {
    return &<EntityName>Controller{
        <entityName>Usecase: uc,
    }
}
```

### Controller方法实现模板
```
func (ctrl *<EntityName>Controller) <MethodName>(c *gin.Context) {
    // 参数绑定和验证
    var req <RequestType>
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
        return
    }

    // 调用UseCase
    result, err := ctrl.<entityName>Usecase.<UseCaseMethod>(c, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }

    // 返回成功响应
    c.JSON(http.StatusOK, result)
}
```

### 带认证的Controller方法模板
```
func (ctrl *<EntityName>Controller) <MethodName>WithAuth(c *gin.Context) {
    // 获取用户ID
    userID := c.GetString("x-user-id")
    
    // 参数绑定和验证
    var req <RequestType>
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
        return
    }

    // 调用UseCase
    result, err := ctrl.<entityName>Usecase.<UseCaseMethod>(c, userID, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }

    // 返回成功响应
    c.JSON(http.StatusOK, result)
}
```

## 代码生成规则

### 参数验证规则
- 使用Gin的`ShouldBind`方法进行参数绑定
- 验证失败返回`400 Bad Request`
- 在结构体中使用验证标签进行字段验证
- 自定义验证逻辑在方法中实现

### 错误处理规则
- 参数错误: `400 Bad Request`
- 未授权: `401 Unauthorized`
- 资源未找到: `404 Not Found`
- 服务器错误: `500 Internal Server Error`
- 请求超时: `408 Request Timeout`
- 资源冲突: `409 Conflict`

### 响应格式规则
- 统一使用JSON格式响应
- 错误响应包含`Message`字段
- 成功响应包含业务数据
- 分页响应包含分页信息

### 认证处理规则
- 从上下文获取用户ID
- 验证用户权限
- 处理认证失败情况

## 路由参数处理

### 路径参数
```
id := c.Param("id")
```

### 查询参数
```
param := c.Query("param")
defaultValue := c.DefaultQuery("param", "default")
```

### 表单参数
```
param := c.PostForm("param")
```

### JSON参数
```
var req <RequestType>
if err := c.ShouldBindJSON(&req); err != nil {
    // 处理错误
}
```

## 命名规范
- Controller结构体: `<EntityName>Controller`
- 构造函数: `New<EntityName>Controller`
- 控制器方法: 遵循RESTful命名规范
- 变量命名: `ctrl` (controller实例), `c` (gin context), `req` (请求对象)

## 常用代码片段

### GET请求处理
```go
func (ctrl *<EntityName>Controller) GetByID(c *gin.Context) {
    id := c.Param("id")
    
    entity, err := ctrl.<entityName>Usecase.GetByID(c, id)
    if err != nil {
        c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "Entity not found"})
        return
    }
    
    c.JSON(http.StatusOK, entity)
}
```

### POST请求处理
```go
func (ctrl *<EntityName>Controller) Create(c *gin.Context) {
    var entity <EntityName>
    if err := c.ShouldBindJSON(&entity); err != nil {
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    err := ctrl.<entityName>Usecase.Create(c, &entity)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, domain.SuccessResponse{Message: "Created successfully"})
}
```

### PUT请求处理
```go
func (ctrl *<EntityName>Controller) Update(c *gin.Context) {
    id := c.Param("id")
    
    var updates map[string]interface{}
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    err := ctrl.<entityName>Usecase.Update(c, id, updates)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, domain.SuccessResponse{Message: "Updated successfully"})
}
```

### DELETE请求处理
```go
func (ctrl *<EntityName>Controller) Delete(c *gin.Context) {
    id := c.Param("id")
    
    err := ctrl.<entityName>Usecase.Delete(c, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, domain.SuccessResponse{Message: "Deleted successfully"})
}
```

### 批量操作处理
```go
func (ctrl *<EntityName>Controller) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
    
    entities, total, err := ctrl.<entityName>Usecase.GetPaginated(c, page, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": entities,
        "total": total,
        "page": page,
        "page_size": pageSize,
        "has_more": int64(page*pageSize) < total,
    })
}
```

### 用户认证信息获取
```go
func (ctrl *<EntityName>Controller) MethodWithAuth(c *gin.Context) {
    userID := c.GetString("x-user-id")  // 从JWT中间件获取
    
    // 使用userID进行业务处理
    result, err := ctrl.<entityName>Usecase.MethodWithUser(c, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, result)
}
```

### 文件上传处理
```go
func (ctrl *<EntityName>Controller) UploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Failed to get file"})
        return
    }
    
    // 限制文件大小
    if file.Size > 10<<20 { // 10MB
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "File too large"})
        return
    }
    
    // 调用UseCase处理文件
    result, err := ctrl.<entityName>Usecase.HandleFileUpload(c, file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, result)
}
```

### 查询过滤处理
```go
func (ctrl *<EntityName>Controller) FilteredList(c *gin.Context) {
    // 获取查询参数
    field := c.Query("field")
    value := c.Query("value")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
    
    // 构建查询条件
    filters := map[string]interface{}{}
    if field != "" && value != "" {
        filters[field] = value
    }
    
    // 调用UseCase
    entities, total, err := ctrl.<entityName>Usecase.GetWithFilters(c, filters, page, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": entities,
        "total": total,
        "page": page,
        "page_size": pageSize,
    })
}
```

### 复杂验证处理
```go
func (ctrl *<EntityName>Controller) CreateWithValidation(c *gin.Context) {
    var req <EntityName>Request
    
    // 自定义验证逻辑
    if err := c.ShouldBindJSON(&req); err != nil {
        // 处理绑定错误
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    // 自定义验证
    if req.Field == "" {
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Field is required"})
        return
    }
    
    if req.NumberField <= 0 {
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "NumberField must be greater than 0"})
        return
    }
    
    // 调用UseCase
    result, err := ctrl.<entityName>Usecase.CreateWithValidation(c, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, result)
}
```

### 错误响应处理
```go
func (ctrl *<EntityName>Controller) handleErrorResponse(c *gin.Context, err error) {
    // 根据错误类型返回不同的HTTP状态码
    if errors.Is(err, ErrNotFound) {
        c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "Resource not found"})
        return
    }
    
    if errors.Is(err, ErrValidationFailed) {
        c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Validation failed"})
        return
    }
    
    if errors.Is(err, ErrUnauthorized) {
        c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Unauthorized"})
        return
    }
    
    // 默认服务器错误
    c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "Internal server error"})
}
```