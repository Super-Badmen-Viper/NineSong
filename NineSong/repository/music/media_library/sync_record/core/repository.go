package core
import domainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/sync_record/core"

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo"

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
func NewMediaLibrarySyncRecordRepository(db mongo.Database, collection string) domainCore.MediaLibrarySyncRecordRepository {
	return &mediaLibrarySyncRecordRepository{
		db:         db,
		collection: collection,
	}
}

// Create 创建媒体库同步记录
func (r *mediaLibrarySyncRecordRepository) Create(ctx context.Context, entity *domainCore.MediaLibrarySyncRecord) error {
	entity.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	entity.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	coll := r.db.Collection(r.collection)
	_, err := coll.InsertOne(ctx, entity)
	return err
}

// UpdateByID 根据ID更新媒体库同步记录
func (r *mediaLibrarySyncRecordRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	filter := bson.M{"_id": id}
	
	// 确保更新操作使用 $set 操作符，并添加时间戳
	now := primitive.NewDateTimeFromTime(time.Now())
	setUpdate := bson.M{}
	
	// 检查传入的更新操作是否已经包含更新操作符
	hasOperator := false
	for key := range update {
		if key[0] == '$' {
			hasOperator = true
			break
		}
	}
	
	if hasOperator {
		// 如果已经有操作符，需要合并 $set 操作
		for op, opValue := range update {
			if op == "$set" {
				if setValues, ok := opValue.(bson.M); ok {
					setValues["updated_at"] = now
					setUpdate[op] = setValues
				} else {
					setUpdate[op] = opValue
				}
			} else {
				setUpdate[op] = opValue
			}
		}
		if _, exists := setUpdate["$set"]; !exists {
			setUpdate["$set"] = bson.M{"updated_at": now}
		}
	} else {
		// 如果没有操作符，使用 $set 并添加时间戳
		for k, v := range update {
			setUpdate[k] = v
		}
		setUpdate["updated_at"] = now
	}
	
	// 包装在 $set 中以确保一致的更新行为
	finalUpdate := bson.M{"$set": setUpdate}
	
	coll := r.db.Collection(r.collection)
	result, err := coll.UpdateOne(ctx, filter, finalUpdate)
	if err != nil {
		return false, err
	}
	
	return result.MatchedCount > 0, nil
}

// DeleteByID 根据ID删除媒体库同步记录
func (r *mediaLibrarySyncRecordRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": id}

	_, err := coll.DeleteOne(ctx, filter)
	return err
}

// GetByID 根据ID获取媒体库同步记录
func (r *mediaLibrarySyncRecordRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domainCore.MediaLibrarySyncRecord, error) {
	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": id}

	var entity domainCore.MediaLibrarySyncRecord
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
func (r *mediaLibrarySyncRecordRepository) FindByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (*domainCore.MediaLibrarySyncRecord, error) {
	filter := bson.M{"media_file_id": mediaFileID}

	var entity domainCore.MediaLibrarySyncRecord
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
func (r *mediaLibrarySyncRecordRepository) FindByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (*domainCore.MediaLibrarySyncRecord, error) {
	filter := bson.M{"media_library_audio_id": mediaLibraryAudioID}

	var entity domainCore.MediaLibrarySyncRecord
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
func (r *mediaLibrarySyncRecordRepository) FindBySyncStatus(ctx context.Context, syncStatus string) ([]*domainCore.MediaLibrarySyncRecord, error) {
	filter := bson.M{"sync_status": syncStatus}

	var entities []*domainCore.MediaLibrarySyncRecord
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
func (r *mediaLibrarySyncRecordRepository) FindBySyncType(ctx context.Context, syncType string) ([]*domainCore.MediaLibrarySyncRecord, error) {
	filter := bson.M{"sync_type": syncType}

	var entities []*domainCore.MediaLibrarySyncRecord
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
	filter := bson.M{"media_library_audio_id": mediaLibraryAudioID}
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
	filter := bson.M{"media_library_audio_id": mediaLibraryAudioID}
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
