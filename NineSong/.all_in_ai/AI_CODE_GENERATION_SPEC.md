# NineSong项目AI代码生成规范总览

## 规范概述
本文档为AI代码生成提供详细规范，涵盖项目所有层次的代码生成规则、模板和最佳实践。

## 层次规范文档

### 1. [MongoDB层AI代码生成规范](./core/mongo/mongo_spec.md)
- **功能**: 定义MongoDB接口和实现的生成规范
- **核心内容**:
  - 接口定义规范
  - 实现类规范
  - 代码生成模板
  - 依赖注入规范

### 2. [Repository层AI代码生成规范](./core/repository/repository_spec.md)
- **功能**: 定义数据访问层的生成规范
- **核心内容**:
  - 基础Repository实现模板
  - 配置Repository实现模板
  - 方法生成规则
  - 命名规范

#### 2.1 [App Config Repository规范](./core/repository/app/config/repository_app_config_spec.md)
- **功能**: 定义应用配置数据访问层的生成规范
- **核心内容**: 配置实体的CRUD操作、版本控制、单例模式实现

#### 2.2 [App Library Repository规范](./core/repository/app/library/repository_app_library_spec.md)
- **功能**: 定义媒体库数据访问层的生成规范
- **核心内容**: 媒体库实体的CRUD操作、路径验证、扫描状态管理

#### 2.3 [Auth Repository规范](./core/repository/auth/repository_auth_spec.md)
- **功能**: 定义认证数据访问层的生成规范
- **核心内容**: 用户实体的CRUD操作、邮箱唯一性验证、密码安全存储

#### 2.4 [Scene Audio DB Repository规范](./core/repository/file_entity/scene_audio/db/repository_scene_audio_db_spec.md)
- **功能**: 定义音频数据库数据访问层的生成规范
- **核心内容**: 音频数据实体的CRUD操作、复杂查询、关联数据处理

#### 2.5 [Scene Audio Route Repository规范](./core/repository/file_entity/scene_audio/route/repository_scene_audio_route_spec.md)
- **功能**: 定义音频路由数据访问层的生成规范
- **核心内容**: 音频路由实体的CRUD操作、路由关联数据、复杂查询

### 3. [Domain层AI代码生成规范](./core/domain/domain_spec.md)
- **功能**: 定义业务实体和接口的生成规范
- **核心内容**:
  - 实体定义模板
  - 接口定义模板
  - 实体字段规则
  - 依赖管理规范

#### 3.1 [App Config Domain规范](./core/domain/app/config/domain_app_config_spec.md)
- **功能**: 定义应用配置业务实体和接口的生成规范
- **核心内容**: 配置实体定义、配置专用接口、版本控制机制

#### 3.2 [App Library Domain规范](./core/domain/app/library/domain_app_library_spec.md)
- **功能**: 定义媒体库业务实体和接口的生成规范
- **核心内容**: 媒体库实体定义、媒体库专用接口、路径验证机制

#### 3.3 [Auth Domain规范](./core/domain/auth/domain_auth_spec.md)
- **功能**: 定义认证业务实体和接口的生成规范
- **核心内容**: 用户实体定义、认证专用接口、密码处理机制

#### 3.4 [Scene Audio DB Interface Domain规范](./core/domain/file_entity/scene_audio/db/interface/domain_scene_audio_db_interface_spec.md)
- **功能**: 定义音频数据库接口的生成规范
- **核心内容**: 音频数据接口定义、数据库实体模型、查询接口

#### 3.5 [Scene Audio Route Interface Domain规范](./core/domain/file_entity/scene_audio/route/interface/domain_scene_audio_route_interface_spec.md)
- **功能**: 定义音频路由接口的生成规范
- **核心内容**: 音频路由接口定义、路由实体模型、路由查询接口

### 4. [UseCase层AI代码生成规范](./core/usecase/usecase_spec.md)
- **功能**: 定义业务逻辑层的生成规范
- **核心内容**:
  - UseCase实现模板
  - 方法实现规则
  - 依赖注入规范
  - 常用代码片段

#### 4.1 [App Config UseCase规范](./core/usecase/app/config/usecase_app_config_spec.md)
- **功能**: 定义应用配置业务逻辑层的生成规范
- **核心内容**: 配置业务逻辑、验证机制、版本控制、缓存机制

#### 4.2 [App Library UseCase规范](./core/usecase/app/library/usecase_app_library_spec.md)
- **功能**: 定义媒体库业务逻辑层的生成规范
- **核心内容**: 媒体库扫描逻辑、同步逻辑、路径验证、状态管理

#### 4.3 [Auth UseCase规范](./core/usecase/auth/usecase_auth_spec.md)
- **功能**: 定义认证业务逻辑层的生成规范
- **核心内容**: 用户认证逻辑、JWT令牌生成、密码加密、权限验证

#### 4.4 [Scene Audio Route UseCase规范](./core/usecase/file_entity/scene_audio/route/usecase_scene_audio_route_spec.md)
- **功能**: 定义音频路由业务逻辑层的生成规范
- **核心内容**: 音频路由数据处理、路由关联逻辑、验证机制

### 5. [Controller层AI代码生成规范](./core/api/controller/controller_spec.md)
- **功能**: 定义接口控制层的生成规范
- **核心内容**:
  - Controller结构定义
  - HTTP响应规范
  - 参数验证规则
  - 常用请求处理模板

### 6. [Middleware层AI代码生成规范](./core/api/middleware/middleware_spec.md)
- **功能**: 定义中间件层的生成规范
- **核心内容**:
  - 中间件函数规范
  - 认证中间件实现
  - 错误处理规则
  - 常用中间件模板

### 7. [Route层AI代码生成规范](./core/api/route/route_spec.md)
- **功能**: 定义路由层的生成规范
- **核心内容**:
  - 路由函数规范
  - 依赖注入规则
  - 路由注册规则
  - 路由组实现模板

## AI生成代码时的注意事项

### 通用规则
1. **遵循Clean Architecture**: 保持各层职责分离
2. **依赖注入**: 通过构造函数注入依赖，使用接口类型
3. **错误处理**: 统一错误处理和返回格式
4. **上下文管理**: 使用context控制请求生命周期
5. **命名规范**: 遵循项目既定的命名约定

### 代码质量要求
1. **类型安全**: 使用Go泛型确保类型安全
2. **接口抽象**: 通过接口实现松耦合
3. **测试友好**: 代码结构便于单元测试
4. **性能考虑**: 合理使用超时控制和资源管理

### 文件组织
1. **目录结构**: 严格遵循项目目录结构规范
2. **导入顺序**: 遵循Go导入顺序规范
3. **包命名**: 与目录结构保持一致
4. **注释规范**: 为导出的类型和方法添加注释

## 使用指导

AI在生成代码时应：
1. 首先参考对应层的规范文档
2. 使用提供的代码生成模板
3. 遵循该层的命名和实现规范
4. 确保与其他层的接口兼容