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

func (r *folderRepo) Insert(ctx context.Context, folder *domain_file_entity.LibraryFolderMetadata) error {
	collection := r.db.Collection(r.collection)

	folder.FolderPath =
		filepath.ToSlash(
			filepath.Clean(folder.FolderPath))
	document := bson.M{
		"_id":          folder.ID,
		"folder_path":  folder.FolderPath,
		"file_types":   folder.FileTypes,
		"file_count":   folder.FileCount,
		"last_scanned": folder.LastScanned,
	}

	_, err := collection.InsertOne(ctx, document)
	return err
}

func (r *folderRepo) FindLibrary(
	ctx context.Context,
	folderPath string,
	fileTypes []domain_file_entity.FileTypeNo,
) (*domain_file_entity.LibraryFolderMetadata, error) {
	collection := r.db.Collection(r.collection)

	// 标准化路径处理
	normalizedPath := strings.TrimSuffix(
		filepath.ToSlash(filepath.Clean(folderPath)), "/")
	fmt.Printf("DEBUG - 查询路径: 原始='%s' 标准化='%s'\n", folderPath, normalizedPath)
	fmt.Printf("DEBUG - 查询文件类型: %v\n", fileTypes)

	// 构建精确匹配查询条件[1,2](@ref)
	query := bson.M{
		"folder_path": normalizedPath,
		"file_types":  bson.M{"$all": fileTypes}, // 使用$all确保包含所有指定文件类型
	}

	var folder domain_file_entity.LibraryFolderMetadata
	err := collection.FindOne(ctx, query).Decode(&folder)

	if err != nil {
		if domain.IsNotFound(err) {
			fmt.Printf("INFO - 未找到匹配记录: 路径=%s, 文件类型=%v\n", normalizedPath, fileTypes)
			return nil, nil
		}
		fmt.Printf("ERROR - 数据库查询失败: %v\n", err)
		return nil, fmt.Errorf("数据库操作失败: %w", err)
	}

	fmt.Printf("DEBUG - 找到匹配记录: ID=%s\n", folder.ID.Hex())
	return &folder, nil
}

func (r *folderRepo) GetAllByType(ctx context.Context, fileType domain_file_entity.FileTypeNo) ([]*domain_file_entity.LibraryFolderMetadata, error) {
	collection := r.db.Collection(r.collection)
	filter := bson.M{"file_types": fileType}

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

func (r *folderRepo) DetectingDuplicates(
	ctx context.Context,
	folderPath string,
	fileTypes []domain_file_entity.FileTypeNo,
) (*domain_file_entity.LibraryFolderMetadata, error) {
	// 标准化路径
	normalizedPath := filepath.ToSlash(filepath.Clean(folderPath))

	// 构建查询条件：路径完全匹配 + 文件类型完全匹配
	query := bson.M{
		"folder_path": normalizedPath,
		"file_types":  bson.M{"$all": fileTypes}, // 确保所有文件类型均匹配[1](@ref)
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
	if existing, _ := r.DetectingDuplicates(ctx, folder.FolderPath, folder.FileTypes); existing != nil {
		return domain_file_entity.ErrLibraryDuplicate
	}

	folder.ID = primitive.NewObjectID()
	folder.CreatedAt = time.Now()
	folder.UpdatedAt = time.Now()
	folder.Status = domain_file_entity.StatusActive

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
