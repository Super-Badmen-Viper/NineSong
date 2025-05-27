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

func (r *albumRepository) AlbumCountByArtist(
	ctx context.Context,
	artistID string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{
		"$or": []bson.M{
			{"artist_id": artistID},
			{"album_artist_id": artistID},
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家专辑数量失败: %w", err)
	}

	return count, nil
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

func (r *albumRepository) GetByMBID(ctx context.Context, mbzID string) (*scene_audio_db_models.AlbumMetadata, error) {
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
