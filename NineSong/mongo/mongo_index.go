package mongo

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func CreateIndexes(db Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Annotation Collection
	annotationCollection := db.Collection(domain.CollectionFileEntityAudioSceneAnnotation)
	createIndex(ctx, annotationCollection, bson.D{{Key: "item_id", Value: 1}, {Key: "item_type", Value: 1}}, "item_id_type")

	// Album Collection
	albumCollection := db.Collection(domain.CollectionFileEntityAudioSceneAlbum)
	createIndex(ctx, albumCollection, bson.D{{Key: "artist_id", Value: 1}}, "artist_id")
	createIndex(ctx, albumCollection, bson.D{{Key: "all_artist_ids.artist_id", Value: 1}}, "all_artist_ids_artist_id")
	createIndex(ctx, albumCollection, bson.D{{Key: "name", Value: "text"}, {Key: "artist", Value: "text"}, {Key: "album_artist", Value: "text"}}, "album_text_search")
	createIndex(ctx, albumCollection, bson.D{{Key: "min_year", Value: 1}, {Key: "max_year", Value: 1}}, "year_range")
	createIndex(ctx, albumCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, albumCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, albumCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")

	// Artist Collection
	artistCollection := db.Collection(domain.CollectionFileEntityAudioSceneArtist)
	createIndex(ctx, artistCollection, bson.D{{Key: "name", Value: "text"}}, "artist_text_search")
	createIndex(ctx, artistCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, artistCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, artistCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")

	// Media File Collection
	mediaFileCollection := db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "album_id", Value: 1}}, "album_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "artist_id", Value: 1}}, "artist_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "all_artist_ids.artist_id", Value: 1}}, "all_artist_ids_artist_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "year", Value: 1}}, "year")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "title", Value: "text"}, {Key: "artist", Value: "text"}, {Key: "album", Value: "text"}}, "media_file_text_search")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")

	// Media File Cue Collection
	mediaFileCueCollection := db.Collection(domain.CollectionFileEntityAudioSceneMediaFileCue)
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "performer_id", Value: 1}}, "performer_id")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "cue_tracks.track_performer_id", Value: 1}}, "cue_tracks_track_performer_id")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "rem.date", Value: 1}}, "rem_date")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "title", Value: "text"}, {Key: "performer", Value: "text"}, {Key: "rem.genre", Value: "text"}, {Key: "cue_tracks.track_title", Value: "text"}}, "media_file_cue_text_search")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")

	// Playlist Collection
	playlistCollection := db.Collection(domain.CollectionFileEntityAudioScenePlaylist)
	createIndex(ctx, playlistCollection, bson.D{{Key: "name", Value: 1}}, "name")

	// Playlist Track Collection
	playlistTrackCollection := db.Collection(domain.CollectionFileEntityAudioScenePlaylistTrack)
	createIndex(ctx, playlistTrackCollection, bson.D{{Key: "playlist_id", Value: 1}, {Key: "media_file_id", Value: 1}}, "playlist_media_file")
	createIndex(ctx, playlistTrackCollection, bson.D{{Key: "index", Value: 1}}, "index")
}

func createIndex(ctx context.Context, collection Collection, keys bson.D, name string) {
	index := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetName(name).SetBackground(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, index)
	if err != nil {
		fmt.Printf("Failed to create index '%s': %v \n", name, err)
	} else {
		fmt.Printf("Index '%s' created successfully. \n", name)
	}
}

func DropAllIndexes(db Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collections := []string{
		domain.CollectionFileEntityAudioSceneAnnotation,
		domain.CollectionFileEntityAudioSceneAlbum,
		domain.CollectionFileEntityAudioSceneArtist,
		domain.CollectionFileEntityAudioSceneMediaFile,
		domain.CollectionFileEntityAudioSceneMediaFileCue,
		domain.CollectionFileEntityAudioScenePlaylist,
		domain.CollectionFileEntityAudioScenePlaylistTrack,
	}

	for _, collName := range collections {
		collection := db.Collection(collName)
		_, err := collection.Indexes().DropAll(ctx)
		if err != nil {
			fmt.Printf("Failed to drop indexes for collection '%s': %v \n", collName, err)
		} else {
			fmt.Printf("Indexes for collection '%s' dropped successfully. \n", collName)
		}
	}
}
