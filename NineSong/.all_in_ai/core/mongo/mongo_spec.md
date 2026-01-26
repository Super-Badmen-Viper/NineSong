# MongoDB层AI代码生成规范

## 代码结构规范

### 接口定义规范
- 文件路径: `mongo/mongo.go`
- 核心接口:
  - `Database` - 数据库操作接口
  - `Collection` - 集合操作接口
  - `Client` - 客户端操作接口
  - `SingleResult, Cursor` - 查询结果接口
  - `BulkWrite, BulkWriteResult` - 批量操作接口

### 实现类规范
- 实现前缀: `mongo` (如 `mongoClient`, `mongoDatabase`)
- 所有实现必须包装对应的官方驱动类型
- 方法命名与官方驱动保持一致
- 实现类必须遵循接口定义

## 代码生成模板

### 新建MongoDB接口实现
```
type <EntityName>DB struct {
    db         mongo.Database
    collection string
}

func New<EntityName>DB(db mongo.Database, collection string) *<EntityName>DB {
    return &<EntityName>DB{
        db:         db,
        collection: collection,
    }
}
```

### 查询方法生成规范
- 使用 `context.Context` 作为第一个参数
- 返回值格式: `(result, error)`
- 查询参数使用 `interface{}` 类型
- 分页查询需包含 `options.FindOptions`
- 使用bson标签进行数据映射

### 错误处理规范
- 所有错误必须包装使用 `fmt.Errorf`
- 使用 `errors.Is` 进行错误类型判断
- 错误消息格式: `"operation failed: %w"`
- 特定错误类型检查使用官方驱动错误类型

## 依赖注入规范
- 所有MongoDB依赖通过构造函数注入
- 不允许在实现中直接创建客户端连接
- 使用 `mongo.Database` 类型作为依赖注入参数
- 依赖注入使用接口类型而非具体实现

## 命名规范
- 变量命名: `db`, `collection`, `client`, `result`, `ctx`, `filter`, `update`
- 方法命名: 遵循官方驱动命名规范
- 结构体命名: `<EntityName>DB` 或 `<FeatureName>DB`
- 集合命名: `domain.Collection<EntityName>`

## 常用代码片段

### 连接初始化
```go
func NewMongoDatabase(env *Env) mongo.Client {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    dbHost := env.DBHost
    dbPort := env.DBPort
    dbUser := env.DBUser
    dbPass := env.DBPass
    
    mongodbURI := fmt.Sprintf("mongodb://%s:%s@%s:%s", dbUser, dbPass, dbHost, "27017")
    
    if dbUser == "" || dbPass == "" {
        mongodbURI = fmt.Sprintf("mongodb://%s:%s", dbHost, dbPort)
    }
    
    client, err := mongo.NewClient(mongodbURI)
    if err != nil {
        log.Fatal(err)
    }
    
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    err = client.Ping(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    return client
}
```

### 通用查询操作
```go
func (r *<EntityName>DB) FindByCondition(ctx context.Context, filter interface{}) ([]*<Entity>, error) {
    collection := r.db.Collection(r.collection)
    cursor, err := collection.Find(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("failed to find entities: %w", err)
    }
    defer cursor.Close(ctx)
    
    var entities []*<Entity>
    if err = cursor.All(ctx, &entities); err != nil {
        return nil, fmt.Errorf("failed to decode entities: %w", err)
    }
    
    return entities, nil
}
```

### 通用插入操作
```go
func (r *<EntityName>DB) Create(ctx context.Context, entity *<Entity>) error {
    if entity == nil {
        return errors.New("entity cannot be nil")
    }
    
    collection := r.db.Collection(r.collection)
    result, err := collection.InsertOne(ctx, entity)
    if err != nil {
        return fmt.Errorf("failed to create entity: %w", err)
    }
    
    // 设置生成的ID
    if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
        entity.ID = oid
    }
    
    return nil
}
```

### 通用更新操作
```go
func (r *<EntityName>DB) UpdateByID(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error {
    collection := r.db.Collection(r.collection)
    
    updateDoc := bson.M{"$set": updates}
    updateDoc["updated_at"] = primitive.NewDateTimeFromTime(time.Now())
    
    result, err := collection.UpdateByID(ctx, id, updateDoc)
    if err != nil {
        return fmt.Errorf("failed to update entity: %w", err)
    }
    
    if result.MatchedCount == 0 {
        return errors.New("entity not found")
    }
    
    return nil
}
```

### 通用删除操作
```go
func (r *<EntityName>DB) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
    collection := r.db.Collection(r.collection)
    
    result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return fmt.Errorf("failed to delete entity: %w", err)
    }
    
    if result.DeletedCount == 0 {
        return errors.New("entity not found")
    }
    
    return nil
}
```

### 通用查找单个操作
```go
func (r *<EntityName>DB) GetByID(ctx context.Context, id primitive.ObjectID) (*<Entity>, error) {
    collection := r.db.Collection(r.collection)
    
    var entity <Entity>
    err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&entity)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("entity not found")
        }
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }
    
    return &entity, nil
}
```