package repository_file_entity

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type folderRepo struct {
	db         mongo.Database
	collection string
}

func NewFolderRepo(db mongo.Database, collection string) domain_file_entity.FolderRepository {
	return &folderRepo{
		db:         db,
		collection: collection,
	}
}

func (r *folderRepo) ListFolders(path string) ([]domain_file_entity.FolderEntry, error) {
	// 处理根目录情况
	if path == "" {
		return r.getRootDirectories()
	}

	// 验证路径是否存在且是目录
	if err := validatePath(path); err != nil {
		return nil, err
	}

	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var folders []domain_file_entity.FolderEntry
	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(path, entry.Name())
			folders = append(folders, domain_file_entity.FolderEntry{
				Name: entry.Name(),
				Path: fullPath,
			})
		}
	}

	// 添加父目录选项（非根目录时）
	if path != "/" && runtime.GOOS != "windows" {
		parentPath := filepath.Dir(path)
		folders = append([]domain_file_entity.FolderEntry{
			{Name: ".. (上级目录)", Path: parentPath},
		}, folders...)
	}

	return folders, nil
}

// 获取根目录（跨平台实现）
func (r *folderRepo) getRootDirectories() ([]domain_file_entity.FolderEntry, error) {
	if runtime.GOOS == "windows" {
		return r.getWindowsDrives()
	}
	return []domain_file_entity.FolderEntry{{Name: "/", Path: "/", IsRoot: true}}, nil
}

// Windows平台：获取所有驱动器
func (r *folderRepo) getWindowsDrives() ([]domain_file_entity.FolderEntry, error) {
	var drives []domain_file_entity.FolderEntry
	for drive := 'A'; drive <= 'Z'; drive++ {
		drivePath := string(drive) + ":\\"
		if _, err := os.Stat(drivePath); err == nil {
			drives = append(drives, domain_file_entity.FolderEntry{
				Name:   drivePath,
				Path:   drivePath,
				IsRoot: true,
			})
		}
	}
	return drives, nil
}

// 路径验证
func validatePath(path string) error {
	// 防止目录遍历攻击
	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	// 检查路径是否存在
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	// 检查是否为目录
	if !fileInfo.IsDir() {
		return errors.New("path is not a directory")
	}

	return nil
}

func (r *folderRepo) Insert(ctx context.Context, folder *domain_file_entity.LibraryFolderMetadata) error {
	collection := r.db.Collection(r.collection)

	folder.FolderPath =
		filepath.ToSlash(
			filepath.Clean(folder.FolderPath))
	document := bson.M{
		"_id":          folder.ID,
		"folder_path":  folder.FolderPath,
		"folder_type":  folder.FolderType,
		"file_count":   folder.FileCount,
		"last_scanned": folder.LastScanned,
	}

	_, err := collection.InsertOne(ctx, document)
	return err
}

func (r *folderRepo) FindLibrary(
	ctx context.Context,
	folderPath string,
	folderType int,
) (*domain_file_entity.LibraryFolderMetadata, error) {
	collection := r.db.Collection(r.collection)

	// 标准化路径处理
	normalizedPath := strings.TrimSuffix(
		filepath.ToSlash(filepath.Clean(folderPath)), "/")
	fmt.Printf("DEBUG - 查询路径: 原始='%s' 标准化='%s'\n", folderPath, normalizedPath)
	fmt.Printf("DEBUG - 查询文件类型: %v\n", folderType)

	// 构建精确匹配查询条件[1,2](@ref)
	query := bson.M{
		"folder_path": normalizedPath,
		"folder_type": folderType, // 使用$all确保包含所有指定文件类型
	}

	var folder domain_file_entity.LibraryFolderMetadata
	err := collection.FindOne(ctx, query).Decode(&folder)

	if err != nil {
		if domain.IsNotFound(err) {
			fmt.Printf("INFO - 未找到匹配记录: 路径=%s, 文件类型=%v\n", normalizedPath, folderType)
			return nil, nil
		}
		fmt.Printf("ERROR - 数据库查询失败: %v\n", err)
		return nil, fmt.Errorf("数据库操作失败: %w", err)
	}

	fmt.Printf("DEBUG - 找到匹配记录: ID=%s\n", folder.ID.Hex())
	return &folder, nil
}

func (r *folderRepo) GetAllByType(ctx context.Context, folderType int) ([]*domain_file_entity.LibraryFolderMetadata, error) {
	collection := r.db.Collection(r.collection)
	filter := bson.M{"folder_type": folderType}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("获取所有库失败: %w", err)
	}
	defer cursor.Close(ctx)

	var allFolders []*domain_file_entity.LibraryFolderMetadata
	for cursor.Next(ctx) {
		var folder domain_file_entity.LibraryFolderMetadata
		if err := cursor.Decode(&folder); err != nil {
			return nil, fmt.Errorf("解码库数据失败: %w", err)
		}
		allFolders = append(allFolders, &folder)
	}

	// 检查路径存在性并过滤
	validFolders := make([]*domain_file_entity.LibraryFolderMetadata, 0, len(allFolders))
	for _, folder := range allFolders {
		exists, err := pathExists(folder.FolderPath)
		if err != nil {
			log.Printf("路径检查错误: %s - %v", folder.FolderPath, err)
			continue
		}
		if exists {
			validFolders = append(validFolders, folder)
		} else {
			log.Printf("跳过不存在路径: %s", folder.FolderPath)
		}
	}

	return validFolders, nil
}
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // 路径存在
	}
	if os.IsNotExist(err) {
		return false, nil // 路径不存在
	}
	return false, err // 其他错误（权限问题等）
}

func (r *folderRepo) UpdateStats(
	ctx context.Context,
	folderID primitive.ObjectID,
	fileCount int, status domain_file_entity.LibraryStatus,
) error {
	collection := r.db.Collection(r.collection)
	_, err := collection.UpdateByID(
		ctx,
		folderID,
		bson.M{
			"$set": bson.M{
				"folder_meta.file_count":   fileCount,
				"folder_meta.last_scanned": time.Now(),
				"status":                   status,
			},
		},
	)
	return err
}

func (r *folderRepo) DetectingDuplicates(ctx context.Context, folderPath string, folderType int) (*domain_file_entity.LibraryFolderMetadata, error) {
	normalizedPath := filepath.ToSlash(filepath.Clean(folderPath))

	query := bson.M{
		"folder_path": normalizedPath,
		"folder_type": folderType,
	}

	var existingLib domain_file_entity.LibraryFolderMetadata
	err := r.db.Collection(r.collection).FindOne(ctx, query).Decode(&existingLib)

	// 错误处理
	switch {
	case errors.Is(err, driver.ErrNoDocuments):
		return nil, nil // 无重复
	case err != nil:
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	default:
		return &existingLib, nil // 返回重复项
	}
}

func (r *folderRepo) Create(ctx context.Context, folder *domain_file_entity.LibraryFolderMetadata) error {
	// 检查重复媒体库
	if existing, _ := r.DetectingDuplicates(ctx, folder.FolderPath, folder.FolderType); existing != nil {
		return domain_file_entity.ErrLibraryDuplicate
	}

	folder.ID = primitive.NewObjectID()
	folder.CreatedAt = time.Now()
	folder.UpdatedAt = time.Now()
	folder.Status = domain_file_entity.StatusActive
	folder.FolderPath = filepath.ToSlash(filepath.Clean(folder.FolderPath))

	_, err := r.db.Collection(r.collection).InsertOne(ctx, folder)
	return err
}

func (r *folderRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	// 检查媒体库状态
	if lib, err := r.GetByID(ctx, id); err == nil && lib.Status == domain_file_entity.StatusScanning {
		return domain_file_entity.ErrLibraryInUse
	}

	filter := bson.M{"_id": id}
	result, err := r.db.Collection(r.collection).DeleteOne(ctx, filter)

	if result == 0 {
		return domain_file_entity.ErrLibraryNotFound
	}
	return err
}

func (r *folderRepo) Update(
	ctx context.Context,
	id primitive.ObjectID,
	newName string,
	newFolderPath string,
) error {
	// 1. 参数校验
	if strings.TrimSpace(newName) == "" {
		return errors.New("媒体库名称不能为空")
	}

	// 2. 构建原子更新操作
	update := bson.M{
		"$set": bson.M{
			"name":        newName,
			"folder_path": newFolderPath,
			"updatedAt":   time.Now(), // 自动更新修改时间
		},
	}

	// 3. 执行更新操作
	_, err := r.db.Collection(r.collection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
	)
	if err != nil {
		return fmt.Errorf("数据库更新失败: %w", err)
	}

	return nil
}

func (r *folderRepo) IsLibraryInUse(ctx context.Context, id primitive.ObjectID) (bool, error) {
	// 查询当前媒体库状态
	var lib domain_file_entity.LibraryFolderMetadata
	filter := bson.M{"_id": id}
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&lib)

	// 错误处理
	if err != nil {
		if errors.Is(err, driver.ErrNoDocuments) {
			return false, domain_file_entity.ErrLibraryNotFound
		}
		return false, fmt.Errorf("数据库查询失败: %w", err)
	}

	// 核心逻辑：状态检测
	return lib.Status == domain_file_entity.StatusScanning || lib.Status == domain_file_entity.StatusDisabled, nil
}

func (r *folderRepo) GetByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*domain_file_entity.LibraryFolderMetadata, error) {
	filter := bson.M{"_id": id}

	var lib domain_file_entity.LibraryFolderMetadata
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&lib)

	// 错误处理
	switch {
	case errors.Is(err, driver.ErrNoDocuments):
		return nil, domain_file_entity.ErrLibraryNotFound // 明确返回领域错误[1](@ref)
	case err != nil:
		return nil, fmt.Errorf("数据库操作失败: %w", err)
	default:
		return &lib, nil
	}
}

func (r *folderRepo) GetAll(ctx context.Context) ([]*domain_file_entity.LibraryFolderMetadata, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var libraries []*domain_file_entity.LibraryFolderMetadata
	if err := cursor.All(ctx, &libraries); err != nil {
		return nil, err
	}
	return libraries, nil
}

func (r *folderRepo) UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain_file_entity.LibraryStatus) error {
	_, err := r.db.Collection(r.collection).UpdateByID(
		ctx,
		id,
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}
