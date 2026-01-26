# Route层AI代码生成规范

## 代码结构规范

### 路由函数规范
- 文件路径: `api/route/route_<domain>/route_<subdomain>/<route_name>.go`
- 函数命名: `New<EntityName>Router`
- 函数签名: `(timeout time.Duration, db mongo.Database, group *gin.RouterGroup)`
- 必须注入所有依赖项

### 路由分组规范
- 公共路由: 不需要认证的API
- 私有路由: 需要JWT认证的API
- 路由组: 使用Gin的路由组功能组织相关API
- 嵌套路由: 支持多级路由分组

### 路由配置规范
- 依赖注入顺序: Repository -> UseCase -> Controller
- 路由注册顺序: 按资源逻辑分组
- 中间件应用: 在路由组级别应用

## 代码生成模板

### 基础路由实现
```
package <subdomain_name>

import (
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_<domain>/controller_<subdomain>"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_<domain>/repository_<subdomain>"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_<domain>/usecase_<subdomain>"
    "github.com/gin-gonic/gin"
    "time"
)

func New<EntityName>Router(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository_<subdomain>.New<EntityName>Repository(db, domain.Collection<EntityName>)
    uc := usecase_<subdomain>.New<EntityName>Usecase(repo, timeout)
    ctrl := controller_<subdomain>.New<EntityName>Controller(uc)
    
    // 注册路由
    group.GET("/<entity_name>", ctrl.Get)
    group.POST("/<entity_name>", ctrl.Create)
    group.PUT("/<entity_name>/:id", ctrl.Update)
    group.DELETE("/<entity_name>/:id", ctrl.Delete)
}
```

### 带分组的路由实现
```
func New<EntityName>Router(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository_<subdomain>.New<EntityName>Repository(db, domain.Collection<EntityName>)
    uc := usecase_<subdomain>.New<EntityName>Usecase(repo, timeout)
    ctrl := controller_<subdomain>.New<EntityName>Controller(uc)
    
    <entityName>Group := group.Group("/<entity_names>")
    {
        <entityName>Group.GET("", ctrl.Get)
        <entityName>Group.POST("", ctrl.Create)
        <entityName>Group.PUT("/:id", ctrl.Update)
        <entityName>Group.DELETE("/:id", ctrl.Delete)
        <entityName>Group.GET("/:id", ctrl.GetByID)
    }
}
```

### 带中间件的路由实现
```
func New<EntityName>Router(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository_<subdomain>.New<EntityName>Repository(db, domain.Collection<EntityName>)
    uc := usecase_<subdomain>.New<EntityName>Usecase(repo, timeout)
    ctrl := controller_<subdomain>.New<EntityName>Controller(uc)
    
    <entityName>Group := group.Group("/<entity_names>")
    <entityName>Group.Use(middleware_<domain>.<MiddlewareName>()) // 应用中间件
    {
        <entityName>Group.GET("", ctrl.Get)
        <entityName>Group.POST("", ctrl.Create)
        <entityName>Group.PUT("/:id", ctrl.Update)
        <entityName>Group.DELETE("/:id", ctrl.Delete)
    }
}
```

### 复杂路由实现（多个Repository依赖）
```
func New<EntityName>Router(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    // 创建多个Repository实例
    repo1 := repository_<subdomain>.New<EntityName1>Repository(db, domain.Collection<EntityName1>)
    repo2 := repository_<subdomain>.New<EntityName2>Repository(db, domain.Collection<EntityName2>)
    
    // 创建UseCase实例（使用多个Repository）
    uc := usecase_<subdomain>.New<EntityName>Usecase(repo1, repo2, timeout)
    
    // 创建Controller实例
    ctrl := controller_<subdomain>.New<EntityName>Controller(uc)
    
    // 注册路由
    <entityName>Group := group.Group("/<entity_names>")
    {
        <entityName>Group.GET("", ctrl.Get)
        <entityName>Group.POST("", ctrl.Create)
        <entityName>Group.PUT("/:id", ctrl.Update)
        <entityName>Group.DELETE("/:id", ctrl.Delete)
    }
}
```

## 代码生成规则

### 依赖注入规则
- 按Repository -> UseCase -> Controller顺序注入依赖
- 通过函数参数传递timeout, db, group
- 使用domain.Collection常量定义集合名称
- 依赖注入使用接口类型而非具体实现

### 路由注册规则
- GET: 获取资源
- POST: 创建资源
- PUT: 更新资源
- DELETE: 删除资源
- PATCH: 部分更新资源
- 使用RESTful风格的URL路径

### 错误处理规则
- 路由函数中的错误应被适当处理
- 数据库连接错误在上层处理
- 依赖注入失败时panic或返回错误
- 路由配置错误应被记录

### 中间件应用规则
- 在路由组级别应用中间件
- 认证中间件用于私有API
- 日志中间件用于所有API
- 限流中间件用于公共API

## 命名规范
- 路由函数: `New<EntityName>Router`
- 路由组变量: `<entityName>Group`
- URL路径: 使用复数形式`/<entity_names>`
- 集合名称: 使用`domain.Collection<EntityName>`
- 路由组前缀: 使用功能相关的前缀

## 常用代码片段

### 简单路由注册
```go
func NewSimpleRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    // 依赖注入
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    // 路由注册
    group.GET("/endpoint", ctrl.Handler)
}
```

### 带中间件的路由
```go
func NewProtectedRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    // 依赖注入
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    // 使用中间件的路由组
    protectedGroup := group.Group("/protected")
    protectedGroup.Use(middleware.JwtAuthMiddleware())
    {
        protectedGroup.GET("/data", ctrl.Handler)
    }
}
```

### 复杂路由组
```go
func NewComplexRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    entityGroup := group.Group("/entities")
    {
        entityGroup.GET("", ctrl.List)
        entityGroup.GET("/:id", ctrl.GetByID)
        entityGroup.POST("", ctrl.Create)
        entityGroup.PUT("/:id", ctrl.Update)
        entityGroup.DELETE("/:id", ctrl.Delete)
        
        // 子资源路由
        entityGroup.GET("/:id/subresource", ctrl.GetSubResource)
        entityGroup.POST("/:id/subresource", ctrl.CreateSubResource)
    }
}
```

### 批量操作路由
```go
func NewBatchRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    group.POST("/batch/create", ctrl.BatchCreate)
    group.PUT("/batch/update", ctrl.BatchUpdate)
    group.DELETE("/batch/delete", ctrl.BatchDelete)
}
```

### 查询路由
```go
func NewQueryRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    group.GET("/search", ctrl.Search)
    group.GET("/filter", ctrl.Filter)
    group.GET("/aggregate", ctrl.Aggregate)
}
```

### 文件上传路由
```go
func NewFileRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    group.POST("/upload", ctrl.UploadFile)
    group.GET("/download/:id", ctrl.DownloadFile)
}
```

### 嵌套路由
```go
func NewNestedRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    // 主资源
    mainRepo := repository.New<MainRepo>(db, domain.CollectionMain)
    mainUc := usecase.New<MainUseCase>(mainRepo, timeout)
    mainCtrl := controller.New<MainController>(mainUc)
    
    // 子资源
    subRepo := repository.New<SubRepo>(db, domain.CollectionSub)
    subUc := usecase.New<SubUseCase>(subRepo, timeout)
    subCtrl := controller.New<SubController>(subUc)
    
    // 主资源路由组
    mainGroup := group.Group("/mains")
    {
        mainGroup.GET("", mainCtrl.List)
        mainGroup.POST("", mainCtrl.Create)
        
        // 嵌套子资源路由组
        subGroup := mainGroup.Group("/:mainId/subs")
        {
            subGroup.GET("", subCtrl.GetByMainId)
            subGroup.POST("", subCtrl.CreateForMain)
            subGroup.PUT("/:id", subCtrl.UpdateForMain)
            subGroup.DELETE("/:id", subCtrl.DeleteForMain)
        }
    }
}
```

### 版本化路由
```go
func NewVersionedRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    // V1版本路由
    v1Group := group.Group("/api/v1/entities")
    {
        v1Group.GET("", ctrl.GetV1)
        v1Group.POST("", ctrl.CreateV1)
    }
    
    // V2版本路由
    v2Group := group.Group("/api/v2/entities")
    {
        v2Group.GET("", ctrl.GetV2)
        v2Group.POST("", ctrl.CreateV2)
    }
}
```

### 带验证的路由
```go
func NewValidatedRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    validatedGroup := group.Group("/entities")
    validatedGroup.Use(middleware.ValidationMiddleware()) // 应用验证中间件
    {
        validatedGroup.GET("", ctrl.Get)
        validatedGroup.POST("", ctrl.Create)
        validatedGroup.PUT("/:id", ctrl.Update)
    }
}
```

### API文档路由
```go
func NewDocumentedRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
    repo := repository.New<RepoName>(db, domain.CollectionName)
    uc := usecase.New<UseCaseName>(repo, timeout)
    ctrl := controller.New<ControllerName>(uc)
    
    // API文档路由
    group.GET("/docs", ctrl.GetDocs)
    group.GET("/docs/swagger", ctrl.GetSwagger)
    
    // 主要API路由
    apiGroup := group.Group("/api")
    {
        apiGroup.GET("/entities", ctrl.Get)
        apiGroup.POST("/entities", ctrl.Create)
    }
}
```