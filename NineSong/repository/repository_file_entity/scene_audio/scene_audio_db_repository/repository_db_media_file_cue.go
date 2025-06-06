package scene_audio_db_repository

import (
	"context"
	"errors"
	"fmt"
	"log"
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

func (r *mediaFileCueRepository) DeleteAllInvalid(ctx context.Context, folderPath string) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{
		"$or": []bson.M{
			{"artist_id": folderPath},
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家单曲数量失败: %w", err)
	}

	return count, nil
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

// 重构Inspect方法使用统一逻辑
func (r *mediaFileCueRepository) inspectMediaCue(
	ctx context.Context,
	filter bson.M,
	filePaths []string,
	folderPath string,
) (int, error) {
	// 构建有效路径集合（只包含指定目录下的文件）
	pathSet := make(map[string]struct{})
	for _, path := range filePaths {
		if strings.HasPrefix(path, folderPath) {
			pathSet[path] = struct{}{}
		}
	}

	if pathSet != nil && len(pathSet) > 0 {
		coll := r.db.Collection(r.collection)

		// 添加目录路径过滤条件
		filter["path"] = bson.M{"$regex": "^" + regexp.QuoteMeta(folderPath)}

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

			if _, exists := pathSet[doc.Path]; !exists {
				toDelete = append(toDelete, doc.ID)
				invalidPaths = append(invalidPaths, doc.Path)
			}
		}

		// 记录无效路径日志
		if len(invalidPaths) > 0 {
			log.Printf("在目录[%s]下检测到 %d 个无效媒体项: %v...",
				folderPath,
				len(invalidPaths),
				invalidPaths[:takeMin(5, len(invalidPaths))])
		}

		// 批量删除无效文档
		if len(toDelete) > 0 {
			batchSize := 1000
			for i := 0; i < len(toDelete); i += batchSize {
				end := i + batchSize
				if end > len(toDelete) {
					end = len(toDelete)
				}

				_, err = coll.DeleteMany(
					ctx,
					bson.M{"_id": bson.M{"$in": toDelete[i:end]}},
				)
				if err != nil {
					log.Printf("部分删除失败: %v", err)
				}
			}
			return len(toDelete), nil
		}
	}
	return 0, nil
}

func (r *mediaFileCueRepository) InspectMediaCueCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
	folderPath string,
) (int, error) {
	filter := bson.M{"performer_id": artistID}
	return r.inspectMediaCue(ctx, filter, filePaths, folderPath)
}

func (r *mediaFileCueRepository) InspectGuestMediaCueCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
	folderPath string,
) (int, error) {
	filter := bson.M{
		"performer_id":                  bson.M{"$ne": artistID},
		"cue_tracks.track_performer_id": artistID,
	}
	return r.inspectMediaCue(ctx, filter, filePaths, folderPath)
}
