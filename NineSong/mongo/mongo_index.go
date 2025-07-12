package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateIndexes(db Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Annotation Collection
	annotationCollection := db.Collection(domain.CollectionFileEntityAudioSceneAnnotation)
	createIndex(ctx, annotationCollection, bson.D{{Key: "item_id", Value: 1}, {Key: "item_type", Value: 1}}, "item_id_type")

	// Album Collection - 原始索引 + 拼音优化
	albumCollection := db.Collection(domain.CollectionFileEntityAudioSceneAlbum)
	createIndex(ctx, albumCollection, bson.D{{Key: "artist_id", Value: 1}}, "artist_id")
	createIndex(ctx, albumCollection, bson.D{{Key: "all_artist_ids.artist_id", Value: 1}}, "all_artist_ids_artist_id")
	createIndex(ctx, albumCollection, bson.D{{Key: "name", Value: "text"}, {Key: "artist", Value: "text"}, {Key: "album_artist", Value: "text"}}, "album_text_search")
	createIndex(ctx, albumCollection, bson.D{{Key: "min_year", Value: 1}, {Key: "max_year", Value: 1}}, "year_range")
	createIndex(ctx, albumCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, albumCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, albumCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")
	// 拼音索引
	createIndex(ctx, albumCollection, bson.D{{Key: "name_pinyin", Value: 1}}, "name_pinyin")
	createIndex(ctx, albumCollection, bson.D{{Key: "artist_pinyin", Value: 1}}, "artist_pinyin")
	createIndex(ctx, albumCollection, bson.D{{Key: "album_artist_pinyin", Value: 1}}, "album_artist_pinyin")
	// 复合索引优化
	createIndex(ctx, albumCollection, bson.D{
		{Key: "artist_id", Value: 1},
		{Key: "starred", Value: 1},
	}, "artist_starred_compound")

	// Artist Collection - 原始索引 + 拼音优化
	artistCollection := db.Collection(domain.CollectionFileEntityAudioSceneArtist)
	createIndex(ctx, artistCollection, bson.D{{Key: "name", Value: "text"}}, "artist_text_search")
	createIndex(ctx, artistCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, artistCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, artistCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")
	// 拼音索引
	createIndex(ctx, artistCollection, bson.D{{Key: "name_pinyin", Value: 1}}, "name_pinyin")
	// 复合索引优化
	createIndex(ctx, artistCollection, bson.D{
		{Key: "starred", Value: 1},
		{Key: "play_count", Value: -1},
	}, "starred_playcount_compound")

	// Media File Collection - 原始索引 + 拼音优化
	mediaFileCollection := db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "album_id", Value: 1}}, "album_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "artist_id", Value: 1}}, "artist_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "all_artist_ids.artist_id", Value: 1}}, "all_artist_ids_artist_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "year", Value: 1}}, "year")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "title", Value: "text"}, {Key: "artist", Value: "text"}, {Key: "album", Value: "text"}}, "media_file_text_search")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")
	// 拼音索引
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "title_pinyin", Value: 1}}, "title_pinyin")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "album_pinyin", Value: 1}}, "album_pinyin")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "artist_pinyin", Value: 1}}, "artist_pinyin")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "album_artist_pinyin", Value: 1}}, "album_artist_pinyin")
	// 复合索引优化
	createIndex(ctx, mediaFileCollection, bson.D{
		{Key: "album_id", Value: 1},
		{Key: "year", Value: 1},
	}, "album_year_compound")

	// Media File Cue Collection - 原始索引 + 拼音优化
	mediaFileCueCollection := db.Collection(domain.CollectionFileEntityAudioSceneMediaFileCue)
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "performer_id", Value: 1}}, "performer_id")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "cue_tracks.track_performer_id", Value: 1}}, "cue_tracks_track_performer_id")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "rem.date", Value: 1}}, "rem_date")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "title", Value: "text"}, {Key: "performer", Value: "text"}, {Key: "rem.genre", Value: "text"}, {Key: "cue_tracks.track_title", Value: "text"}}, "media_file_cue_text_search")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")
	// 拼音索引
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "title_pinyin", Value: 1}}, "title_pinyin")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "performer_pinyin", Value: 1}}, "performer_pinyin")
	createIndex(ctx, mediaFileCueCollection, bson.D{
		{Key: "cue_tracks.track_title_pinyin", Value: 1}}, "track_title_pinyin")
	createIndex(ctx, mediaFileCueCollection, bson.D{
		{Key: "cue_tracks.track_performer_pinyin", Value: 1}}, "track_performer_pinyin")
	// 复合索引优化
	createIndex(ctx, mediaFileCueCollection, bson.D{
		{Key: "performer_id", Value: 1},
		{Key: "rem.date", Value: 1},
	}, "performer_date_compound")

	// Playlist Collection
	playlistCollection := db.Collection(domain.CollectionFileEntityAudioScenePlaylist)
	createIndex(ctx, playlistCollection, bson.D{{Key: "name", Value: 1}}, "name")
	// 优化复合索引
	createIndex(ctx, playlistCollection, bson.D{
		{Key: "name", Value: 1},
		{Key: "updated_at", Value: -1}}, "name_updated_compound")

	// Playlist Track Collection
	playlistTrackCollection := db.Collection(domain.CollectionFileEntityAudioScenePlaylistTrack)
	createIndex(ctx, playlistTrackCollection, bson.D{{Key: "playlist_id", Value: 1}, {Key: "media_file_id", Value: 1}}, "playlist_media_file")
	createIndex(ctx, playlistTrackCollection, bson.D{{Key: "index", Value: 1}}, "index")
	// 优化复合索引
	createIndex(ctx, playlistTrackCollection, bson.D{
		{Key: "playlist_id", Value: 1},
		{Key: "index", Value: 1}}, "playlist_index_compound")
}

func createIndex(
	ctx context.Context,
	collection Collection,
	keys bson.D,
	name string,
) {
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetName(name).SetBackground(true),
	}

	if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		fmt.Printf("创建索引 '%s' 失败: %v\n", name, err)
	} else {
		fmt.Printf("索引 '%s' 创建成功\n", name)
	}
}

func DropAllIndexes(db Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		if _, err := collection.Indexes().DropAll(ctx); err != nil {
			fmt.Printf("删除 '%s' 索引失败: %v\n", collName, err)
		} else {
			fmt.Printf("'%s' 索引删除成功\n", collName)
		}
	}
}
