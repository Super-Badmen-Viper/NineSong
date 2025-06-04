package scene_audio_db_repository

import (
	"context"
	"errors"
	"fmt"
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

// Upsert 创建或更新CUE文件元数据
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

// DeleteByID 根据ID删除CUE文件
func (r *mediaFileCueRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("删除CUE文件失败: %w", err)
	}
	return nil
}

// DeleteByPath 根据路径删除CUE文件
func (r *mediaFileCueRepository) DeleteByPath(ctx context.Context, path string) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"path": path})
	if err != nil {
		return fmt.Errorf("根据路径删除CUE文件失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取CUE文件
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

// GetByPath 根据路径获取CUE文件
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

// UpdateByID 根据ID更新CUE文件
func (r *mediaFileCueRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	coll := r.db.Collection(r.collection)

	// 添加更新时间戳
	if _, exists := update["$set"]; exists {
		update["$set"].(bson.M)["updated_at"] = time.Now().UTC()
	} else {
		update["$set"] = bson.M{"updated_at": time.Now().UTC()}
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

func (r *mediaFileCueRepository) MediaCueCountByArtist(ctx context.Context, artistID string) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{
		"$or": []bson.M{
			{"performer_id": artistID}, // 主表演者
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家CUE文件数量失败: %w", err)
	}

	return count, nil
}

func (r *mediaFileCueRepository) GuestMediaCueCountByArtist(ctx context.Context, artistID string) (int64, error) {
	coll := r.db.Collection(r.collection)

	// 复合查询条件：排除主表演者 + 匹配嘉宾表演者
	filter := bson.M{
		"$and": []bson.M{
			{"performer_id": bson.M{"$ne": artistID}},   // 排除主表演者[3,5](@ref)
			{"cue_tracks.track_performer_id": artistID}, // 匹配嘉宾表演者[1,4](@ref)
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家嘉宾CUE文件失败: %w", err)
	}

	return count, nil
}

// MediaCountByAlbum 统计专辑相关的CUE文件数量
func (r *mediaFileCueRepository) MediaCountByAlbum(ctx context.Context, albumID string) (int64, error) {
	coll := r.db.Collection(r.collection)

	// 假设专辑ID存储在CATALOG字段中
	filter := bson.M{"catalog": albumID}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计专辑CUE文件数量失败: %w", err)
	}

	return count, nil
}
