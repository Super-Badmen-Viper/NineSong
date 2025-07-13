package scene_audio_db_repository

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type wordCloudRepository struct {
	db         mongo.Database
	collection string
}

func NewWordCloudRepository(db mongo.Database, collection string) scene_audio_db_interface.WordCloudRepository {
	return &wordCloudRepository{
		db:         db,
		collection: collection,
	}
}

func (w *wordCloudRepository) DropAllIndex(ctx context.Context) error {
	coll := w.db.Collection(w.collection)

	// 1. 检测索引是否存在（排除默认_id索引）
	indexes, err := coll.Indexes().ListSpecifications(ctx)
	if err != nil {
		return fmt.Errorf("索引查询失败: %w", err)
	}

	// 2. 若仅有默认_id索引，直接返回（避免无效删除）
	if len(indexes) <= 1 { // 默认_id索引始终存在[4](@ref)
		return nil
	}

	// 3. 存在非默认索引时才执行删除
	_, err = coll.Indexes().DropAll(ctx)
	if err != nil {
		return fmt.Errorf("索引删除失败: %w", err)
	}
	return nil
}

func (w *wordCloudRepository) CreateIndex(ctx context.Context, fieldName string, unique bool) error {
	coll := w.db.Collection(w.collection)
	indexModel := driver.IndexModel{
		Keys:    bson.D{{Key: fieldName, Value: 1}},
		Options: options.Index().SetUnique(unique).SetBackground(true),
	}
	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("索引创建失败(%s): %w", fieldName, err)
	}
	return nil
}

// 数据操作方法
func (w *wordCloudRepository) BulkUpsert(ctx context.Context, files []*scene_audio_db_models.WordCloudMetadata) (int, error) {
	coll := w.db.Collection(w.collection)
	bulk := coll.BulkWrite()

	for _, file := range files {
		filter := bson.M{"_id": file.ID}
		model := driver.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(bson.M{"$set": file}).
			SetUpsert(true)
		bulk.AddModel(model)
	}

	result, err := bulk.Execute(ctx)
	if err != nil {
		return 0, fmt.Errorf("批量更新失败: %w", err)
	}

	return int(result.UpsertedCount() + result.ModifiedCount()), nil
}

func (w *wordCloudRepository) AllDelete(ctx context.Context) error {
	coll := w.db.Collection(w.collection)
	_, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("全表清空失败: %w", err)
	}
	return nil
}

func (w *wordCloudRepository) GetAll(ctx context.Context) ([]*scene_audio_db_models.WordCloudMetadata, error) {
	coll := w.db.Collection(w.collection)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("数据查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*scene_audio_db_models.WordCloudMetadata
	for cursor.Next(ctx) {
		var item scene_audio_db_models.WordCloudMetadata
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("数据解码失败: %w", err)
		}
		results = append(results, &item)
	}

	return results, nil
}
