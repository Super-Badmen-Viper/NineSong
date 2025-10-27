package scene_audio_route_repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type artistRepository struct {
	db         mongo.Database
	collection string
}

func NewArtistRepository(db mongo.Database, collection string) scene_audio_route_interface.ArtistRepository {
	return &artistRepository{
		db:         db,
		collection: collection,
	}
}

func (r *artistRepository) GetArtistItems(
	ctx context.Context,
	start, end, sort, order, search, starred string,
) ([]scene_audio_route_models.ArtistMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	coll := r.db.Collection(r.collection)

	pipeline := []bson.D{
		// 使用$lookup但不立即$unwind
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "artistId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$artistId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "artist"}}},
								}},
							}},
						}},
					},
					// 新增：只取第一个注解（避免重复）
					{{Key: "$limit", Value: 1}},
				}},
				{Key: "as", Value: "annotations"},
			}},
		},
		// 修改$unwind阶段
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$annotations"},
				{Key: "preserveNullAndEmptyArrays", Value: true}, // 保留无注解的艺术家
			}},
		},
		{
			{Key: "$addFields", Value: bson.D{
				{Key: "play_count", Value: "$annotations.play_count"},
				{Key: "play_complete_count", Value: "$annotations.play_complete_count"},
				{Key: "play_date", Value: "$annotations.play_date"},
				{Key: "rating", Value: "$annotations.rating"},
				{Key: "starred", Value: "$annotations.starred"},
				{Key: "starred_at", Value: "$annotations.starred_at"},
			}},
		},
	}

	// 添加过滤条件
	if match := buildArtistMatch(search, starred); len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// 处理特殊排序
	validatedSort := validateArtistSortField(sort)
	if validatedSort == "play_date" {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "play_count", Value: bson.D{{Key: "$gt", Value: 0}}},
			}},
		})
	}

	// 修复排序稳定性
	pipeline = append(pipeline, buildArtistSortStage(validatedSort, order))

	// 增强分页参数验证
	paginationStages := buildArtistPaginationStage(start, end)
	if paginationStages != nil {
		pipeline = append(pipeline, paginationStages...)
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer func(cursor mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			fmt.Printf("error closing cursor: %v\n", err)
		}
	}(cursor, ctx)

	var results []scene_audio_route_models.ArtistMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

func (r *artistRepository) GetArtistMetadataItems(
	ctx context.Context,
	start, end, sort, order, search, starred string,
) ([]scene_audio_db_models.ArtistMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	coll := r.db.Collection(r.collection)

	pipeline := []bson.D{
		// 使用$lookup但不立即$unwind
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "artistId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$artistId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "artist"}}},
								}},
							}},
						}},
					},
					// 新增：只取第一个注解（避免重复）
					{{Key: "$limit", Value: 1}},
				}},
				{Key: "as", Value: "annotations"},
			}},
		},
		// 修改$unwind阶段
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$annotations"},
				{Key: "preserveNullAndEmptyArrays", Value: true}, // 保留无注解的艺术家
			}},
		},
		{
			{Key: "$addFields", Value: bson.D{
				{Key: "play_count", Value: "$annotations.play_count"},
				{Key: "play_complete_count", Value: "$annotations.play_complete_count"},
				{Key: "play_date", Value: "$annotations.play_date"},
				{Key: "rating", Value: "$annotations.rating"},
				{Key: "starred", Value: "$annotations.starred"},
				{Key: "starred_at", Value: "$annotations.starred_at"},
			}},
		},
	}

	// 添加过滤条件
	if match := buildArtistMatch(search, starred); len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// 处理特殊排序
	validatedSort := validateArtistSortField(sort)
	if validatedSort == "play_date" {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "play_count", Value: bson.D{{Key: "$gt", Value: 0}}},
			}},
		})
	}

	// 修复排序稳定性
	pipeline = append(pipeline, buildArtistSortStage(validatedSort, order))

	// 增强分页参数验证
	paginationStages := buildArtistPaginationStage(start, end)
	if paginationStages != nil {
		pipeline = append(pipeline, paginationStages...)
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer func(cursor mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			fmt.Printf("error closing cursor: %v\n", err)
		}
	}(cursor, ctx)

	var results []scene_audio_db_models.ArtistMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

func (r *artistRepository) GetArtistItemsMultipleSorting(
	ctx context.Context,
	start, end string,
	sortOrder []domain_util.SortOrder,
	search, starred string,
) ([]scene_audio_route_models.ArtistMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	coll := r.db.Collection(r.collection)

	// 构建聚合管道
	pipeline := []bson.D{
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "artistId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$artistId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "artist"}}},
								}},
							}},
						}},
					},
					{{Key: "$limit", Value: 1}}, // 只取第一个注解
				}},
				{Key: "as", Value: "annotations"},
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$annotations"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
		{
			{Key: "$addFields", Value: bson.D{
				{Key: "play_count", Value: "$annotations.play_count"},
				{Key: "play_complete_count", Value: "$annotations.play_complete_count"},
				{Key: "play_date", Value: "$annotations.play_date"},
				{Key: "rating", Value: "$annotations.rating"},
				{Key: "starred", Value: "$annotations.starred"},
				{Key: "starred_at", Value: "$annotations.starred_at"},
			}},
		},
	}

	// 添加基础过滤条件
	if match := buildArtistMatch(search, starred); len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// 检查是否需要播放相关过滤
	if hasPlaySortArtist(sortOrder) {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "play_count", Value: bson.D{{Key: "$gt", Value: 0}}},
			}},
		})
	}

	// 添加多重排序阶段
	if sortStage := buildArtistMultiSortStage(sortOrder); sortStage != nil {
		pipeline = append(pipeline, *sortStage)
	}

	// 添加分页阶段
	if paginationStages := buildArtistPaginationStage(start, end); paginationStages != nil {
		pipeline = append(pipeline, paginationStages...)
	}

	// 执行聚合查询
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []scene_audio_route_models.ArtistMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

func (r *artistRepository) GetArtistTreeItems(ctx context.Context, start, end, artistId string) ([]scene_audio_route_models.ArtistTreeMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	coll := r.db.Collection(r.collection)

	// 构建查询条件
	matchStage := bson.D{}
	if artistId != "" {
		// 如果提供了artistId，则只查询该艺术家
		objectId, err := primitive.ObjectIDFromHex(artistId)
		if err != nil {
			return nil, fmt.Errorf("invalid artistId format: %w", err)
		}
		matchStage = bson.D{{Key: "_id", Value: objectId}}
	}

	// 首先获取艺术家列表
	artistPipeline := []bson.D{
		// 添加匹配条件
		{{Key: "$match", Value: matchStage}},
		// 添加注解数据
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "artistId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$artistId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "artist"}}},
								}},
							}},
						}},
					},
				}},
				{Key: "as", Value: "annotations"},
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$annotations"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
		{
			{Key: "$addFields", Value: bson.D{
				{Key: "play_count", Value: "$annotations.play_count"},
				{Key: "play_complete_count", Value: "$annotations.play_complete_count"},
				{Key: "play_date", Value: "$annotations.play_date"},
				{Key: "rating", Value: "$annotations.rating"},
				{Key: "starred", Value: "$annotations.starred"},
				{Key: "starred_at", Value: "$annotations.starred_at"},
			}},
		},
	}

	// 添加分页处理
	paginationStages := buildArtistPaginationStage(start, end)
	if paginationStages != nil {
		artistPipeline = append(artistPipeline, paginationStages...)
	}

	// 执行艺术家查询
	artistCursor, err := coll.Aggregate(ctx, artistPipeline)
	if err != nil {
		return nil, fmt.Errorf("artist query failed: %w", err)
	}
	defer artistCursor.Close(ctx)

	var artists []scene_audio_route_models.ArtistMetadata
	if err := artistCursor.All(ctx, &artists); err != nil {
		return nil, fmt.Errorf("decode artist error: %w", err)
	}

	// 为每个艺术家获取专辑数据（按年份倒序排列）
	result := make([]scene_audio_route_models.ArtistTreeMetadata, len(artists))
	albumRepo := NewAlbumRepository(r.db, domain.CollectionFileEntityAudioSceneAlbum)
	mediaFileRepo := NewMediaFileRepository(r.db, domain.CollectionFileEntityAudioSceneMediaFile)

	for i, artist := range artists {
		result[i].Artist = artist

		// 获取艺术家的专辑，按年份倒序排列
		albums, err := albumRepo.GetAlbumItems(ctx, "0", "1000", "max_year", "desc", "", "", artist.ID.Hex(), "", "")
		if err != nil {
			return nil, fmt.Errorf("album query failed for artist %s: %w", artist.ID.Hex(), err)
		}

		// 为每个专辑获取媒体文件
		albumTree := make([]scene_audio_route_models.AlbumTreeMetadata, len(albums))
		for j, album := range albums {
			albumTree[j].Album = album

			// 获取专辑的所有媒体文件，按索引升序排列
			mediaFiles, err := mediaFileRepo.GetMediaFileItems(ctx, "0", "1000", "index", "asc", "", "", album.ID.Hex(), "", "", "", "", "", "", "")
			if err != nil {
				return nil, fmt.Errorf("media file query failed for album %s: %w", album.ID.Hex(), err)
			}

			albumTree[j].MediaFiles = mediaFiles
		}

		result[i].Albums = albumTree
	}

	return result, nil
}

// 新增：构建艺术家多重排序阶段
func buildArtistMultiSortStage(sortOrder []domain_util.SortOrder) *bson.D {
	if len(sortOrder) == 0 {
		return nil
	}

	sortCriteria := bson.D{}
	for _, so := range sortOrder {
		mappedField := mapArtistSortField(so.Sort)
		orderVal := 1
		if strings.ToLower(so.Order) == "desc" {
			orderVal = -1
		}
		sortCriteria = append(sortCriteria, bson.E{Key: mappedField, Value: orderVal})
	}
	// 添加稳定性排序字段
	sortCriteria = append(sortCriteria, bson.E{Key: "_id", Value: 1})

	return &bson.D{{Key: "$sort", Value: sortCriteria}}
}

// 新增：艺术家排序字段映射
func mapArtistSortField(sort string) string {
	sortMappings := map[string]string{
		"name":           "order_artist_name",
		"album_count":    "album_count",
		"song_count":     "song_count",
		"play_count":     "play_count",
		"play_date":      "play_date",
		"rating":         "rating",
		"starred_at":     "starred_at",
		"size":           "size",
		"created_at":     "created_at",
		"updated_at":     "updated_at",
		"recently_added": "created_at",
	}
	if mapped, ok := sortMappings[strings.ToLower(sort)]; ok {
		return mapped
	}
	return sort
}

// 新增：检查排序条件是否包含播放相关字段
func hasPlaySortArtist(sortOrder []domain_util.SortOrder) bool {
	for _, so := range sortOrder {
		if mapArtistSortField(so.Sort) == "play_count" || mapArtistSortField(so.Sort) == "play_date" {
			return true
		}
	}
	return false
}

func (r *artistRepository) GetArtistFilterItemsCount(
	ctx context.Context,
) (*scene_audio_route_models.ArtistFilterCounts, error) {
	coll := r.db.Collection(r.collection)

	pipeline := []bson.D{
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "artistId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$artistId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "artist"}}},
								}},
							}},
						}},
					},
				}},
				{Key: "as", Value: "annotations"},
			}},
		},
		{
			{Key: "$facet", Value: bson.D{
				{Key: "total", Value: []bson.D{
					{{Key: "$count", Value: "count"}},
				}},
				{Key: "starred", Value: []bson.D{
					{{Key: "$match", Value: bson.D{
						{Key: "annotations.starred", Value: true},
					}}},
					{{Key: "$count", Value: "count"}},
				}},
				{Key: "recent_play", Value: []bson.D{
					{{Key: "$match", Value: bson.D{
						{Key: "annotations.play_count", Value: bson.D{
							{Key: "$gt", Value: 0},
						}},
					}}},
					{{Key: "$count", Value: "count"}},
				}},
			}},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("count query failed: %w", err)
	}
	defer func(cursor mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			fmt.Printf("error closing cursor: %v\n", err)
		}
	}(cursor, ctx)

	var result []struct {
		Total      []map[string]int `bson:"total"`
		Starred    []map[string]int `bson:"starred"`
		RecentPlay []map[string]int `bson:"recent_play"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("decode count error: %w", err)
	}

	counts := &scene_audio_route_models.ArtistFilterCounts{}
	if len(result) > 0 {
		counts.Total = extractCount(result[0].Total)
		counts.Starred = extractCount(result[0].Starred)
		counts.RecentPlay = extractCount(result[0].RecentPlay)
	}

	return counts, nil
}

// Helper functions
func buildArtistMatch(search, starred string) bson.D {
	filter := bson.D{}

	if search != "" {
		// 生成简繁体四重匹配模式[1,6](@ref)
		pattern := buildFTSRegexPattern(search)

		filter = append(filter, bson.E{
			Key: "$or",
			Value: bson.A{
				// 原始名称字段的模糊搜索（正则匹配）
				bson.D{{Key: "name", Value: bson.D{
					{Key: "$regex", Value: pattern},
					{Key: "$options", Value: "i"},
				}}},
				// 拼音字段的模糊搜索（正则匹配）
				bson.D{{Key: "name_pinyin_full", Value: bson.D{
					{Key: "$regex", Value: search},
					{Key: "$options", Value: "i"},
				}}},
			},
		})
	}

	if starred != "" {
		if isStarred, err := strconv.ParseBool(starred); err == nil {
			filter = append(filter, bson.E{Key: "starred", Value: isStarred})
		}
	}

	return filter
}

func validateArtistSortField(sort string) string {
	sortMappings := map[string]string{
		"name":           "order_artist_name",
		"album_count":    "album_count",
		"song_count":     "song_count",
		"play_count":     "play_count",
		"play_date":      "play_date",
		"rating":         "rating",
		"starred_at":     "starred_at",
		"size":           "size",
		"created_at":     "created_at",
		"updated_at":     "updated_at",
		"recently_added": "created_at",
	}

	lowerSort := strings.ToLower(sort)
	if mappedField, exists := sortMappings[lowerSort]; exists {
		return mappedField
	}

	validSortFields := map[string]bool{
		"order_artist_name": true,
		"album_count":       true,
		"song_count":        true,
		"play_count":        true,
		"play_date":         true,
		"rating":            true,
		"starred_at":        true,
		"size":              true,
		"created_at":        true,
		"updated_at":        true,
	}

	if validSortFields[lowerSort] {
		return lowerSort
	}

	return "_id"
}

// 修复排序稳定性
func buildArtistSortStage(sort, order string) bson.D {
	sortOrder := 1
	if order == "desc" {
		sortOrder = -1
	}
	return bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: sort, Value: sortOrder},
			{Key: "_id", Value: 1}, // 关键修复
		}},
	}
}

// 增强分页验证
func buildArtistPaginationStage(start, end string) []bson.D {
	startInt, err1 := strconv.Atoi(start)
	endInt, err2 := strconv.Atoi(end)

	if err1 != nil || err2 != nil || startInt < 0 || endInt <= startInt {
		return nil // 无效参数不添加分页
	}

	skip := startInt
	limit := endInt - startInt

	var stages []bson.D
	if skip > 0 {
		stages = append(stages, bson.D{{Key: "$skip", Value: skip}})
	}
	if limit > 0 {
		stages = append(stages, bson.D{{Key: "$limit", Value: limit}})
	}
	return stages
}
