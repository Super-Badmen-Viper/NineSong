package scene_audio_db_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	scene_audio_repository_interface "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_repository_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
)

// mediaLibrarySyncRecordRepository 媒体库同步记录仓库实现
type mediaLibrarySyncRecordRepository struct {
	db         mongo.Database
	collection string
}

// NewMediaLibrarySyncRecordRepository 创建新的媒体库同步记录仓库实例
func NewMediaLibrarySyncRecordRepository(db mongo.Database, collection string) scene_audio_repository_interface.MediaLibrarySyncRecordRepository {
	return &mediaLibrarySyncRecordRepository{
		db:         db,
		collection: collection,
	}
}

// Create 创建媒体库同步记录
func (r *mediaLibrarySyncRecordRepository) Create(ctx context.Context, entity *scene_audio_db_models.MediaLibrarySyncRecord) error {
	entity.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	entity.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	coll := r.db.Collection(r.collection)
	_, err := coll.InsertOne(ctx, entity)
	return err
}

// UpdateByID 根据ID更新媒体库同步记录
func (r *mediaLibrarySyncRecordRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, entity *scene_audio_db_models.MediaLibrarySyncRecord) error {
	entity.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": entity}

	_, err := coll.UpdateOne(ctx, filter, update)
	return err
}

// DeleteByID 根据ID删除媒体库同步记录
func (r *mediaLibrarySyncRecordRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": id}

	_, err := coll.DeleteOne(ctx, filter)
	return err
}

// GetByID 根据ID获取媒体库同步记录
func (r *mediaLibrarySyncRecordRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.MediaLibrarySyncRecord, error) {
	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": id}

	var entity scene_audio_db_models.MediaLibrarySyncRecord
	err := coll.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if errors.Is(err, driver.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindByMediaFileID 根据媒体文件ID查找同步记录
func (r *mediaLibrarySyncRecordRepository) FindByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (*scene_audio_db_models.MediaLibrarySyncRecord, error) {
	filter := bson.M{"media_file_id": mediaFileID}

	var entity scene_audio_db_models.MediaLibrarySyncRecord
	coll := r.db.Collection(r.collection)
	err := coll.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if errors.Is(err, driver.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindByMediaLibraryAudioID 根据媒体库音频文件ID查找同步记录
func (r *mediaLibrarySyncRecordRepository) FindByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (*scene_audio_db_models.MediaLibrarySyncRecord, error) {
	filter := bson.M{"library_audio_id": mediaLibraryAudioID}

	var entity scene_audio_db_models.MediaLibrarySyncRecord
	coll := r.db.Collection(r.collection)
	err := coll.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if errors.Is(err, driver.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindBySyncStatus 根据同步状态查找同步记录
func (r *mediaLibrarySyncRecordRepository) FindBySyncStatus(ctx context.Context, syncStatus string) ([]*scene_audio_db_models.MediaLibrarySyncRecord, error) {
	filter := bson.M{"sync_status": syncStatus}

	var entities []*scene_audio_db_models.MediaLibrarySyncRecord
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	return entities, nil
}

// FindBySyncType 根据同步类型查找同步记录
func (r *mediaLibrarySyncRecordRepository) FindBySyncType(ctx context.Context, syncType string) ([]*scene_audio_db_models.MediaLibrarySyncRecord, error) {
	filter := bson.M{"sync_type": syncType}

	var entities []*scene_audio_db_models.MediaLibrarySyncRecord
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	return entities, nil
}

// UpdateSyncStatus 更新同步状态
func (r *mediaLibrarySyncRecordRepository) UpdateSyncStatus(ctx context.Context, id primitive.ObjectID, status string, progress float64, errorMessage string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"sync_status":   status,
			"progress":      progress,
			"error_message": errorMessage,
			"updated_at":    primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	coll := r.db.Collection(r.collection)
	_, err := coll.UpdateOne(ctx, filter, update)
	return err
}

// UpdateSyncProgress 更新同步进度
func (r *mediaLibrarySyncRecordRepository) UpdateSyncProgress(ctx context.Context, id primitive.ObjectID, progress float64, lastUpdateTime primitive.DateTime) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"progress":         progress,
			"last_update_time": lastUpdateTime,
			"updated_at":       primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	coll := r.db.Collection(r.collection)
	_, err := coll.UpdateOne(ctx, filter, update)
	return err
}

// UpdateSyncResult 更新同步结果
func (r *mediaLibrarySyncRecordRepository) UpdateSyncResult(ctx context.Context, id primitive.ObjectID, status string, progress float64, errorMessage string, filePath string, fileSize int64, checksum string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"sync_status":   status,
			"progress":      progress,
			"error_message": errorMessage,
			"file_path":     filePath,
			"file_size":     fileSize,
			"checksum":      checksum,
			"updated_at":    primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	coll := r.db.Collection(r.collection)
	_, err := coll.UpdateOne(ctx, filter, update)
	return err
}

// DeleteByMediaFileID 根据媒体文件ID删除同步记录
func (r *mediaLibrarySyncRecordRepository) DeleteByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) error {
	filter := bson.M{"media_file_id": mediaFileID}
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteMany(ctx, filter)
	return err
}

// DeleteByMediaLibraryAudioID 根据媒体库音频文件ID删除同步记录
func (r *mediaLibrarySyncRecordRepository) DeleteByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) error {
	filter := bson.M{"library_audio_id": mediaLibraryAudioID}
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteMany(ctx, filter)
	return err
}

// ExistsByMediaFileID 检查是否已存在关联指定媒体文件ID的同步记录
func (r *mediaLibrarySyncRecordRepository) ExistsByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (bool, error) {
	filter := bson.M{"media_file_id": mediaFileID}
	coll := r.db.Collection(r.collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByMediaLibraryAudioID 检查是否已存在关联指定媒体库音频文件ID的同步记录
func (r *mediaLibrarySyncRecordRepository) ExistsByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (bool, error) {
	filter := bson.M{"library_audio_id": mediaLibraryAudioID}
	coll := r.db.Collection(r.collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetSyncStatistics 获取同步统计信息
func (r *mediaLibrarySyncRecordRepository) GetSyncStatistics(ctx context.Context) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$sync_status",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
	}

	coll := r.db.Collection(r.collection)
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate sync statistics: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode aggregation results: %w", err)
	}

	statistics := make(map[string]interface{})
	totalCount := 0

	for _, result := range results {
		statistics[result.ID] = result.Count
		totalCount += result.Count
	}

	statistics["total"] = totalCount
	return statistics, nil
}
