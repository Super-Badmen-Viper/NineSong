package scene_audio_db_repository

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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

func (r *artistRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	coll := r.db.Collection(r.collection)

	// 构建原子更新操作
	result, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
		options.Update().SetUpsert(false),
	)

	if err != nil {
		return false, fmt.Errorf("艺术家更新失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return false, nil
	}

	return true, nil
}

func (r *artistRepository) UpdateMany(
	ctx context.Context,
	filter interface{},
	update interface{},
	opts ...*options.UpdateOptions,
) (*driver.UpdateResult, error) {
	coll := r.db.Collection(r.collection)
	return coll.UpdateMany(ctx, filter, update, opts...)
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

func (r *artistRepository) DeleteAll(ctx context.Context) (int64, error) {
	coll := r.db.Collection(r.collection)

	// 删除集合中的所有文档
	result, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("删除所有艺术家失败: %w", err)
	}

	// 记录操作日志
	if result > 0 {
		log.Printf("已删除全部 %d 个艺术家记录", result)
	} else {
		log.Printf("集合中无艺术家记录可删除")
	}

	return result, nil
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

func (r *artistRepository) GetAllCounts(ctx context.Context) ([]scene_audio_db_models.ArtistAlbumAndSongCounts, error) {
	coll := r.db.Collection(r.collection)

	// 使用 find() + 投影直接返回字段值
	filter := bson.M{} // 空过滤条件，查询所有文档
	projection := bson.M{
		"_id":               1, // 保留 ID
		"updated_at":        1,
		"album_count":       1, // 返回专辑数
		"guest_album_count": 1, // 返回合作专辑数
		"song_count":        1, // 返回单曲数
		"guest_song_count":  1, // 返回合作单曲数
		"cue_count":         1, // 返回光盘数
		"guest_cue_count":   1, // 返回合作光盘数
	}

	cursor, err := coll.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []scene_audio_db_models.ArtistAlbumAndSongCounts
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解码失败: %w", err)
	}

	return results, nil
}

func (r *artistRepository) ResetALLField(ctx context.Context) (int64, error) {
	coll := r.db.Collection(r.collection)

	resetFields := []string{
		"album_count",
		"song_count",
		"size",
		"duration",
		"guest_album_count",
		"guest_song_count",
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

func (r *artistRepository) GetByMbzID(ctx context.Context, mbzID string) (*scene_audio_db_models.ArtistMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"mbz_artist_id": mbzID})

	var artist scene_audio_db_models.ArtistMetadata
	if err := result.Decode(&artist); err != nil {
		return nil, fmt.Errorf("通过MBID获取艺术家失败: %w", err)
	}
	return &artist, nil
}

func (r *artistRepository) DeleteAllInvalid(ctx context.Context) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{
		"$and": []bson.M{
			{"album_count": bson.M{"$eq": 0, "$exists": true}},
			{"guest_album_count": bson.M{"$eq": 0, "$exists": true}},
			{"song_count": bson.M{"$eq": 0, "$exists": true}},
			{"guest_song_count": bson.M{"$eq": 0, "$exists": true}},
			{"cue_count": bson.M{"$eq": 0, "$exists": true}},
			{"guest_cue_count": bson.M{"$eq": 0, "$exists": true}},
		},
	}

	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除无效艺术家失败: %w", err)
	}

	if result > 0 {
		log.Printf("已删除 %d 个无效艺术家", result)
	}
	return result, nil
}

func (r *artistRepository) inspectCounter(
	ctx context.Context,
	artistID string,
	field string,
	operand int,
) (int, error) {
	objID, err := primitive.ObjectIDFromHex(artistID)
	if err != nil {
		return 0, fmt.Errorf("无效的艺术家ID格式: %w", err)
	}

	artist, err := r.GetByID(ctx, objID)
	if err != nil {
		return 0, fmt.Errorf("获取艺术家信息失败: %w", err)
	}
	if artist == nil {
		return 0, fmt.Errorf("艺术家不存在")
	}

	deleteArtist := false
	if field == "song_count" {
		deleteArtist = artist.SongCount-operand <= 0
	} else if field == "guest_song_count" {
		deleteArtist = artist.GuestSongCount-operand <= 0
	} else if field == "album_count" {
		deleteArtist = artist.AlbumCount-operand <= 0
	} else if field == "guest_album_count" {
		deleteArtist = artist.GuestAlbumCount-operand <= 0
	} else if field == "cue_count" {
		deleteArtist = artist.CueCount-operand <= 0
	} else if field == "guest_cue_count" {
		deleteArtist = artist.GuestCueCount-operand <= 0
	} else {
		return 0, fmt.Errorf("未知的计数字段: %s", field)
	}

	if deleteArtist {
		_, err := r.UpdateCounter(ctx, objID, field, 0)
		if err != nil {
			return 0, err
		}
		return 0, nil
	} else {
		modifiedCount, err := r.UpdateCounter(ctx, objID, field, -operand)
		if err != nil {
			return 0, fmt.Errorf("更新%s失败: %w", field, err)
		}
		return int(modifiedCount), nil
	}
}

func (r *artistRepository) InspectAlbumCountByArtist(
	ctx context.Context,
	artistID string,
	operand int,
) (int, error) {
	return r.inspectCounter(ctx, artistID, "album_count", operand)
}

func (r *artistRepository) InspectGuestAlbumCountByArtist(
	ctx context.Context,
	artistID string,
	operand int,
) (int, error) {
	return r.inspectCounter(ctx, artistID, "guest_album_count", operand)
}

func (r *artistRepository) InspectMediaCountByArtist(
	ctx context.Context,
	artistID string,
	operand int,
) (int, error) {
	return r.inspectCounter(ctx, artistID, "song_count", operand)
}

func (r *artistRepository) InspectGuestMediaCountByArtist(
	ctx context.Context,
	artistID string,
	operand int,
) (int, error) {
	return r.inspectCounter(ctx, artistID, "guest_song_count", operand)
}

func (r *artistRepository) InspectMediaCueCountByArtist(
	ctx context.Context,
	artistID string,
	operand int,
) (int, error) {
	return r.inspectCounter(ctx, artistID, "cue_count", operand)
}

func (r *artistRepository) InspectGuestMediaCueCountByArtist(
	ctx context.Context,
	artistID string,
	operand int,
) (int, error) {
	return r.inspectCounter(ctx, artistID, "guest_cue_count", operand)
}
