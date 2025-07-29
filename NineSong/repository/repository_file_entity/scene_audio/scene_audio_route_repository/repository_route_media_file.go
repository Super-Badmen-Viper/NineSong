package scene_audio_route_repository

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/siongui/gojianfan"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type mediaFileRepository struct {
	db         mongo.Database
	collection string
}

func NewMediaFileRepository(db mongo.Database, collection string) scene_audio_route_interface.MediaFileRepository {
	return &mediaFileRepository{
		db:         db,
		collection: collection,
	}
}

func (r *mediaFileRepository) GetMediaFileItems(
	ctx context.Context,
	start, end, sort, order, search, starred, albumId, artistId, year,
	suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter string,
) ([]scene_audio_route_models.MediaFileMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	coll := r.db.Collection(r.collection)

	// 构建聚合管道（完全使用bson.D结构）
	pipeline := []bson.D{
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "mediaId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$mediaId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "media"}}},
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

	// 添加基础过滤条件
	if match := buildMatchStage(search, starred, albumId, artistId, year,
		suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter); len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// 处理play_date排序的特殊过滤
	validatedSort := validateSortField(sort, albumId)
	if validatedSort == "play_date" {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "play_count", Value: bson.D{{Key: "$gt", Value: 0}}},
			}},
		})
	}

	// 添加排序阶段 - 关键修改：添加唯一字段作为次要排序条件
	pipeline = append(pipeline, buildSortStage(validatedSort, order))

	// 添加分页阶段
	paginationStages := buildMediaPaginationStage(start, end)
	if paginationStages != nil {
		pipeline = append(pipeline, paginationStages...)
	}

	// 执行聚合查询
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			fmt.Printf("cursor close error: %v\n", cerr)
		}
	}()

	var results []scene_audio_route_models.MediaFileMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

func (r *mediaFileRepository) GetMediaFileMetadataItems(
	ctx context.Context,
	start, end, sort, order, search, starred, albumId, artistId, year,
	suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter string,
) ([]scene_audio_db_models.MediaFileMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	coll := r.db.Collection(r.collection)

	// 构建聚合管道（完全使用bson.D结构）
	pipeline := []bson.D{
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "mediaId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$mediaId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "media"}}},
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

	// 添加基础过滤条件
	if match := buildMatchStage(search, starred, albumId, artistId, year,
		suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter); len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// 处理play_date排序的特殊过滤
	validatedSort := validateSortField(sort, albumId)
	if validatedSort == "play_date" {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "play_count", Value: bson.D{{Key: "$gt", Value: 0}}},
			}},
		})
	}

	// 添加排序阶段 - 关键修改：添加唯一字段作为次要排序条件
	pipeline = append(pipeline, buildSortStage(validatedSort, order))

	// 添加分页阶段
	paginationStages := buildMediaPaginationStage(start, end)
	if paginationStages != nil {
		pipeline = append(pipeline, paginationStages...)
	}

	// 执行聚合查询
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			fmt.Printf("cursor close error: %v\n", cerr)
		}
	}()

	var results []scene_audio_db_models.MediaFileMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

func (r *mediaFileRepository) GetMediaFileItemsIds(
	ctx context.Context, ids []string,
) ([]scene_audio_route_models.MediaFileMetadata, error) {
	if len(ids) == 0 {
		return []scene_audio_route_models.MediaFileMetadata{}, nil
	}

	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, idStr := range ids {
		if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
			objectIDs = append(objectIDs, oid)
		}
	}

	coll := r.db.Collection(r.collection)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ID查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []scene_audio_route_models.MediaFileMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("结果解析失败: %w", err)
	}

	return results, nil
}

func (r *mediaFileRepository) GetMediaFileItemsMultipleSorting(
	ctx context.Context,
	start, end string,
	sortOrder []domain_util.SortOrder,
	search, starred, albumId, artistId, year,
	suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter string,
) ([]scene_audio_route_models.MediaFileMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	coll := r.db.Collection(r.collection)

	// 构建聚合管道
	pipeline := []bson.D{
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "mediaId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$mediaId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "media"}}},
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

	// 添加基础过滤条件
	if match := buildMatchStage(search, starred, albumId, artistId, year,
		suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter); len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// 检查是否需要play_date过滤
	if hasPlayDateSort(sortOrder) {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "play_count", Value: bson.D{{Key: "$gt", Value: 0}}},
			}},
		})
	}

	// 添加多重排序阶段
	if sortStage := buildMultiSortStage(sortOrder); sortStage != nil {
		pipeline = append(pipeline, *sortStage)
	}

	// 添加分页阶段
	if paginationStages := buildMediaPaginationStage(start, end); paginationStages != nil {
		pipeline = append(pipeline, paginationStages...)
	}

	// 执行聚合查询
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []scene_audio_route_models.MediaFileMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

// 新增：构建多重排序阶段
func buildMultiSortStage(sortOrder []domain_util.SortOrder) *bson.D {
	if len(sortOrder) == 0 {
		return nil
	}

	sortCriteria := bson.D{}
	for _, so := range sortOrder {
		mappedField := mapSortField(so.Sort)
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

// 新增：检查排序条件是否包含play_date
func hasPlayDateSort(sortOrder []domain_util.SortOrder) bool {
	for _, so := range sortOrder {
		if mapSortField(so.Sort) == "play_date" {
			return true
		}
	}
	return false
}

// 新增：排序字段映射（无albumId逻辑）
func mapSortField(sort string) string {
	sortMappings := map[string]string{
		"title":          "order_title",
		"album":          "order_album_name",
		"artist":         "order_artist_name",
		"album_artist":   "order_album_artist_name",
		"year":           "year",
		"rating":         "rating",
		"starred_at":     "starred_at",
		"genre":          "genre",
		"play_count":     "play_count",
		"play_date":      "play_date",
		"duration":       "duration",
		"bit_rate":       "bit_rate",
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

func (r *mediaFileRepository) GetMediaFileFilterItemsCount(
	ctx context.Context,
) (*scene_audio_route_models.MediaFileFilterCounts, error) {
	coll := r.db.Collection(r.collection)

	pipeline := []bson.D{
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneAnnotation},
				{Key: "let", Value: bson.D{{Key: "mediaId", Value: "$_id"}}},
				{Key: "pipeline", Value: []bson.D{
					{
						{Key: "$match", Value: bson.D{
							{Key: "$expr", Value: bson.D{
								{Key: "$and", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$item_id", "$$mediaId"}}},
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "media"}}},
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
	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			fmt.Printf("cursor close error: %v\n", cerr)
		}
	}()

	var result []struct {
		Total      []map[string]int `bson:"total"`
		Starred    []map[string]int `bson:"starred"`
		RecentPlay []map[string]int `bson:"recent_play"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("decode count error: %w", err)
	}

	counts := &scene_audio_route_models.MediaFileFilterCounts{}
	if len(result) > 0 {
		counts.Total = extractCount(result[0].Total)
		counts.Starred = extractCount(result[0].Starred)
		counts.RecentPlay = extractCount(result[0].RecentPlay)
	}
	return counts, nil
}

// 排序字段映射
func validateSortField(sort, albumId string) string {
	sortMappings := map[string]string{
		"title":          "order_title",
		"album":          "order_album_name",
		"artist":         "order_artist_name",
		"album_artist":   "order_album_artist_name",
		"year":           "year",
		"rating":         "rating",
		"starred_at":     "starred_at",
		"genre":          "genre",
		"play_count":     "play_count",
		"play_date":      "play_date",
		"duration":       "duration",
		"bit_rate":       "bit_rate",
		"size":           "size",
		"created_at":     "created_at",
		"updated_at":     "updated_at",
		"recently_added": "created_at",
	}

	if mapped, ok := sortMappings[strings.ToLower(sort)]; ok {
		return mapped
	}

	if len(albumId) > 0 {
		return "file_name"
	} else {
		return "_id"
	}
}

// 排序稳定性：添加唯一字段作为次要排序条件
func buildSortStage(sort, order string) bson.D {
	sortOrder := 1
	if order == "desc" {
		sortOrder = -1
	}
	return bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: sort, Value: sortOrder},
			{Key: "_id", Value: 1}, // 关键修复：添加唯一字段保证排序稳定性[5,10](@ref)
		}},
	}
}

// 增强分页参数验证
func buildMediaPaginationStage(start, end string) []bson.D {
	startInt, err1 := strconv.Atoi(start)
	endInt, err2 := strconv.Atoi(end)

	// 参数验证
	if err1 != nil || err2 != nil || startInt < 0 || endInt <= startInt {
		return nil // 无效参数不添加分页阶段
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

// Helper functions
func extractCount(data []map[string]int) int {
	if len(data) > 0 {
		return data[0]["count"]
	}
	return 0
}

func buildMatchStage(search, starred, albumId, artistId, year, suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter string) bson.D {
	filter := bson.D{}

	if artistId != "" {
		artistFilter := bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "artist_id", Value: artistId}},
				bson.D{{Key: "all_artist_ids.artist_id", Value: artistId}},
			}},
		}
		filter = append(filter, bson.E{Key: "$and", Value: bson.A{artistFilter}})
	}
	if albumId != "" {
		filter = append(filter, bson.E{Key: "album_id", Value: albumId})
	}
	if year != "" {
		if yearInt, err := strconv.Atoi(year); err == nil {
			filter = append(filter, bson.E{Key: "year", Value: yearInt})
		}
	}
	if search != "" {
		// 生成简繁体四重匹配模式[1,6](@ref)
		pattern := buildFTSRegexPattern(search)

		searchFilter := bson.A{
			// 原始名称字段的简繁体模糊搜索[6,8](@ref)
			bson.D{{Key: "title", Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: "artist", Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: "album", Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: "lyrics", Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: "i"}}}},
			// 拼音字段的简繁体模糊搜索
			bson.D{{Key: "title_pinyin_full", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: "album_pinyin_full", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: "artist_pinyin_full", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: "album_artist_pinyin_full", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
		}
		filter = append(filter, bson.E{Key: "$or", Value: searchFilter})
	}
	if starred != "" {
		if isStarred, err := strconv.ParseBool(starred); err == nil {
			filter = append(filter, bson.E{Key: "starred", Value: isStarred})
		}
	}

	// 修复后缀过滤条件
	if suffix != "" {
		filter = append(filter, bson.E{Key: "suffix", Value: suffix})
	}

	// 文件夹路径过滤条件
	if folderPath != "" {
		// 预编译正则表达式 [6,7](@ref)
		pathRegex := regexp.MustCompile("^" + regexp.QuoteMeta(folderPath))
		filter = append(filter, bson.E{
			Key:   "library_path",
			Value: bson.D{{Key: "$regex", Value: pathRegex.String()}, {Key: "$options", Value: "i"}},
		})
		// 子路径模糊匹配优化
		if folderPathSubFilter != "" {
			// 使用非贪婪匹配减少回溯 [5,7](@ref)
			subRegex := regexp.MustCompile(".*?" + regexp.QuoteMeta(folderPathSubFilter))
			filter = append(filter, bson.E{
				Key:   "path",
				Value: bson.D{{Key: "$regex", Value: subRegex.String()}, {Key: "$options", Value: "i"}},
			})
		}
	}

	// 修复比特率范围过滤条件
	minVal, maxVal := 0, 0
	bitrateCondition := bson.D{}
	var err error
	if minBitrate != "" {
		if minVal, err = strconv.Atoi(minBitrate); err == nil && minVal > 0 {
			bitrateCondition = append(bitrateCondition, bson.E{Key: "$gte", Value: minVal})
		}
	}
	if maxBitrate != "" {
		if maxVal, err = strconv.Atoi(maxBitrate); err == nil && maxVal > 0 {
			// 允许单独设置最大值，不依赖最小值
			if minVal <= 0 || minVal < maxVal {
				bitrateCondition = append(bitrateCondition, bson.E{Key: "$lte", Value: maxVal})
			}
		}
	}
	if len(bitrateCondition) > 0 {
		filter = append(filter, bson.E{Key: "bit_rate", Value: bitrateCondition})
	}

	return filter
}

// 生成简繁体四重匹配的正则模式
func buildFTSRegexPattern(input string) string {
	// 生成四种编码形式（覆盖所有简繁体组合）
	t2s := gojianfan.T2S(input)         // 繁体转简体
	s2t := gojianfan.S2T(input)         // 简体转繁体
	doubleConvert := gojianfan.S2T(t2s) // 双重转换

	// 对每个模式进行正则转义并组合
	patterns := []string{
		regexp.QuoteMeta(input),
		regexp.QuoteMeta(t2s),
		regexp.QuoteMeta(s2t),
		regexp.QuoteMeta(doubleConvert),
	}
	return strings.Join(patterns, "|")
}
