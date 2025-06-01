package repository_file_entity

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		"_id":         folder.ID,
		"folder_path": folder.FolderPath,
		"file_types":  folder.FileTypes,
		"folder_meta": bson.M{
			"file_count":   folder.FolderMeta.FileCount,
			"last_scanned": folder.FolderMeta.LastScanned,
		},
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

func (r *folderRepo) UpdateStats(ctx context.Context, folderID primitive.ObjectID, fileCount int) error {
	collection := r.db.Collection(r.collection)
	_, err := collection.UpdateByID(
		ctx,
		folderID,
		bson.M{
			"$set": bson.M{
				"folder_meta.file_count":   fileCount,
				"folder_meta.last_scanned": time.Now(),
			},
		},
	)
	return err
}
