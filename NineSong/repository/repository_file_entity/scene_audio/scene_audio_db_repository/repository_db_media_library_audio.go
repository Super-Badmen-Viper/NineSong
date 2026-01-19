package scene_audio_db_repository

import (
	"context"
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

type mediaLibraryAudioRepository struct {
	db         mongo.Database
	collection string
}

func NewMediaLibraryAudioRepository(db mongo.Database, collection string) scene_audio_db_interface.MediaLibraryAudioRepository {
	return &mediaLibraryAudioRepository{
		db:         db,
		collection: collection,
	}
}

// 实现 BaseRepository 的基本方法
func (r *mediaLibraryAudioRepository) Create(ctx context.Context, entity *scene_audio_db_models.MediaLibraryAudio) error {
	entity.SetTimestamps()
	_, err := r.db.Collection(r.collection).InsertOne(ctx, entity)
	return err
}

func (r *mediaLibraryAudioRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	filter := bson.M{"_id": id}

	// 确保更新操作使用 $set 操作符，并添加时间戳
	now := primitive.NewDateTimeFromTime(time.Now())
	setUpdate := bson.M{}

	// 检查传入的更新操作是否已经包含更新操作符（如 $set, $inc 等）
	hasOperator := false
	for key := range update {
		if key[0] == '$' { // 检查是否是操作符，如 $set, $inc, $unset 等
			hasOperator = true
			break
		}
	}

	if hasOperator {
		// 如果已经有操作符，需要合并 $set 操作
		for op, opValue := range update {
			if op == "$set" {
				// 如果有 $set 操作，添加时间戳
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
		// 确保有 $set 操作用于时间戳
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

	result, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, finalUpdate)
	if err != nil {
		return false, err
	}

	return result.MatchedCount > 0, nil
}

func (r *mediaLibraryAudioRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.db.Collection(r.collection).DeleteOne(ctx, filter)
	return err
}

func (r *mediaLibraryAudioRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.MediaLibraryAudio, error) {
	var entity scene_audio_db_models.MediaLibraryAudio
	filter := bson.M{"_id": id}
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *mediaLibraryAudioRepository) FindAll(ctx context.Context, page, limit int64) ([]*scene_audio_db_models.MediaLibraryAudio, error) {
	var entities []*scene_audio_db_models.MediaLibraryAudio

	opts := options.Find().
		SetSkip((page - 1) * limit).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.db.Collection(r.collection).Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var entity scene_audio_db_models.MediaLibraryAudio
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// 上传下载相关方法
func (r *mediaLibraryAudioRepository) GetByLibraryID(ctx context.Context, libraryID primitive.ObjectID) ([]*scene_audio_db_models.MediaLibraryAudio, error) {
	var entities []*scene_audio_db_models.MediaLibraryAudio

	filter := bson.M{"library_id": libraryID}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var entity scene_audio_db_models.MediaLibraryAudio
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

func (r *mediaLibraryAudioRepository) GetByUploaderID(ctx context.Context, uploaderID primitive.ObjectID) ([]*scene_audio_db_models.MediaLibraryAudio, error) {
	var entities []*scene_audio_db_models.MediaLibraryAudio

	filter := bson.M{"uploader_id": uploaderID}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var entity scene_audio_db_models.MediaLibraryAudio
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

func (r *mediaLibraryAudioRepository) GetByStatus(ctx context.Context, status string) ([]*scene_audio_db_models.MediaLibraryAudio, error) {
	var entities []*scene_audio_db_models.MediaLibraryAudio

	filter := bson.M{"status": status}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var entity scene_audio_db_models.MediaLibraryAudio
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

func (r *mediaLibraryAudioRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status, "updated_at": primitive.NewDateTimeFromTime(time.Now())}}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *mediaLibraryAudioRepository) UpdateFilePath(ctx context.Context, id primitive.ObjectID, filePath string) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"file_path": filePath, "updated_at": primitive.NewDateTimeFromTime(time.Now())}}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *mediaLibraryAudioRepository) GetByChecksum(ctx context.Context, checksum string) (*scene_audio_db_models.MediaLibraryAudio, error) {
	var entity scene_audio_db_models.MediaLibraryAudio
	filter := bson.M{"checksum": checksum}
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *mediaLibraryAudioRepository) GetByFileName(ctx context.Context, fileName string) (*scene_audio_db_models.MediaLibraryAudio, error) {
	var entity scene_audio_db_models.MediaLibraryAudio
	filter := bson.M{"file_name": fileName}
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *mediaLibraryAudioRepository) IncrementDownloadCount(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$inc": bson.M{"download_count": 1}, "$set": bson.M{"updated_at": primitive.NewDateTimeFromTime(time.Now())}}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

// BulkUpsert 批量创建/更新媒体库音频文件
func (r *mediaLibraryAudioRepository) BulkUpsert(ctx context.Context, files []*scene_audio_db_models.MediaLibraryAudio) (int, error) {
	coll := r.db.Collection(r.collection)
	var successCount int

	for _, file := range files {
		now := primitive.NewDateTimeFromTime(time.Now())
		if file.ID.IsZero() {
			file.CreatedAt = now
		}
		file.UpdatedAt = now

		filter := bson.M{"_id": file.ID}
		update := bson.M{"$set": file}

		_, err := coll.UpdateOne(
			ctx,
			filter,
			update,
		)
		if err != nil {
			continue // 继续处理下一个文件，而不是返回错误
		}
		successCount++
	}

	return successCount, nil
}

// GetByID 根据ID获取实体
func (r *mediaLibraryAudioRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.MediaLibraryAudio, error) {
	return r.FindByID(ctx, id)
}

// Upsert 插入或更新实体
func (r *mediaLibraryAudioRepository) Upsert(ctx context.Context, entity *scene_audio_db_models.MediaLibraryAudio) error {
	entity.SetTimestamps()
	filter := bson.M{"_id": entity.ID}
	update := bson.M{"$set": entity}
	opts := options.Update().SetUpsert(true)
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update, opts)
	return err
}

// CreateMany 批量创建实体
func (r *mediaLibraryAudioRepository) CreateMany(ctx context.Context, entities []*scene_audio_db_models.MediaLibraryAudio) error {
	var docs []interface{}
	for _, entity := range entities {
		entity.SetTimestamps()
		docs = append(docs, entity)
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := r.db.Collection(r.collection).InsertMany(ctx, docs)
	return err
}

// DeleteMany 批量删除实体
func (r *mediaLibraryAudioRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	result, err := r.db.Collection(r.collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	// The DeleteMany method returns an int64 directly in the wrapper
	return result, nil
}

// GetAll 获取所有实体
func (r *mediaLibraryAudioRepository) GetAll(ctx context.Context) ([]*scene_audio_db_models.MediaLibraryAudio, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*scene_audio_db_models.MediaLibraryAudio
	for cursor.Next(ctx) {
		var entity scene_audio_db_models.MediaLibraryAudio
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// GetByFilter 根据过滤条件获取实体列表
func (r *mediaLibraryAudioRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*scene_audio_db_models.MediaLibraryAudio, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*scene_audio_db_models.MediaLibraryAudio
	for cursor.Next(ctx) {
		var entity scene_audio_db_models.MediaLibraryAudio
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// GetOneByFilter 根据过滤条件获取单个实体
func (r *mediaLibraryAudioRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*scene_audio_db_models.MediaLibraryAudio, error) {
	var entity scene_audio_db_models.MediaLibraryAudio
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// Count 统计符合条件的实体数量
func (r *mediaLibraryAudioRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetPaginated 分页获取实体
func (r *mediaLibraryAudioRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*scene_audio_db_models.MediaLibraryAudio, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.db.Collection(r.collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*scene_audio_db_models.MediaLibraryAudio
	for cursor.Next(ctx) {
		var entity scene_audio_db_models.MediaLibraryAudio
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// UpdateMany 批量更新实体
func (r *mediaLibraryAudioRepository) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*driver.UpdateResult, error) {
	return r.db.Collection(r.collection).UpdateMany(ctx, filter, update, opts...)
}

// Exists 检查实体是否存在
func (r *mediaLibraryAudioRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": id}
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByFilter 根据过滤条件检查实体是否存在
func (r *mediaLibraryAudioRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
