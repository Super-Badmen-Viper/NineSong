package scene_audio_db_repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type mediaFileRepository struct {
	db         mongo.Database
	collection string
}

func NewMediaFileRepository(db mongo.Database, collection string) scene_audio_db_interface.MediaFileRepository {
	return &mediaFileRepository{
		db:         db,
		collection: collection,
	}
}

func (r *mediaFileRepository) Upsert(ctx context.Context, file *scene_audio_db_models.MediaFileMetadata) (*scene_audio_db_models.MediaFileMetadata, error) {
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
		return nil, fmt.Errorf("upsert操作失败: %w", err)
	}

	if result.UpsertedID != nil {
		file.ID = result.UpsertedID.(primitive.ObjectID)
		file.CreatedAt = now
	} else {
		var existing struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		err := coll.FindOne(
			ctx,
			bson.M{"path": file.Path},
		).Decode(&existing)

		if err == nil {
			file.ID = existing.ID
		} else if errors.Is(err, driver.ErrNoDocuments) {
			return nil, fmt.Errorf("文档既未插入也未更新: %w", err)
		} else {
			return nil, fmt.Errorf("ID查询失败: %w", err)
		}
	}

	file.UpdatedAt = now
	return file, nil
}

func (r *mediaFileRepository) BulkUpsert(ctx context.Context, files []*scene_audio_db_models.MediaFileMetadata) (int, error) {
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

func (r *mediaFileRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete media file failed: %w", err)
	}
	return nil
}

func (r *mediaFileRepository) DeleteByPath(ctx context.Context, path string) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"path": path})
	if err != nil {
		return fmt.Errorf("delete by path failed: %w", err)
	}
	return nil
}

func (r *mediaFileRepository) DeleteAllInvalid(
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

	// 场景1：全量删除（无folderPath过滤）
	if len(filePaths) == 0 {
		// 直接删除所有文档
		delResult, err := coll.DeleteMany(ctx, bson.M{})
		if err != nil {
			return 0, deletedArtists, fmt.Errorf("全量删除失败: %w", err)
		}

		// 查询所有艺术家ID及其计数（需单独统计）
		artistCounts := make(map[primitive.ObjectID]int64)
		cur, err := coll.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"artist_id": 1}))
		if err == nil {
			defer cur.Close(ctx)
			for cur.Next(ctx) {
				var doc struct {
					ArtistID primitive.ObjectID `bson:"artist_id"`
				}
				if err := cur.Decode(&doc); err == nil {
					artistCounts[doc.ArtistID]++
				}
			}
		}

		// 构建艺术家统计
		for artistID, count := range artistCounts {
			deletedArtists = append(deletedArtists, struct {
				ArtistID primitive.ObjectID
				Count    int64
			}{ArtistID: artistID, Count: count})
		}

		return delResult, deletedArtists, nil
	}

	// 场景2：路径比对删除（全局处理）
	validFilePaths := make(map[string]struct{})
	for _, path := range filePaths {
		cleanPath := filepath.Clean(path)
		validFilePaths[cleanPath] = struct{}{}
	}

	// 查询所有文档（移除folderPath过滤）
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

		cleanPath := filepath.Clean(doc.Path)
		if _, valid := validFilePaths[cleanPath]; !valid {
			toDelete = append(toDelete, doc.ID)
			artistToIDs[doc.ArtistID] = append(artistToIDs[doc.ArtistID], doc.ID)
		}
	}

	// 批量删除并统计（优化性能）
	totalDeleted := int64(0)
	const batchSize = 1000
	artistCounts := make(map[primitive.ObjectID]int64)

	// 使用预统计避免嵌套循环
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

func (r *mediaFileRepository) DeleteByFolder(ctx context.Context, folderPath string) (int64, error) {
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

func (r *mediaFileRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.MediaFileMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"_id": id})

	var file scene_audio_db_models.MediaFileMetadata
	if err := result.Decode(&file); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get media file failed: %w", err)
	}
	return &file, nil
}

func (r *mediaFileRepository) GetByPath(ctx context.Context, path string) (*scene_audio_db_models.MediaFileMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"path": path})

	var file scene_audio_db_models.MediaFileMetadata
	if err := result.Decode(&file); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get by path failed: %w", err)
	}
	return &file, nil
}

func (r *mediaFileRepository) GetByFolder(ctx context.Context, folderPath string) ([]string, error) {
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

func (r *mediaFileRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	coll := r.db.Collection(r.collection)

	// 构建原子更新操作
	result, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
		options.Update().SetUpsert(false),
	)

	if err != nil {
		return false, fmt.Errorf("媒体文件更新失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return false, nil
	}

	return true, nil
}

func (r *mediaFileRepository) MediaCountByArtist(
	ctx context.Context,
	artistID string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{
		"$or": []bson.M{
			{"artist_id": artistID},
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家单曲数量失败: %w", err)
	}

	return count, nil
}

func (r *mediaFileRepository) GuestMediaCountByArtist(
	ctx context.Context,
	artistID string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	// 构造复合查询条件
	filter := bson.M{
		"$and": []bson.M{
			{"artist_id": bson.M{"$ne": artistID}}, // 排除主导者[3](@ref)
			{"all_artist_ids": bson.M{ // 匹配合作者[4](@ref)
				"$elemMatch": bson.M{
					"artist_id": artistID,
				},
			}},
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家合作单曲失败: %w", err)
	}

	return count, nil
}

func (r *mediaFileRepository) MediaCountByAlbum(
	ctx context.Context,
	albumID string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{
		"$or": []bson.M{
			{"album_id": albumID},
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计专辑单曲数量失败: %w", err)
	}

	return count, nil
}

func takeMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *mediaFileRepository) inspectMedia(
	ctx context.Context,
	filter bson.M,
	validFilePaths []string, // 全局有效路径集合
	collection string,
	clean bool, // 是否清理无效路径
) (int, error) {
	// 构建全局有效路径集合
	validSet := make(map[string]struct{})
	for _, path := range validFilePaths {
		cleanPath := filepath.Clean(path)
		validSet[cleanPath] = struct{}{}
	}

	// 没有有效路径时直接返回
	if len(validSet) == 0 {
		return -1, nil
	}

	coll := r.db.Collection(collection)

	// 精确查询条件（双重过滤）
	invalidDocsFilter := bson.M{
		"$and": []bson.M{
			filter,                                   // 原始过滤条件
			{"path": bson.M{"$nin": validFilePaths}}, // 路径不在全局有效列表中
		},
	}

	// 直接获取需要删除的文档ID
	cur, err := coll.Find(
		ctx,
		invalidDocsFilter,
		options.Find().SetProjection(bson.M{"_id": 1, "path": 1}),
	)
	if err != nil {
		return 0, fmt.Errorf("查询无效文档失败: %w", err)
	}
	defer cur.Close(ctx)

	// 批量收集无效文档ID
	var toDelete []primitive.ObjectID
	var invalidPaths []string
	for cur.Next(ctx) {
		var doc struct {
			ID   primitive.ObjectID `bson:"_id"`
			Path string             `bson:"path"`
		}
		if err := cur.Decode(&doc); err != nil {
			continue
		}
		if clean {
			toDelete = append(toDelete, doc.ID)
		}
		invalidPaths = append(invalidPaths, doc.Path)
	}

	// 记录无效路径日志（全局处理）
	if len(invalidPaths) > 0 {
		log.Printf("检测到 %d 个无效媒体项: %v...",
			len(invalidPaths),
			invalidPaths[:min(5, len(invalidPaths))])
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
				delResult, err := coll.DeleteMany(
					ctx,
					bson.M{"_id": bson.M{"$in": batch}},
				)
				if err != nil {
					log.Printf("部分删除失败: %v", err)
				} else {
					totalDeleted += int(delResult)
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

func (r *mediaFileRepository) InspectMediaCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
) (int, error) {
	filter := bson.M{"artist_id": artistID}
	return r.inspectMedia(ctx, filter, filePaths, r.collection, false)
}

func (r *mediaFileRepository) InspectGuestMediaCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
) (int, error) {
	filter := bson.M{
		"artist_id":      bson.M{"$ne": artistID},
		"all_artist_ids": bson.M{"$elemMatch": bson.M{"artist_id": artistID}},
	}
	return r.inspectMedia(ctx, filter, filePaths, r.collection, false)
}

func (r *mediaFileRepository) InspectMediaCountByAlbum(
	ctx context.Context,
	albumID string,
	filePaths []string,
) (int, error) {
	filter := bson.M{"album_id": albumID}
	return r.inspectMedia(ctx, filter, filePaths, r.collection, false)
}

func (r *mediaFileRepository) InspectGuestMediaCountByAlbum(
	ctx context.Context,
	artistID string,
	filePaths []string,
) (int, error) {
	filter := bson.M{
		"artist_id":            bson.M{"$ne": artistID},
		"all_album_artist_ids": bson.M{"$elemMatch": bson.M{"artist_id": artistID}},
	}
	return r.inspectMedia(ctx, filter, filePaths, r.collection, false)
}
