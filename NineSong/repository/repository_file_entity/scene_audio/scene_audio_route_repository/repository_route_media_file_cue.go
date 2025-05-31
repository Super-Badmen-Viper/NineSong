package scene_audio_route_repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type mediaFileCueRepository struct {
	db         mongo.Database
	collection string
}

func NewMediaFileCueRepository(db mongo.Database, collection string) scene_audio_route_interface.MediaFileCueRepository {
	return &mediaFileCueRepository{
		db:         db,
		collection: collection,
	}
}

func (r *mediaFileCueRepository) GetMediaFileCueItems(
	ctx context.Context,
	start, end, sort, order, search, starred, albumId, artistId, year string,
) ([]scene_audio_route_models.MediaFileCueMetadata, error) {
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
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "media_cue"}}}, // 修改为 media_cue 类型
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
				{Key: "play_date", Value: "$annotations.play_date"},
				{Key: "rating", Value: "$annotations.rating"},
				{Key: "starred", Value: "$annotations.starred"},
				{Key: "starred_at", Value: "$annotations.starred_at"},
			}},
		},
	}

	// 添加过滤条件
	if match := r.buildMatchStage(search, starred, albumId, artistId, year); len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// 处理排序字段
	validatedSort := r.validateSortField(sort)
	if validatedSort == "play_date" {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "play_count", Value: bson.D{{Key: "$gt", Value: 0}}},
			}},
		})
	}

	// 添加排序
	pipeline = append(pipeline, r.buildSortStage(validatedSort, order))

	// 添加分页
	pipeline = append(pipeline, r.buildPaginationStage(start, end)...)

	// 执行查询
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []scene_audio_route_models.MediaFileCueMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return results, nil
}

func (r *mediaFileCueRepository) GetMediaFileCueFilterItemsCount(
	ctx context.Context,
	search, starred, albumId, artistId, year string,
) (*scene_audio_route_models.MediaFileCueFilterCounts, error) {
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
									bson.D{{Key: "$eq", Value: bson.A{"$item_type", "media_cue"}}}, // 修改为 media_cue 类型
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
				{Key: "starred", Value: "$annotations.starred"},
			}},
		},
		{
			{Key: "$match", Value: r.buildBaseMatch(search, albumId, artistId, year)},
		},
		{
			{Key: "$facet", Value: bson.D{
				{Key: "total", Value: []bson.D{
					{{Key: "$count", Value: "count"}},
				}},
				{Key: "starred", Value: []bson.D{
					{{Key: "$match", Value: bson.D{
						{Key: "starred", Value: true},
					}}},
					{{Key: "$count", Value: "count"}},
				}},
				{Key: "recent_play", Value: []bson.D{
					{{Key: "$match", Value: bson.D{
						{Key: "play_count", Value: bson.D{
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
	defer cursor.Close(ctx)

	var result []struct {
		Total      []map[string]int `bson:"total"`
		Starred    []map[string]int `bson:"starred"`
		RecentPlay []map[string]int `bson:"recent_play"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("decode count error: %w", err)
	}

	counts := &scene_audio_route_models.MediaFileCueFilterCounts{}
	if len(result) > 0 {
		counts.Total = r.extractCount(result[0].Total)
		counts.Starred = r.extractCount(result[0].Starred)
		counts.RecentPlay = r.extractCount(result[0].RecentPlay)
	}
	return counts, nil
}

// 辅助函数
func (r *mediaFileCueRepository) extractCount(data []map[string]int) int {
	if len(data) > 0 {
		return data[0]["count"]
	}
	return 0
}

func (r *mediaFileCueRepository) validateSortField(sort string) string {
	sortMappings := map[string]string{
		"title":           "title",
		"performer":       "performer",
		"year":            "rem.date", // 使用REM中的日期字段
		"rating":          "rating",
		"starred_at":      "starred_at",
		"genre":           "rem.genre",
		"play_count":      "play_count",
		"play_date":       "play_date",
		"size":            "size",
		"created_at":      "created_at",
		"updated_at":      "updated_at",
		"cue_track_count": "cue_track_count",
	}

	if mapped, ok := sortMappings[strings.ToLower(sort)]; ok {
		return mapped
	}
	return "_id"
}

func (r *mediaFileCueRepository) buildMatchStage(search, starred, albumId, artistId, year string) bson.D {
	filter := bson.D{}

	// 艺术家过滤（CUE中可能包含多个艺术家）
	if artistId != "" {
		artistFilter := bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "performer_id", Value: artistId}}, // 假设有艺术家ID字段
				bson.D{{Key: "cue_tracks.track_performer_id", Value: artistId}},
			}},
		}
		filter = append(filter, bson.E{Key: "$and", Value: bson.A{artistFilter}})
	}

	// 年份过滤（使用REM中的日期）
	if year != "" {
		filter = append(filter, bson.E{
			Key: "$expr",
			Value: bson.D{
				{Key: "$eq", Value: bson.A{
					bson.D{{Key: "$year", Value: bson.D{{Key: "$toDate", Value: "$rem.date"}}}},
					year,
				}},
			},
		})
	}

	// 全文搜索
	if search != "" {
		filter = append(filter, bson.E{Key: "$or", Value: []bson.D{
			{{Key: "performer", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
			{{Key: "title", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
			{{Key: "rem.genre", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
			{{Key: "cue_tracks.track_title", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
			{{Key: "full_text", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
		}})
	}

	// 星标过滤
	if starred != "" {
		if isStarred, err := strconv.ParseBool(starred); err == nil {
			filter = append(filter, bson.E{Key: "starred", Value: isStarred})
		}
	}

	return filter
}

func (r *mediaFileCueRepository) buildBaseMatch(search, albumId, artistId, year string) bson.D {
	return r.buildMatchStage(search, "", albumId, artistId, year)
}

func (r *mediaFileCueRepository) buildSortStage(sort, order string) bson.D {
	sortOrder := 1
	if order == "desc" {
		sortOrder = -1
	}
	return bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: sort, Value: sortOrder},
		}},
	}
}

func (r *mediaFileCueRepository) buildPaginationStage(start, end string) []bson.D {
	var stages []bson.D

	startInt, _ := strconv.Atoi(start)
	endInt, _ := strconv.Atoi(end)

	skip := startInt
	limit := endInt - startInt

	if skip > 0 {
		stages = append(stages, bson.D{{Key: "$skip", Value: skip}})
	}
	if limit > 0 {
		stages = append(stages, bson.D{{Key: "$limit", Value: limit}})
	}

	return stages
}
