# Repository Auth AI代码生成规范

## 代码结构规范

### Repository实现规范
- 文件路径: `repository/repository_auth/<auth_entity>_repository.go`
- 继承基础Repository: `repository.BaseMongoRepository[<AuthEntity>]`
- 使用`base_mongo_repository.go`作为基础实现
- 认证Repository支持邮箱唯一性验证和密码安全存储

### 接口实现规范
- 实现`domain.BaseRepository[<AuthEntity>]`接口
- 支持认证实体的获取、创建、更新、删除操作
- 包含认证专用方法如按邮箱获取、邮箱唯一性检查等

## 代码生成模板

### 新建用户Repository实现
```
package repository_auth

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
    "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

type UserRepositoryImpl struct {
    repository.BaseMongoRepository[domain.User]
}

func NewUserRepositoryImpl(db *mongo.Database) domain.UserRepository {
    return &UserRepositoryImpl{
        BaseMongoRepository: repository.NewBaseMongoRepository[domain.User](db, "users"),
    }
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (domain.User, error) {
    var user domain.User
    filter := bson.M{"email": email}
    err := r.BaseMongoRepository.Collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return user, domain.ErrUserNotFound
        }
        return user, err
    }
    return user, nil
}

// IsEmailTaken 检查邮箱是否已被使用
func (r *UserRepositoryImpl) IsEmailTaken(ctx context.Context, email string, skipID primitive.ObjectID) (bool, error) {
    filter := bson.M{"email": email}
    if !skipID.IsZero() {
        filter["_id"] = bson.M{"$ne": skipID}
    }
    
    var user domain.User
    err := r.BaseMongoRepository.Collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return false, nil // 邮箱未被使用
        }
        return false, err
    }
    
    return true, nil // 邮箱已被使用
}
```

### 认证Repository接口实现
```
// 实现 domain.BaseRepository[User] 接口的所有方法
// 通过嵌入 repository.BaseMongoRepository[User] 自动实现基础CRUD方法

// 自定义认证相关方法
func (r *UserRepositoryImpl) GetActiveUsers(ctx context.Context) ([]domain.User, error) {
    filter := bson.M{"admin": false} // 假设非管理员为普通用户
    cursor, err := r.BaseMongoRepository.Collection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var users []domain.User
    if err = cursor.All(ctx, &users); err != nil {
        return nil, err
    }
    
    return users, nil
}

func (r *UserRepositoryImpl) GetAdminUsers(ctx context.Context) ([]domain.User, error) {
    filter := bson.M{"admin": true}
    cursor, err := r.BaseMongoRepository.Collection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var users []domain.User
    if err = cursor.All(ctx, &users); err != nil {
        return nil, err
    }
    
    return users, nil
}
```

## 代码生成规则

### Repository结构规则
- 使用结构体嵌入实现基础功能
- 通过`repository.BaseMongoRepository[<AuthEntity>]`实现通用操作
- 自定义方法实现认证专用逻辑

### 查询操作规则
- 支持按邮箱、用户名等字段查询
- 使用bson.M构建查询条件
- 实现用户权限过滤

### 更新操作规则
- 用户更新时设置时间戳
- 验证更新前后的数据一致性
- 支持密码更新的安全处理

### 命名规则
- Repository实现名: `<AuthEntity>RepositoryImpl`
- 集合名称: 使用小写复数形式（如"users", "tokens"）
- 方法名: 使用驼峰命名法

## 依赖管理规范
- Repository层依赖Domain层接口定义
- 依赖MongoDB驱动和基础Repository实现
- 通过构造函数注入数据库连接

## 常用代码片段

### 用户邮箱唯一性验证方法
```go
func (r *UserRepositoryImpl) ValidateEmailUniqueness(ctx context.Context, email string, userID primitive.ObjectID) error {
    isTaken, err := r.IsEmailTaken(ctx, email, userID)
    if err != nil {
        return err
    }
    
    if isTaken {
        return domain.ErrEmailAlreadyExists
    }
    
    return nil
}
```

### 用户安全查询方法
```go
func (r *UserRepositoryImpl) GetUserSecure(ctx context.Context, id primitive.ObjectID) (domain.User, error) {
    filter := bson.M{"_id": id}
    var user domain.User
    err := r.BaseMongoRepository.Collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return user, domain.ErrUserNotFound
        }
        return user, err
    }
    
    // 移除敏感信息
    user.Password = ""
    return user, nil
}
```

### 用户权限检查方法
```go
func (r *UserRepositoryImpl) IsUserAdmin(ctx context.Context, userID primitive.ObjectID) (bool, error) {
    filter := bson.M{
        "_id":  userID,
        "admin": true,
    }
    
    var user domain.User
    err := r.BaseMongoRepository.Collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return false, nil // 不是管理员
        }
        return false, err
    }
    
    return true, nil // 是管理员
}
```

### 用户统计查询方法
```go
func (r *UserRepositoryImpl) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
    pipeline := []bson.M{
        {
            "$group": bson.M{
                "_id": nil,
                "total": bson.M{"$sum": 1},
                "admins": bson.M{
                    "$sum": bson.M{
                        "$cond": []interface{}{
                            bson.M{"$eq": []interface{}{"$admin", true}},
                            1,
                            0,
                        },
                    },
                },
                "regularUsers": bson.M{
                    "$sum": bson.M{
                        "$cond": []interface{}{
                            bson.M{"$eq": []interface{}{"$admin", false}},
                            1,
                            0,
                        },
                    },
                },
            },
        },
    }
    
    cursor, err := r.BaseMongoRepository.Collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var results []map[string]interface{}
    if err = cursor.All(ctx, &results); err != nil {
        return nil, err
    }
    
    if len(results) == 0 {
        return map[string]interface{}{
            "total":       0,
            "admins":      0,
            "regularUsers": 0,
        }, nil
    }
    
    return results[0], nil
}
```

### 用户登录历史记录方法
```go
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID primitive.ObjectID) error {
    filter := bson.M{"_id": userID}
    
    update := bson.M{
        "$set": bson.M{
            "last_login": primitive.NewDateTimeFromTime(time.Now()),
            "updated_at": primitive.NewDateTimeFromTime(time.Now()),
        },
    }
    
    result, err := r.BaseMongoRepository.Collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }
    
    if result.MatchedCount == 0 {
        return domain.ErrUserNotFound
    }
    
    return nil
}
```

### 用户安全更新方法
```go
func (r *UserRepositoryImpl) UpdateUserPassword(ctx context.Context, userID primitive.ObjectID, newPassword string) error {
    filter := bson.M{"_id": userID}
    
    update := bson.M{
        "$set": bson.M{
            "password":   newPassword,
            "updated_at": primitive.NewDateTimeFromTime(time.Now()),
        },
    }
    
    result, err := r.BaseMongoRepository.Collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }
    
    if result.MatchedCount == 0 {
        return domain.ErrUserNotFound
    }
    
    return nil
}
```