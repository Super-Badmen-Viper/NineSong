package mongo

import (
	"context"
	"fmt"
	"strings"
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
	createIndex(ctx, annotationCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, annotationCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, annotationCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")
	createIndex(ctx, annotationCollection, bson.D{{Key: "rating", Value: -1}}, "rating")
	createIndex(ctx, annotationCollection, bson.D{{Key: "created_at", Value: -1}}, "created_at")
	createIndex(ctx, annotationCollection, bson.D{{Key: "updated_at", Value: -1}}, "updated_at")
	// 复合索引优化
	createIndex(ctx, annotationCollection, bson.D{
		{Key: "item_id", Value: 1},
		{Key: "item_type", Value: 1},
		{Key: "starred", Value: 1}}, "item_id_type_starred_compound")
	createIndex(ctx, annotationCollection, bson.D{
		{Key: "item_id", Value: 1},
		{Key: "item_type", Value: 1},
		{Key: "play_count", Value: -1}}, "item_id_type_playcount_compound")
	createIndex(ctx, annotationCollection, bson.D{
		{Key: "starred", Value: 1},
		{Key: "created_at", Value: -1}}, "starred_created_compound")

	// Album Collection - 原始索引 + 拼音优化
	albumCollection := db.Collection(domain.CollectionFileEntityAudioSceneAlbum)
	createIndex(ctx, albumCollection, bson.D{{Key: "artist_id", Value: 1}}, "artist_id")
	createIndex(ctx, albumCollection, bson.D{{Key: "all_artist_ids.artist_id", Value: 1}}, "all_artist_ids_artist_id")
	createTextIndex(ctx, albumCollection, bson.D{{Key: "name", Value: "text"}, {Key: "artist", Value: "text"}, {Key: "album_artist", Value: "text"}}, "album_text_search")
	createIndex(ctx, albumCollection, bson.D{{Key: "min_year", Value: 1}, {Key: "max_year", Value: 1}}, "year_range")
	createIndex(ctx, albumCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, albumCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, albumCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")
	createIndex(ctx, albumCollection, bson.D{{Key: "song_count", Value: 1}}, "song_count")
	createIndex(ctx, albumCollection, bson.D{{Key: "created_at", Value: -1}}, "created_at")
	// 拼音索引
	createIndex(ctx, albumCollection, bson.D{{Key: "name_pinyin", Value: 1}}, "name_pinyin")
	createIndex(ctx, albumCollection, bson.D{{Key: "artist_pinyin", Value: 1}}, "artist_pinyin")
	createIndex(ctx, albumCollection, bson.D{{Key: "album_artist_pinyin", Value: 1}}, "album_artist_pinyin")
	// 复合索引优化
	createIndex(ctx, albumCollection, bson.D{
		{Key: "artist_id", Value: 1},
		{Key: "starred", Value: 1},
	}, "artist_starred_compound")
	createIndex(ctx, albumCollection, bson.D{
		{Key: "min_year", Value: 1},
		{Key: "max_year", Value: 1},
		{Key: "play_count", Value: -1},
	}, "year_playcount_compound")
	createIndex(ctx, albumCollection, bson.D{
		{Key: "starred", Value: 1},
		{Key: "created_at", Value: -1},
	}, "starred_created_compound")

	// Artist Collection - 原始索引 + 拼音优化
	artistCollection := db.Collection(domain.CollectionFileEntityAudioSceneArtist)
	createTextIndex(ctx, artistCollection, bson.D{{Key: "name", Value: "text"}}, "artist_text_search")
	createIndex(ctx, artistCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, artistCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, artistCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")
	createIndex(ctx, artistCollection, bson.D{{Key: "album_count", Value: 1}}, "album_count")
	createIndex(ctx, artistCollection, bson.D{{Key: "song_count", Value: 1}}, "song_count")
	createIndex(ctx, artistCollection, bson.D{{Key: "created_at", Value: -1}}, "created_at")
	// 拼音索引
	createIndex(ctx, artistCollection, bson.D{{Key: "name_pinyin", Value: 1}}, "name_pinyin")
	// 复合索引优化
	createIndex(ctx, artistCollection, bson.D{
		{Key: "starred", Value: 1},
		{Key: "play_count", Value: -1},
	}, "starred_playcount_compound")
	createIndex(ctx, artistCollection, bson.D{
		{Key: "starred", Value: 1},
		{Key: "created_at", Value: -1},
	}, "starred_created_compound")
	createIndex(ctx, artistCollection, bson.D{
		{Key: "play_count", Value: -1},
		{Key: "created_at", Value: -1},
	}, "playcount_created_compound")

	// Media File Collection - 原始索引 + 拼音优化
	mediaFileCollection := db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "album_id", Value: 1}}, "album_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "artist_id", Value: 1}}, "artist_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "all_artist_ids.artist_id", Value: 1}}, "all_artist_ids_artist_id")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "year", Value: 1}}, "year")
	// 为媒体文件创建文本索引
	createTextIndex(ctx, mediaFileCollection, bson.D{
		{Key: "title", Value: "text"},
		{Key: "artist", Value: "text"},
		{Key: "album", Value: "text"},
		{Key: "genre", Value: "text"},
		{Key: "lyrics", Value: "text"},
	}, "media_file_text_search")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "starred", Value: 1}}, "starred")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "play_count", Value: -1}}, "play_count")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "play_date", Value: -1}}, "play_date")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "genre", Value: 1}}, "genre")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "suffix", Value: 1}}, "suffix")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "bit_rate", Value: 1}}, "bit_rate")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "library_path", Value: 1}}, "library_path")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "path", Value: 1}}, "path")
	createIndex(ctx, mediaFileCollection, bson.D{{Key: "created_at", Value: -1}}, "created_at")
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
	createIndex(ctx, mediaFileCollection, bson.D{
		{Key: "artist_id", Value: 1},
		{Key: "starred", Value: 1},
	}, "artist_starred_compound")
	createIndex(ctx, mediaFileCollection, bson.D{
		{Key: "genre", Value: 1},
		{Key: "play_count", Value: -1},
	}, "genre_play_count_compound")
	createIndex(ctx, mediaFileCollection, bson.D{
		{Key: "library_path", Value: 1},
		{Key: "suffix", Value: 1},
	}, "library_path_suffix_compound")
	createIndex(ctx, mediaFileCollection, bson.D{
		{Key: "year", Value: -1},
		{Key: "play_date", Value: -1},
	}, "year_play_date_compound")

	// Media File Cue Collection - 原始索引 + 拼音优化
	mediaFileCueCollection := db.Collection(domain.CollectionFileEntityAudioSceneMediaFileCue)
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "performer_id", Value: 1}}, "performer_id")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "cue_tracks.track_performer_id", Value: 1}}, "cue_tracks_track_performer_id")
	createIndex(ctx, mediaFileCueCollection, bson.D{{Key: "rem.date", Value: 1}}, "rem_date")
	createTextIndex(ctx, mediaFileCueCollection, bson.D{{Key: "title", Value: "text"}, {Key: "performer", Value: "text"}, {Key: "rem.genre", Value: "text"}, {Key: "cue_tracks.track_title", Value: "text"}}, "media_file_cue_text_search")
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
	createIndex(ctx, playlistCollection, bson.D{{Key: "created_at", Value: -1}}, "created_at")
	// 优化复合索引
	createIndex(ctx, playlistCollection, bson.D{
		{Key: "name", Value: 1},
		{Key: "updated_at", Value: -1}}, "name_updated_compound")
	createIndex(ctx, playlistCollection, bson.D{
		{Key: "created_at", Value: -1},
		{Key: "updated_at", Value: -1}}, "created_updated_compound")

	// Playlist Track Collection
	playlistTrackCollection := db.Collection(domain.CollectionFileEntityAudioScenePlaylistTrack)
	createIndex(ctx, playlistTrackCollection, bson.D{{Key: "playlist_id", Value: 1}, {Key: "media_file_id", Value: 1}}, "playlist_media_file")
	createIndex(ctx, playlistTrackCollection, bson.D{{Key: "index", Value: 1}}, "index")
	// 修复索引命名冲突，只保留一个复合索引
	createIndex(ctx, playlistTrackCollection, bson.D{
		{Key: "playlist_id", Value: 1},
		{Key: "index", Value: 1}}, "playlist_id_index_compound")
	createIndex(ctx, playlistTrackCollection, bson.D{
		{Key: "playlist_id", Value: 1},
		{Key: "media_file_id", Value: 1},
		{Key: "index", Value: 1}}, "playlist_media_index_compound")
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

// 新增函数：创建文本索引，避免重复创建
func createTextIndex(
	ctx context.Context,
	collection Collection,
	keys bson.D,
	name string,
) {
	// 先检查是否已存在同名索引
	specs, err := collection.Indexes().ListSpecifications(ctx)
	if err != nil {
		fmt.Printf("检查索引失败: %v\n", err)
		// 如果检查失败，仍然尝试创建索引
		indexModel := mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetName(name).SetBackground(true),
		}

		if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
			fmt.Printf("创建索引 '%s' 失败: %v\n", name, err)
		} else {
			fmt.Printf("索引 '%s' 创建成功\n", name)
		}
		return
	}

	// 遍历现有索引
	for _, spec := range specs {
		// 如果已存在同名索引，跳过创建
		if spec.Name == name {
			fmt.Printf("索引 '%s' 已存在，跳过创建\n", name)
			return
		}

		// 检查是否已存在文本索引（MongoDB每个集合只能有一个文本索引）
		isExistingTextIndex := false
		var specKeys bson.D
		if err := bson.Unmarshal(spec.KeysDocument, &specKeys); err == nil {
			for _, key := range specKeys {
				if key.Value == "text" {
					isExistingTextIndex = true
					break
				}
			}
		}

		// 检查是否要创建的也是文本索引
		isNewTextIndex := false
		for _, key := range keys {
			if key.Value == "text" {
				isNewTextIndex = true
				break
			}
		}

		// 如果已存在文本索引且要创建的也是文本索引，则跳过
		if isExistingTextIndex && isNewTextIndex {
			fmt.Printf("集合已存在文本索引 '%s'，跳过创建新的文本索引 '%s'\n", spec.Name, name)
			return
		}
	}

	// 创建索引
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetName(name).SetBackground(true),
	}

	if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		// 如果是因为已存在文本索引导致的错误，给出提示信息
		if strings.Contains(err.Error(), "language override unsupported") {
			fmt.Printf("集合已存在文本索引，无法创建新的文本索引 '%s'。请检查数据库中是否已存在其他文本索引。\n", name)
		} else {
			fmt.Printf("创建索引 '%s' 失败: %v\n", name, err)
		}
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

// 新增函数：删除指定集合的文本索引
func DropTextIndexes(db Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collections := []struct {
		name           string
		collectionName string
	}{
		{"媒体文件集合", domain.CollectionFileEntityAudioSceneMediaFile},
		{"专辑集合", domain.CollectionFileEntityAudioSceneAlbum},
		{"艺术家集合", domain.CollectionFileEntityAudioSceneArtist},
		{"媒体文件CUE集合", domain.CollectionFileEntityAudioSceneMediaFileCue},
	}

	for _, coll := range collections {
		collection := db.Collection(coll.collectionName)
		specs, err := collection.Indexes().ListSpecifications(ctx)
		if err != nil {
			fmt.Printf("获取%s索引列表失败: %v\n", coll.name, err)
			continue
		}

		// 检查是否需要删除索引
		needDrop := false
		for _, spec := range specs {
			// 检查是否为文本索引
			var specKeys bson.D
			if err := bson.Unmarshal(spec.KeysDocument, &specKeys); err == nil {
				for _, key := range specKeys {
					if key.Value == "text" {
						needDrop = true
						break
					}
				}
			}
			if needDrop {
				break
			}
		}

		// 如果有文本索引，删除所有索引
		if needDrop {
			fmt.Printf("%s中发现文本索引，删除所有索引\n", coll.name)

			// 删除所有索引
			if _, err := collection.Indexes().DropAll(ctx); err != nil {
				fmt.Printf("删除%s所有索引失败: %v\n", coll.name, err)
			} else {
				fmt.Printf("已删除%s的所有索引\n", coll.name)
			}
		}
	}
}
