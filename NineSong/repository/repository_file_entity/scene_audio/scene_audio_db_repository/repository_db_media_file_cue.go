package scene_audio_db_repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mediaFileCueRepository struct {
	db         mongo.Database
	collection string
}

func NewMediaFileCueRepository(db mongo.Database, collection string) scene_audio_db_interface.MediaFileCueRepository {
	return &mediaFileCueRepository{
		db:         db,
		collection: collection,
	}
}

// Upsert 创建或更新CUE文件元数据 (添加并发锁)
func (r *mediaFileCueRepository) Upsert(ctx context.Context, file *scene_audio_db_models.MediaFileCueMetadata) (*scene_audio_db_models.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	now := time.Now().UTC()

	filter := bson.M{
		"path": file.Path,
	}

	update := file.ToUpdateDoc()
	update["$setOnInsert"] = bson.M{
		"created_at": now,
	}

	opts := options.Update().SetUpsert(true)
	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("CUE文件upsert失败: %w", err)
	}

	if result.UpsertedID != nil {
		file.ID = result.UpsertedID.(primitive.ObjectID)
		file.CreatedAt = now
	} else {
		var existing struct{ ID primitive.ObjectID }
		if err := coll.FindOne(ctx, bson.M{"path": file.Path}).Decode(&existing); err == nil {
			file.ID = existing.ID
		} else if errors.Is(err, driver.ErrNoDocuments) {
			return nil, fmt.Errorf("CUE文档未插入也未更新: %w", err)
		} else {
			return nil, fmt.Errorf("CUE文件ID查询失败: %w", err)
		}
	}

	file.UpdatedAt = now
	return file, nil
}

// BulkUpsert 批量创建/更新CUE文件
func (r *mediaFileCueRepository) BulkUpsert(ctx context.Context, files []*scene_audio_db_models.MediaFileCueMetadata) (int, error) {
	coll := r.db.Collection(r.collection)
	var successCount int

	for _, file := range files {
		filter := bson.M{"_id": file.ID}
		update := bson.M{"$set": file}

		_, err := coll.UpdateOne(
			ctx,
			filter,
			update,
			options.Update().SetUpsert(true),
		)

		if err != nil {
			return successCount, fmt.Errorf("bulk upsert失败于索引%d: %w", successCount, err)
		}
		successCount++
	}
	return successCount, nil
}

// DeleteByID 根据ID删除CUE文件 (保持原样)
func (r *mediaFileCueRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("删除CUE文件失败: %w", err)
	}
	return nil
}

// DeleteByPath 根据路径删除CUE文件 (保持原样)
func (r *mediaFileCueRepository) DeleteByPath(ctx context.Context, path string) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"path": path})
	if err != nil {
		return fmt.Errorf("根据路径删除CUE文件失败: %w", err)
	}
	return nil
}

func (r *mediaFileCueRepository) DeleteAllInvalid(
	ctx context.Context,
	filePaths []string,
) (int64, []struct {
	ArtistID primitive.ObjectID
	Count    int64
}, error) {
	coll := r.db.Collection(r.collection)
	deletedArtists := make([]struct {
		ArtistID primitive.ObjectID
		Count    int64
	}, 0)

	// 场景1：当filePaths为空时，不执行删除操作，避免误删所有文件
	if len(filePaths) == 0 {
		log.Printf("DeleteAllInvalid: filePaths为空，跳过删除操作")
		return 0, deletedArtists, nil
	}

	// 构建全局有效文件路径集合，使用正斜杠规范化
	validFilePaths := make(map[string]bool)
	for _, path := range filePaths {
		// 规范化文件路径，使用正斜杠格式，与数据库中存储的格式一致
		normalizedPath := filepath.ToSlash(filepath.Clean(path))
		validFilePaths[normalizedPath] = true
	}

	// 查询所有文档（移除folderPath过滤条件）
	cur, err := coll.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1, "path": 1, "artist_id": 1}))
	if err != nil {
		return 0, deletedArtists, fmt.Errorf("查询失败: %w", err)
	}
	defer cur.Close(ctx)

	// 按艺术家分组待删除项
	artistToIDs := make(map[primitive.ObjectID][]primitive.ObjectID)
	var toDelete []primitive.ObjectID

	for cur.Next(ctx) {
		var doc struct {
			ID       primitive.ObjectID `bson:"_id"`
			Path     string             `bson:"path"`
			ArtistID primitive.ObjectID `bson:"artist_id"`
		}
		if err := cur.Decode(&doc); err != nil {
			continue
		}

		// 检查文件路径是否在有效文件路径集合中，或者是否实际存在
		isValid := false

		// 检查文件是否在有效文件路径集合中
		if validFilePaths[doc.Path] {
			isValid = true
		} else {
			// 检查文件是否实际存在于文件系统中
			absPath := doc.Path
			// 将正斜杠转换为反斜杠，适配Windows系统
			absPath = strings.ReplaceAll(absPath, "/", "\\")
			if _, err := os.Stat(absPath); err == nil {
				// 文件实际存在，是有效文件
				isValid = true
			}
		}

		// 如果文件既不在有效文件路径集合中，也不存在于文件系统中，才是无效文件
		if !isValid {
			// 文件确实不存在，是无效文件，添加到删除列表
			toDelete = append(toDelete, doc.ID)
			artistToIDs[doc.ArtistID] = append(artistToIDs[doc.ArtistID], doc.ID)
		}
	}

	// 批量删除并统计（优化性能）
	totalDeleted := int64(0)
	const batchSize = 500
	artistCounts := make(map[primitive.ObjectID]int64)

	// 直接使用artistToIDs统计，避免嵌套循环
	for artistID, ids := range artistToIDs {
		artistCounts[artistID] = int64(len(ids))
	}

	// 批量删除
	for i := 0; i < len(toDelete); i += batchSize {
		end := i + batchSize
		if end > len(toDelete) {
			end = len(toDelete)
		}
		batch := toDelete[i:end]

		delResult, err := coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": batch}})
		if err != nil {
			return totalDeleted, deletedArtists, fmt.Errorf("批量删除失败: %w", err)
		}
		totalDeleted += delResult
	}

	// 构建艺术家删除统计
	for artistID, count := range artistCounts {
		deletedArtists = append(deletedArtists, struct {
			ArtistID primitive.ObjectID
			Count    int64
		}{ArtistID: artistID, Count: count})
	}

	return totalDeleted, deletedArtists, nil
}

func (r *mediaFileCueRepository) DeleteByFolder(ctx context.Context, folderPath string) (int64, error) {
	coll := r.db.Collection(r.collection)

	// 标准化路径格式（确保以反斜杠结尾）
	normalizedFolderPath := strings.Replace(folderPath, "/", "\\", -1)
	if !strings.HasSuffix(normalizedFolderPath, "\\") {
		normalizedFolderPath += "\\"
	}

	// 构建精确匹配library_path的正则表达式
	regexPattern := regexp.QuoteMeta(normalizedFolderPath)
	filter := bson.M{
		"library_path": bson.M{
			"$regex":   "^" + regexPattern,
			"$options": "i", // 不区分大小写
		},
	}

	// 执行删除操作
	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文件夹内容失败: %w", err)
	}

	return result, nil
}

// DeleteAll 删除所有CUE媒体文件
func (r *mediaFileCueRepository) DeleteAll(ctx context.Context) (int64, error) {
	coll := r.db.Collection(r.collection)
	result, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("删除所有CUE媒体文件失败: %w", err)
	}
	return result, nil
}

func (r *mediaFileCueRepository) GetAll(ctx context.Context) ([]*scene_audio_db_models.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("获取所有CUE媒体文件失败: %w", err)
	}
	defer cursor.Close(ctx)

	var files []*scene_audio_db_models.MediaFileCueMetadata
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("解析CUE媒体文件失败: %w", err)
	}
	return files, nil
}

// GetByID 根据ID获取CUE文件 (保持原样)
func (r *mediaFileCueRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"_id": id})

	var file scene_audio_db_models.MediaFileCueMetadata
	if err := result.Decode(&file); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("获取CUE文件失败: %w", err)
	}
	return &file, nil
}

// GetByPath 根据路径获取CUE文件 (保持原样)
func (r *mediaFileCueRepository) GetByPath(ctx context.Context, path string) (*scene_audio_db_models.MediaFileCueMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"path": path})

	var file scene_audio_db_models.MediaFileCueMetadata
	if err := result.Decode(&file); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("根据路径获取CUE文件失败: %w", err)
	}
	return &file, nil
}

func (r *mediaFileCueRepository) GetByFolder(ctx context.Context, folderPath string) ([]string, error) {
	coll := r.db.Collection(r.collection)

	// 标准化路径格式（确保以反斜杠结尾）
	normalizedFolderPath := strings.Replace(folderPath, "/", "\\", -1)
	if !strings.HasSuffix(normalizedFolderPath, "\\") {
		normalizedFolderPath += "\\"
	}

	// 构建精确匹配library_path的正则表达式
	regexPattern := regexp.QuoteMeta(normalizedFolderPath)
	filter := bson.M{
		"library_path": bson.M{
			"$regex":   "^" + regexPattern,
			"$options": "i", // 不区分大小写
		},
	}

	// 只返回path字段
	opts := options.Find().SetProjection(bson.M{"path": 1})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询文件夹内容失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 提取路径结果
	var results []string
	for cursor.Next(ctx) {
		var item struct {
			Path string `bson:"path"`
		}
		if err := cursor.Decode(&item); err != nil {
			log.Printf("解码路径失败: %v", err)
			continue
		}
		results = append(results, item.Path)
	}

	return results, nil
}

// UpdateByID 根据ID更新CUE文件 (使用原子操作)
func (r *mediaFileCueRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	coll := r.db.Collection(r.collection)

	// 使用$currentDate自动设置更新时间
	if _, exists := update["$set"]; exists {
		update["$currentDate"] = bson.M{"updated_at": true}
	} else {
		update["$set"] = bson.M{}
		update["$currentDate"] = bson.M{"updated_at": true}
	}

	result, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
		options.Update().SetUpsert(false),
	)

	if err != nil {
		return false, fmt.Errorf("CUE文件更新失败: %w", err)
	}

	return result.ModifiedCount > 0, nil
}

// 重构计数方法使用聚合管道
func (r *mediaFileCueRepository) MediaCueCountByArtist(ctx context.Context, artistID string) (int64, error) {
	coll := r.db.Collection(r.collection)

	pipeline := []bson.M{
		{"$match": bson.M{"performer_id": artistID}},
		{"$count": "count"},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家CUE文件数量失败: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct{ Count int64 }
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Count, nil
	}
	return 0, nil
}

// 使用聚合管道统计嘉宾数量
func (r *mediaFileCueRepository) GuestMediaCueCountByArtist(ctx context.Context, artistID string) (int64, error) {
	coll := r.db.Collection(r.collection)

	pipeline := []bson.M{
		{"$match": bson.M{
			"performer_id":                  bson.M{"$ne": artistID},
			"cue_tracks.track_performer_id": artistID,
		}},
		{"$count": "count"},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家嘉宾CUE文件失败: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct{ Count int64 }
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Count, nil
	}
	return 0, nil
}

// 使用聚合管道统计专辑数量
func (r *mediaFileCueRepository) MediaCountByAlbum(ctx context.Context, albumID string) (int64, error) {
	coll := r.db.Collection(r.collection)

	pipeline := []bson.M{
		{"$match": bson.M{"catalog": albumID}},
		{"$count": "count"},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("统计专辑CUE文件数量失败: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct{ Count int64 }
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Count, nil
	}
	return 0, nil
}

func (r *mediaFileCueRepository) inspectMediaCue(
	ctx context.Context,
	filter bson.M,
	filePaths []string,
	clean bool, // 是否清理无效路径
) (int, error) {
	// 构建全局有效文件路径集合，使用正斜杠规范化
	validFilePaths := make(map[string]bool)
	for _, path := range filePaths {
		// 规范化文件路径，使用正斜杠格式，与数据库中存储的格式一致
		normalizedPath := filepath.ToSlash(filepath.Clean(path))
		validFilePaths[normalizedPath] = true
	}

	// 没有有效路径时返回0，表示没有无效文件，避免误删
	if len(validFilePaths) == 0 {
		log.Printf("inspectMediaCue: filePaths为空，返回0表示没有无效文件")
		return 0, nil
	}

	coll := r.db.Collection(r.collection)

	// 批量获取数据库中所有CUE文件的路径（仅获取ID和Path字段）
	allFilesCur, err := coll.Find(
		ctx,
		filter,
		options.Find().SetProjection(bson.M{"_id": 1, "path": 1}),
	)
	if err != nil {
		return 0, fmt.Errorf("查询所有CUE文档失败: %w", err)
	}
	defer allFilesCur.Close(ctx)

	// 批量收集无效文档ID和路径
	var toDelete []primitive.ObjectID
	var invalidPaths []string

	// 检查文件路径是否有效
	for allFilesCur.Next(ctx) {
		var doc struct {
			ID   primitive.ObjectID `bson:"_id"`
			Path string             `bson:"path"`
		}
		if err := allFilesCur.Decode(&doc); err != nil {
			continue // 跳过错误项
		}

		// 检查文件路径是否在有效文件路径集合中，或者是否实际存在
		isValid := false

		// 检查文件是否在有效文件路径集合中
		if validFilePaths[doc.Path] {
			isValid = true
		} else {
			// 检查文件是否实际存在于文件系统中
			absPath := doc.Path
			// 将正斜杠转换为反斜杠，适配Windows系统
			absPath = strings.ReplaceAll(absPath, "/", "\\")
			if _, err := os.Stat(absPath); err == nil {
				// 文件实际存在，是有效文件
				isValid = true
			}
		}

		// 如果文件确实无效，添加到删除列表
		if !isValid {
			if clean {
				toDelete = append(toDelete, doc.ID)
			}
			invalidPaths = append(invalidPaths, doc.Path)
		}
	}

	// 只有当确实有无效文件时才记录日志
	if len(invalidPaths) > 0 {
		// 优化日志输出，使用更简洁的格式，避免过多信息
		log.Printf("检测到 %d 个无效CUE媒体项，示例: %v...",
			len(invalidPaths),
			invalidPaths[:min(3, len(invalidPaths))])
	} else {
		return 0, nil // 没有无效项
	}

	// 批量删除无效文档
	if clean {
		if len(toDelete) > 0 {
			batchSize := 1000
			totalDeleted := 0

			for i := 0; i < len(toDelete); i += batchSize {
				end := i + batchSize
				if end > len(toDelete) {
					end = len(toDelete)
				}

				batch := toDelete[i:end]
				_, err = coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": batch}})
				if err != nil {
					log.Printf("部分删除失败: %v", err)
				} else {
					totalDeleted += len(batch)
				}
			}
			return totalDeleted, nil
		}
	} else {
		if len(invalidPaths) > 0 {
			return len(invalidPaths), nil
		}
	}
	return 0, nil // 没有无效项
}

func (r *mediaFileCueRepository) InspectMediaCueCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
) (int, error) {
	objID, err := primitive.ObjectIDFromHex(artistID)
	if err != nil {
		return 0, fmt.Errorf("无效的艺术家ID格式: %w", err)
	}

	filter := bson.M{"performer_id": objID}
	return r.inspectMediaCue(ctx, filter, filePaths, false)
}

func (r *mediaFileCueRepository) InspectGuestMediaCueCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
) (int, error) {
	objID, err := primitive.ObjectIDFromHex(artistID)
	if err != nil {
		return 0, fmt.Errorf("无效的艺术家ID格式: %w", err)
	}

	filter := bson.M{
		"performer_id":                  bson.M{"$ne": objID},
		"cue_tracks.track_performer_id": objID,
	}
	return r.inspectMediaCue(ctx, filter, filePaths, false)
}
