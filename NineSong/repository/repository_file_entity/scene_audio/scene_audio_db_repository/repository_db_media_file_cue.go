package scene_audio_db_repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
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

	// 场景1：全量删除（无需folderPath过滤）
	if len(filePaths) == 0 {
		// 直接删除所有文档
		delResult, err := coll.DeleteMany(ctx, bson.M{})
		if err != nil {
			return 0, deletedArtists, fmt.Errorf("全量删除失败: %w", err)
		}
		return delResult, deletedArtists, nil
	}

	// 场景2：路径比对删除（全局处理）
	validFilePaths := make(map[string]struct{})
	for _, rawPath := range filePaths {
		cleanPath := filepath.Clean(rawPath)
		validFilePaths[cleanPath] = struct{}{}
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

		cleanPath := filepath.Clean(doc.Path)
		if _, valid := validFilePaths[cleanPath]; !valid {
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
	// 构建全局有效路径集合
	pathSet := make(map[string]struct{})
	for _, path := range filePaths {
		cleanPath := filepath.Clean(path)
		pathSet[cleanPath] = struct{}{}
	}

	// 没有有效路径时直接返回
	if len(pathSet) == 0 {
		return -1, nil
	}

	coll := r.db.Collection(r.collection)

	// 查询所有可能的文档（移除folderPath过滤）
	cur, err := coll.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1, "path": 1}))
	if err != nil {
		return 0, fmt.Errorf("查询失败: %w", err)
	}
	defer cur.Close(ctx)

	var toDelete []primitive.ObjectID
	var invalidPaths []string

	for cur.Next(ctx) {
		var doc struct {
			ID   primitive.ObjectID `bson:"_id"`
			Path string             `bson:"path"`
		}
		if err := cur.Decode(&doc); err != nil {
			continue // 跳过错误项
		}

		cleanPath := filepath.Clean(doc.Path)
		if _, exists := pathSet[cleanPath]; !exists {
			if clean {
				toDelete = append(toDelete, doc.ID)
			}
			invalidPaths = append(invalidPaths, doc.Path)
		}
	}

	// 记录无效路径日志
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
