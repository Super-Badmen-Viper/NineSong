package scene_audio_db_repository

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type albumRepository struct {
	db         mongo.Database
	collection string
}

func NewAlbumRepository(db mongo.Database, collection string) scene_audio_db_interface.AlbumRepository {
	return &albumRepository{
		db:         db,
		collection: collection,
	}
}

func (r *albumRepository) Upsert(ctx context.Context, album *scene_audio_db_models.AlbumMetadata) error {
	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": album.ID}
	update := bson.M{"$set": album}

	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("album upsert failed: %w", err)
	}
	return nil
}

func (r *albumRepository) BulkUpsert(ctx context.Context, albums []*scene_audio_db_models.AlbumMetadata) (int, error) {
	coll := r.db.Collection(r.collection)

	var successCount int
	for _, album := range albums {
		filter := bson.M{"_id": album.ID}
		update := bson.M{"$set": album}

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

func (r *albumRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	coll := r.db.Collection(r.collection)

	// 构建原子更新操作
	result, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
		options.Update().SetUpsert(false),
	)

	if err != nil {
		return false, fmt.Errorf("专辑更新失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return false, nil
	}

	return true, nil
}

func (r *albumRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("album delete by ID failed: %w", err)
	}
	return nil
}

func (r *albumRepository) DeleteByName(ctx context.Context, name string) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		return fmt.Errorf("album delete by name failed: %w", err)
	}
	return nil
}

func (r *albumRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.AlbumMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"_id": id})

	var album scene_audio_db_models.AlbumMetadata
	if err := result.Decode(&album); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get album by ID failed: %w", err)
	}
	return &album, nil
}

func (r *albumRepository) GetByName(ctx context.Context, name string) (*scene_audio_db_models.AlbumMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"name": name})

	var album scene_audio_db_models.AlbumMetadata
	if err := result.Decode(&album); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get album by name failed: %w", err)
	}
	return &album, nil
}

func (r *albumRepository) GetByArtist(ctx context.Context, artistID string) ([]*scene_audio_db_models.AlbumMetadata, error) {
	coll := r.db.Collection(r.collection)
	filter := bson.M{
		"$or": []bson.M{
			{"artist_id": artistID},
			{"album_artist_id": artistID},
		},
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get albums by artist failed: %w", err)
	}
	defer func(cursor mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			fmt.Printf("close cursor failed: %v\n", err)
		}
	}(cursor, ctx)

	var albums []*scene_audio_db_models.AlbumMetadata
	if err := cursor.All(ctx, &albums); err != nil {
		return nil, fmt.Errorf("decode album results failed: %w", err)
	}

	return albums, nil
}

func (r *albumRepository) ResetALLField(ctx context.Context) (int64, error) {
	coll := r.db.Collection(r.collection)

	resetFields := []string{
		"song_count",
		"size",
		"duration",
	}

	update := make(bson.M)
	for _, field := range resetFields {
		update[field] = 0
	}

	result, err := coll.UpdateMany(
		ctx,
		bson.M{},
		bson.M{"$set": update},
	)

	if err != nil {
		return 0, fmt.Errorf("批量重置专辑字段失败: %w", err)
	}

	return result.ModifiedCount, nil
}

func (r *albumRepository) ResetField(
	ctx context.Context,
	field string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{}

	update := bson.M{"$set": bson.M{field: ""}}

	result, err := coll.UpdateMany(
		ctx,
		filter,
		update,
	)

	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func (r *albumRepository) UpdateCounter(
	ctx context.Context,
	albumID primitive.ObjectID,
	field string,
	increment int,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	var update bson.M
	if increment == 0 {
		update = bson.M{"$set": bson.M{field: 0}}
	} else {
		update = bson.M{"$inc": bson.M{field: increment}}
	}

	result, err := coll.UpdateByID(ctx, albumID, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func (r *albumRepository) GetByMbzID(ctx context.Context, mbzID string) (*scene_audio_db_models.AlbumMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"mbz_album_id": mbzID})

	var album scene_audio_db_models.AlbumMetadata
	if err := result.Decode(&album); err != nil {
		return nil, fmt.Errorf("通过MBID获取专辑失败: %w", err)
	}
	return &album, nil
}

func (r *albumRepository) GetByFilter(
	ctx context.Context,
	filter interface{},
) (*scene_audio_db_models.AlbumMetadata, error) {
	coll := r.db.Collection(r.collection)

	bsonFilter, ok := filter.(bson.M)
	if !ok {
		return nil, fmt.Errorf("invalid filter type: %T", filter)
	}

	result := coll.FindOne(ctx, bsonFilter)

	var album scene_audio_db_models.AlbumMetadata
	if err := result.Decode(&album); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("专辑查询失败: %w", err)
	}

	return &album, nil
}

func (r *albumRepository) GetAllIDs(ctx context.Context) ([]primitive.ObjectID, error) {
	coll := r.db.Collection(r.collection)

	// 使用聚合管道只获取ID字段
	pipeline := []bson.D{
		{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1}, // 只返回ID字段
			}},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("专辑ID查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("专辑ID解码失败: %w", err)
	}

	// 提取ID到切片
	ids := make([]primitive.ObjectID, len(results))
	for i, item := range results {
		ids[i] = item.ID
	}

	return ids, nil
}

func (r *albumRepository) AlbumCountByArtist(
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
		return 0, fmt.Errorf("统计艺术家专辑数量失败: %w", err)
	}

	return count, nil
}

func (r *albumRepository) GuestAlbumCountByArtist(
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
		return 0, fmt.Errorf("统计合作专辑失败: %w", err)
	}

	return count, nil
}

func (r *albumRepository) InspectAlbumMediaCountByAlbum(ctx context.Context, albumID string, operand int) (bool, error) {
	if operand == 0 {
		return false, fmt.Errorf("操作数不能为零")
	}

	coll := r.db.Collection(r.collection)

	objID, err := primitive.ObjectIDFromHex(albumID)
	if err != nil {
		return false, fmt.Errorf("无效的专辑ID格式: %w", err)
	}

	album, err := r.GetByID(ctx, objID)
	if err != nil {
		return false, fmt.Errorf("获取专辑信息失败: %w", err)
	}
	if album == nil {
		return false, fmt.Errorf("专辑不存在")
	}

	if operand > -1 {
		newCount := album.SongCount - operand
		if newCount <= 0 {
			if err := r.DeleteByID(ctx, objID); err != nil {
				return false, fmt.Errorf("删除专辑失败: %w", err)
			}
			return true, nil
		} else {
			update := bson.M{"$set": bson.M{"song_count": newCount}}
			_, err := coll.UpdateByID(
				ctx,
				objID,
				update,
			)
			if err != nil {
				return false, fmt.Errorf("更新歌曲计数失败: %w", err)
			}
			return false, nil
		}
	} else {
		if err := r.DeleteByID(ctx, objID); err != nil {
			return false, fmt.Errorf("删除专辑失败: %w", err)
		}
		return true, nil
	}
}
