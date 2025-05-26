package scene_audio_db_repository

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

type artistRepository struct {
	db         mongo.Database
	collection string
}

func NewArtistRepository(db mongo.Database, collection string) scene_audio_db_interface.ArtistRepository {
	return &artistRepository{
		db:         db,
		collection: collection,
	}
}

func (r *artistRepository) Upsert(ctx context.Context, artist *scene_audio_db_models.ArtistMetadata) error {
	coll := r.db.Collection(r.collection)
	filter := bson.M{"_id": artist.ID}
	update := bson.M{"$set": artist}

	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("artist upsert failed: %w", err)
	}
	return nil
}

func (r *artistRepository) BulkUpsert(ctx context.Context, artists []*scene_audio_db_models.ArtistMetadata) (int, error) {
	coll := r.db.Collection(r.collection)

	var successCount int
	for _, artist := range artists {
		filter := bson.M{"_id": artist.ID}
		update := bson.M{"$set": artist}

		_, err := coll.UpdateOne(
			ctx,
			filter,
			update,
			options.Update().SetUpsert(true),
		)

		if err != nil {
			return successCount, fmt.Errorf("bulk upsert failed at index %d: %w", successCount, err)
		}
		successCount++
	}
	return successCount, nil
}

func (r *artistRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("artist delete by ID failed: %w", err)
	}
	return nil
}

func (r *artistRepository) DeleteByName(ctx context.Context, name string) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		return fmt.Errorf("artist delete by name failed: %w", err)
	}
	return nil
}

func (r *artistRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.ArtistMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"_id": id})

	var artist scene_audio_db_models.ArtistMetadata
	if err := result.Decode(&artist); err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			return nil, nil
		}
		return nil, fmt.Errorf("get artist by ID failed: %w", err)
	}
	return &artist, nil
}

func (r *artistRepository) GetByName(ctx context.Context, name string) (*scene_audio_db_models.ArtistMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"name": name})

	var artist scene_audio_db_models.ArtistMetadata
	if err := result.Decode(&artist); err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			return nil, nil
		}
		return nil, fmt.Errorf("get artist by name failed: %w", err)
	}
	return &artist, nil
}

func (r *artistRepository) GetAllIDs(ctx context.Context) ([]primitive.ObjectID, error) {
	coll := r.db.Collection(r.collection)

	opts := options.Find().SetProjection(bson.M{"_id": 1})

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("查询艺术家ID失败: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	var ids []primitive.ObjectID
	for cursor.Next(ctx) {
		var result struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("解码艺术家ID失败: %w", err)
		}
		ids = append(ids, result.ID)
	}

	return ids, nil
}

func (r *artistRepository) ResetALLField(ctx context.Context) (int64, error) {
	coll := r.db.Collection(r.collection)

	resetFields := []string{
		"album_count",
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

func (r *artistRepository) ResetField(
	ctx context.Context,
	field string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{}

	update := bson.M{"$set": bson.M{field: 0}}

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

func (r *artistRepository) UpdateCounter(
	ctx context.Context,
	artistID primitive.ObjectID,
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

	result, err := coll.UpdateByID(ctx, artistID, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func (r *artistRepository) GetByMBID(ctx context.Context, mbzID string) (*scene_audio_db_models.ArtistMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"mbz_artist_id": mbzID})

	var artist scene_audio_db_models.ArtistMetadata
	if err := result.Decode(&artist); err != nil {
		return nil, fmt.Errorf("通过MBID获取艺术家失败: %w", err)
	}
	return &artist, nil
}
