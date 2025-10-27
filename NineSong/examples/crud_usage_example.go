/*
==================================================================================
                        NineSong CRUD模板使用示例
                        Clean Architecture 分层架构示例
==================================================================================

本示例按照Clean Architecture模式，展示四种CRUD模板的完整用法：
1. BaseRepository     - 标准CRUD操作
2. SearchableRepository - 带搜索功能的CRUD
3. ConfigRepository   - 配置管理专用
4. AuditableRepository - 自动审计功能

每个示例包含：Domain层定义 → Repository层实现 → Usecase层业务逻辑 → 完整集成示例
*/

package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
==================================================================================
                            1. 标准CRUD模板 (BaseRepository)
==================================================================================
适用场景：基础实体管理，如用户、订单、商品等标准业务实体
核心功能：Create、Read、Upsert、Delete + 批量操作 + 分页查询
*/

// ========== Domain 层 ==========

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Email    string             `bson:"email" json:"email"`
	Phone    string             `bson:"phone" json:"phone"`
	Status   string             `bson:"status" json:"status"` // active, inactive
	CreateAt primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdateAt primitive.DateTime `bson:"updated_at" json:"updated_at"`
}

// 继承基础CRUD接口
type UserRepository interface {
	domain.BaseRepository[User]
	// 自定义业务方法
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetActiveUsers(ctx context.Context) ([]*User, error)
}

type UserUsecase interface {
	usecase.BaseUsecase[User]
	// 自定义业务方法
	RegisterUser(ctx context.Context, name, email, phone string) (*User, error)
	ActivateUser(ctx context.Context, userID string) error
}

// ========== Repository 层 ==========

type userRepositoryImpl struct {
	*repository.BaseMongoRepository[User]
}

func NewUserRepository(db mongo.Database) UserRepository {
	base := repository.NewBaseMongoRepository[User](db, "users")
	return &userRepositoryImpl{BaseMongoRepository: base.(*repository.BaseMongoRepository[User])}
}

func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*User, error) {
	filter := map[string]interface{}{"email": email}
	return r.GetOneByFilter(ctx, filter)
}

func (r *userRepositoryImpl) GetActiveUsers(ctx context.Context) ([]*User, error) {
	filter := map[string]interface{}{"status": "active"}
	return r.GetByFilter(ctx, filter)
}

// ========== Usecase 层 ==========

type userUsecaseImpl struct {
	usecase.BaseUsecase[User]
	repo UserRepository
}

func NewUserUsecase(repo UserRepository, timeout time.Duration) UserUsecase {
	base := usecase.NewBaseUsecase[User](repo, timeout)
	return &userUsecaseImpl{BaseUsecase: base, repo: repo}
}

func (uc *userUsecaseImpl) RegisterUser(ctx context.Context, name, email, phone string) (*User, error) {
	// 业务逻辑：检查邮箱是否已存在
	existing, _ := uc.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	user := &User{Name: name, Email: email, Phone: phone, Status: "active"}
	createdUser, err := uc.Create(ctx, user)
	return createdUser, err
}

func (uc *userUsecaseImpl) ActivateUser(ctx context.Context, userID string) error {
	updates := map[string]interface{}{"status": "active"}
	return uc.UpdateByID(ctx, userID, updates)
}

/*
==================================================================================
                        2. 搜索CRUD模板 (SearchableRepository)
==================================================================================
适用场景：需要复杂查询的实体，如商品搜索、文档管理、内容检索
核心功能：基础CRUD + 名称搜索 + 模式匹配 + 排序查询
*/

// ========== Domain 层 ==========

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Category    string             `bson:"category" json:"category"`
	Price       float64            `bson:"price" json:"price"`
	Description string             `bson:"description" json:"description"`
	Tags        []string           `bson:"tags" json:"tags"`
}

type ProductRepository interface {
	domain.SearchableRepository[Product]
	// 业务特定方法
	GetByPriceRange(ctx context.Context, min, max float64) ([]*Product, error)
	GetByCategory(ctx context.Context, category string) ([]*Product, error)
}

type ProductUsecase interface {
	usecase.SearchableUsecase[Product]
	// 业务方法
	SearchProducts(ctx context.Context, keyword string, category string, minPrice, maxPrice float64) ([]*Product, error)
}

// ========== Repository 层 ==========

type productRepositoryImpl struct {
	*repository.SearchableMongoRepository[Product]
}

func NewProductRepository(db mongo.Database) ProductRepository {
	base := repository.NewSearchableMongoRepository[Product](db, "products")
	return &productRepositoryImpl{SearchableMongoRepository: base.(*repository.SearchableMongoRepository[Product])}
}

func (r *productRepositoryImpl) GetByPriceRange(ctx context.Context, min, max float64) ([]*Product, error) {
	filter := map[string]interface{}{
		"price": map[string]interface{}{"$gte": min, "$lte": max},
	}
	return r.GetByFilter(ctx, filter)
}

func (r *productRepositoryImpl) GetByCategory(ctx context.Context, category string) ([]*Product, error) {
	filter := map[string]interface{}{"category": category}
	return r.GetByFilter(ctx, filter)
}

// ========== Usecase 层 ==========

type productUsecaseImpl struct {
	usecase.SearchableUsecase[Product]
	repo ProductRepository
}

func NewProductUsecase(repo ProductRepository, timeout time.Duration) ProductUsecase {
	base := usecase.NewSearchableUsecase[Product](repo, timeout)
	return &productUsecaseImpl{SearchableUsecase: base, repo: repo}
}

func (uc *productUsecaseImpl) SearchProducts(ctx context.Context, keyword, category string, minPrice, maxPrice float64) ([]*Product, error) {
	var products []*Product
	var err error

	// 根据不同条件组合查询
	if keyword != "" {
		products, err = uc.Search(ctx, keyword)
	} else if category != "" {
		products, err = uc.repo.GetByCategory(ctx, category)
	} else if minPrice > 0 && maxPrice > 0 {
		products, err = uc.repo.GetByPriceRange(ctx, minPrice, maxPrice)
	} else {
		products, err = uc.GetAll(ctx)
	}

	return products, err
}

/*
==================================================================================
                         3. 配置CRUD模板 (ConfigRepository)
==================================================================================
适用场景：系统配置、应用设置、全局参数等单例或少量配置管理
核心功能：Get、Upsert、批量替换 + 单例模式支持
*/

// ========== Domain 层 ==========

type SystemConfig struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AppName      string             `bson:"app_name" json:"app_name"`
	Version      string             `bson:"version" json:"version"`
	MaxFileSize  int64              `bson:"max_file_size" json:"max_file_size"`
	EnableDebug  bool               `bson:"enable_debug" json:"enable_debug"`
	DatabaseURL  string             `bson:"database_url" json:"database_url"`
	CacheTimeout int                `bson:"cache_timeout" json:"cache_timeout"`
}

type SystemConfigRepository interface {
	domain.ConfigRepository[SystemConfig]
}

type SystemConfigUsecase interface {
	usecase.ConfigUsecase[SystemConfig]
	// 业务方法
	InitDefaultConfig(ctx context.Context) error
	UpdateCacheTimeout(ctx context.Context, timeout int) error
	ToggleDebugMode(ctx context.Context) error
}

// ========== Repository 层 ==========

type systemConfigRepositoryImpl struct {
	*repository.ConfigMongoRepository[SystemConfig]
}

func NewSystemConfigRepository(db mongo.Database) SystemConfigRepository {
	base := repository.NewConfigMongoRepository[SystemConfig](db, "system_config")
	return &systemConfigRepositoryImpl{ConfigMongoRepository: base.(*repository.ConfigMongoRepository[SystemConfig])}
}

// ========== Usecase 层 ==========

type systemConfigUsecaseImpl struct {
	usecase.ConfigUsecase[SystemConfig]
	repo SystemConfigRepository
}

func NewSystemConfigUsecase(repo SystemConfigRepository, timeout time.Duration) SystemConfigUsecase {
	base := usecase.NewConfigUsecase[SystemConfig](repo, timeout)
	return &systemConfigUsecaseImpl{ConfigUsecase: base, repo: repo}
}

func (uc *systemConfigUsecaseImpl) InitDefaultConfig(ctx context.Context) error {
	defaultConfig := &SystemConfig{
		AppName:      "NineSong",
		Version:      "1.0.0",
		MaxFileSize:  10 * 1024 * 1024, // 10MB
		EnableDebug:  false,
		CacheTimeout: 3600, // 1小时
	}
	return uc.Upsert(ctx, defaultConfig)
}

func (uc *systemConfigUsecaseImpl) UpdateCacheTimeout(ctx context.Context, timeout int) error {
	config, err := uc.Get(ctx)
	if err != nil {
		return err
	}
	config.CacheTimeout = timeout
	return uc.Upsert(ctx, config)
}

func (uc *systemConfigUsecaseImpl) ToggleDebugMode(ctx context.Context) error {
	config, err := uc.Get(ctx)
	if err != nil {
		return err
	}
	config.EnableDebug = !config.EnableDebug
	return uc.Upsert(ctx, config)
}

/*
==================================================================================
                        4. 审计CRUD模板 (AuditableRepository)
==================================================================================
适用场景：需要追踪创建/修改时间的实体，如日志、操作记录、审计数据
核心功能：基础CRUD + 自动时间戳 + 时间范围查询
*/

// ========== Domain 层 ==========

type OperationLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Operation string             `bson:"operation" json:"operation"`
	Resource  string             `bson:"resource" json:"resource"`
	Details   string             `bson:"details" json:"details"`
	IPAddress string             `bson:"ip_address" json:"ip_address"`
	Success   bool               `bson:"success" json:"success"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at" json:"updated_at"`
}

type OperationLogRepository interface {
	domain.AuditableRepository[OperationLog]
	// 业务方法
	GetByUser(ctx context.Context, userID primitive.ObjectID) ([]*OperationLog, error)
	GetFailedOperations(ctx context.Context) ([]*OperationLog, error)
}

type OperationLogUsecase interface {
	usecase.BaseUsecase[OperationLog]
	// 业务方法
	LogOperation(ctx context.Context, userID primitive.ObjectID, operation, resource, details, ip string, success bool) error
	GetUserOperationHistory(ctx context.Context, userID string, days int) ([]*OperationLog, error)
	GetRecentFailures(ctx context.Context, hours int) ([]*OperationLog, error)
}

// ========== Repository 层 ==========

type operationLogRepositoryImpl struct {
	*repository.AuditableMongoRepository[OperationLog]
}

func NewOperationLogRepository(db mongo.Database) OperationLogRepository {
	base := repository.NewAuditableMongoRepository[OperationLog](db, "operation_logs")
	return &operationLogRepositoryImpl{AuditableMongoRepository: base.(*repository.AuditableMongoRepository[OperationLog])}
}

func (r *operationLogRepositoryImpl) GetByUser(ctx context.Context, userID primitive.ObjectID) ([]*OperationLog, error) {
	filter := map[string]interface{}{"user_id": userID}
	return r.GetByFilter(ctx, filter)
}

func (r *operationLogRepositoryImpl) GetFailedOperations(ctx context.Context) ([]*OperationLog, error) {
	filter := map[string]interface{}{"success": false}
	return r.GetByFilter(ctx, filter)
}

// ========== Usecase 层 ==========

type operationLogUsecaseImpl struct {
	usecase.BaseUsecase[OperationLog]
	repo OperationLogRepository
}

func NewOperationLogUsecase(repo OperationLogRepository, timeout time.Duration) OperationLogUsecase {
	base := usecase.NewBaseUsecase[OperationLog](repo, timeout)
	return &operationLogUsecaseImpl{BaseUsecase: base, repo: repo}
}

func (uc *operationLogUsecaseImpl) LogOperation(ctx context.Context, userID primitive.ObjectID, operation, resource, details, ip string, success bool) error {
	log := &OperationLog{
		UserID:    userID,
		Operation: operation,
		Resource:  resource,
		Details:   details,
		IPAddress: ip,
		Success:   success,
	}
	_, err := uc.Create(ctx, log)
	return err
}

func (uc *operationLogUsecaseImpl) GetUserOperationHistory(ctx context.Context, userID string, days int) ([]*OperationLog, error) {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 获取指定天数内的操作记录
	since := primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, -days))
	return uc.repo.GetCreatedAfter(ctx, since)
}

func (uc *operationLogUsecaseImpl) GetRecentFailures(ctx context.Context, hours int) ([]*OperationLog, error) {
	since := primitive.NewDateTimeFromTime(time.Now().Add(-time.Duration(hours) * time.Hour))
	allRecent, err := uc.repo.GetCreatedAfter(ctx, since)
	if err != nil {
		return nil, err
	}

	// 过滤失败的操作
	var failures []*OperationLog
	for _, log := range allRecent {
		if !log.Success {
			failures = append(failures, log)
		}
	}
	return failures, nil
}

/*
==================================================================================
                            完整集成使用示例
==================================================================================
展示如何在实际项目中整合所有CRUD模板
*/

// CleanArchitectureExample 展示Clean Architecture模式下的完整集成
func CleanArchitectureExample(db mongo.Database) {
	ctx := context.Background()
	timeout := 30 * time.Second

	// ========== 1. 标准CRUD示例 ==========
	fmt.Println("=== 标准CRUD示例 ===")

	userRepo := NewUserRepository(db)
	userUC := NewUserUsecase(userRepo, timeout)

	// 注册用户
	user, err := userUC.RegisterUser(ctx, "张三", "zhangsan@example.com", "13800138000")
	if err != nil {
		fmt.Printf("用户注册失败: %v\n", err)
		return
	}
	fmt.Printf("用户注册成功: %s (ID: %s)\n", user.Name, user.ID.Hex())

	// ========== 2. 搜索CRUD示例 ==========
	fmt.Println("\n=== 搜索CRUD示例 ===")

	productRepo := NewProductRepository(db)
	productUC := NewProductUsecase(productRepo, timeout)

	// 创建商品
	product := &Product{Name: "iPhone 15", Category: "手机", Price: 7999.0, Tags: []string{"苹果", "手机", "电子产品"}}
	_, err = productUC.Create(ctx, product)
	if err != nil {
		fmt.Printf("商品创建失败: %v\n", err)
		return
	}

	// 搜索商品
	products, err := productUC.SearchProducts(ctx, "iPhone", "手机", 5000, 10000)
	if err == nil && len(products) > 0 {
		fmt.Printf("搜索到商品: %s, 价格: %.2f\n", products[0].Name, products[0].Price)
	}

	// ========== 3. 配置CRUD示例 ==========
	fmt.Println("\n=== 配置CRUD示例 ===")

	configRepo := NewSystemConfigRepository(db)
	configUC := NewSystemConfigUsecase(configRepo, timeout)

	// 初始化默认配置
	err = configUC.InitDefaultConfig(ctx)
	if err != nil {
		fmt.Printf("配置初始化失败: %v\n", err)
		return
	}

	// 获取配置
	config, err := configUC.Get(ctx)
	if err == nil {
		fmt.Printf("当前配置: %s v%s, 调试模式: %t\n", config.AppName, config.Version, config.EnableDebug)
	}

	// 切换调试模式
	configUC.ToggleDebugMode(ctx)
	fmt.Println("调试模式已切换")

	// ========== 4. 审计CRUD示例 ==========
	fmt.Println("\n=== 审计CRUD示例 ===")

	logRepo := NewOperationLogRepository(db)
	logUC := NewOperationLogUsecase(logRepo, timeout)

	// 记录操作日志
	err = logUC.LogOperation(ctx, user.ID, "CREATE", "Product", "创建商品: iPhone 15", "192.168.1.100", true)
	if err != nil {
		fmt.Printf("日志记录失败: %v\n", err)
		return
	}

	// 查询用户操作历史
	history, err := logUC.GetUserOperationHistory(ctx, user.ID.Hex(), 7)
	if err == nil {
		fmt.Printf("用户 %s 近7天操作记录: %d 条\n", user.Name, len(history))
	}

	fmt.Println("\n=== 所有CRUD模板示例完成 ===")
}

/*
==================================================================================
                               最佳实践总结
==================================================================================

1. 分层职责明确：
   - Domain层：定义实体和接口契约
   - Repository层：实现数据访问逻辑
   - Usecase层：实现业务逻辑和规则

2. 模板选择指南：
   - BaseRepository: 标准业务实体
   - SearchableRepository: 需要复杂查询的实体
   - ConfigRepository: 配置类数据
   - AuditableRepository: 需要时间追踪的数据

3. 扩展模式：
   - 继承模板接口获得标准功能
   - 在接口中添加业务特定方法
   - Repository层实现数据访问细节
   - Usecase层实现业务逻辑验证

4. 错误处理：
   - Repository层返回数据访问错误
   - Usecase层处理业务逻辑错误
   - 统一错误格式和错误码

5. 性能优化：
   - 合理使用分页查询
   - 创建必要的数据库索引
   - 利用批量操作提高效率
   - 设置合适的超时时间
*/
