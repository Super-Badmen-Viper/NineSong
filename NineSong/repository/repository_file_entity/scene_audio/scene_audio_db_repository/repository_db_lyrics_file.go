package scene_audio_db_repository

import (
	"context"
	"errors"
	driver "go.mongodb.org/mongo-driver/mongo"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type lyricsFileRepository struct {
	db         mongo.Database
	collection string
}

func NewLyricsFileRepository(db mongo.Database, collection string) scene_audio_db_interface.LyricsFileRepository {
	return &lyricsFileRepository{
		db:         db,
		collection: collection,
	}
}

func (l lyricsFileRepository) GetLyricsFilePath(ctx context.Context, artist, title, fileType string) (string, error) {
	filter := bson.M{
		"artist":   artist,
		"title":    title,
		"fileType": fileType,
	}

	var result struct {
		Path string `bson:"path"`
	}

	err := l.db.Collection(l.collection).FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, driver.ErrNoDocuments) {
			return "", errors.New("lyrics file not found")
		}
		return "", err
	}
	return result.Path, nil
}

func (l lyricsFileRepository) UpdateLyricsFilePath(ctx context.Context, artist, title, fileType, path string) (bool, error) {
	filter := bson.M{
		"artist":   artist,
		"title":    title,
		"fileType": fileType,
	}

	update := bson.M{
		"$set": bson.M{
			"path":     path,
			"artist":   artist,
			"title":    title,
			"fileType": fileType,
		},
	}

	opts := options.Update().SetUpsert(true)

	result, err := l.db.Collection(l.collection).UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return false, err
	}

	return result.ModifiedCount >= 0 || result.UpsertedCount >= 0, nil
}

func (l lyricsFileRepository) CleanAll(ctx context.Context) (bool, error) {
	result, err := l.db.Collection(l.collection).DeleteMany(ctx, bson.M{})
	if err != nil {
		return false, err
	}
	return result >= 0, nil
}
