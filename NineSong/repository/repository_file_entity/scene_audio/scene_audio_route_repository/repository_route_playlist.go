package scene_audio_route_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type playlistRepository struct {
	db         mongo.Database
	collection string
}

func NewPlaylistRepository(db mongo.Database, collection string) scene_audio_route_interface.PlaylistRepository {
	return &playlistRepository{
		db:         db,
		collection: collection,
	}
}

// 获取所有播放列表
func (p *playlistRepository) GetPlaylistsAll(ctx context.Context) ([]scene_audio_route_models.PlaylistMetadata, error) {
	coll := p.db.Collection(p.collection)
	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{"created_at", -1}}))
	if err != nil {
		return nil, fmt.Errorf("find operation failed: %w", err)
	}
	defer cursor.Close(ctx)

	var dbModels []scene_audio_db_models.PlaylistMetadata
	if err := cursor.All(ctx, &dbModels); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	routeModels := convertToRouteModels(dbModels)

	// 批量计算所有播放列表的统计数据
	statsMap, err := p.calculatePlaylistsStatsBatch(ctx, routeModels)
	if err != nil {
		return nil, fmt.Errorf("calculate stats failed: %w", err)
	}

	// 更新每个播放列表的统计数据
	for i := range routeModels {
		if stats, ok := statsMap[routeModels[i].ID.Hex()]; ok {
			routeModels[i].SongCount = stats.SongCount
			routeModels[i].Duration = stats.Duration
		} else {
			// 如果没有统计数据，设置为0
			routeModels[i].SongCount = 0
			routeModels[i].Duration = 0
		}
	}

	return routeModels, nil
}

// 获取单个播放列表
func (p *playlistRepository) GetPlaylist(ctx context.Context, playlistId string) (*scene_audio_route_models.PlaylistMetadata, error) {
	objID, err := primitive.ObjectIDFromHex(playlistId)
	if err != nil {
		return nil, errors.New("invalid playlist id format")
	}

	coll := p.db.Collection(p.collection)
	var dbModel scene_audio_db_models.PlaylistMetadata
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&dbModel)
	if err != nil {
		return nil, fmt.Errorf("find one error: %w", err)
	}

	routeModel := convertToRouteModel(dbModel)

	// 计算播放列表的统计数据
	stats, err := p.calculatePlaylistStats(ctx, playlistId)
	if err != nil {
		return nil, fmt.Errorf("calculate stats failed: %w", err)
	}

	routeModel.SongCount = stats.SongCount
	routeModel.Duration = stats.Duration

	return routeModel, nil
}

// 创建播放列表
func (p *playlistRepository) CreatePlaylist(ctx context.Context, playlist scene_audio_route_models.PlaylistMetadata) (*scene_audio_route_models.PlaylistMetadata, error) {
	// 构造新的唯一性校验条件
	filter := bson.D{
		{"name", playlist.Name}, // 仅保留name字段校验[3,4](@ref)
	}

	// 查询重复项
	count, err := p.db.Collection(p.collection).CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	if count > 0 {
		return nil, errors.New("playlist name already exists")
	}

	if playlist.ID.IsZero() {
		playlist.ID = primitive.NewObjectID()
	}

	now := time.Now().UTC()
	playlist.CreatedAt = now
	playlist.UpdatedAt = now

	dbModel := convertToDBModel(playlist)

	coll := p.db.Collection(p.collection)
	insertResult, err := coll.InsertOne(ctx, dbModel)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}

	// 获取生成的ObjectID
	if oid, ok := insertResult.(primitive.ObjectID); ok {
		playlist.ID = oid
	} else {
		return nil, errors.New("invalid objectid generated")
	}

	return &playlist, nil
}

// 删除播放列表
func (p *playlistRepository) DeletePlaylist(ctx context.Context, playlistId string) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(playlistId)
	if err != nil {
		return false, errors.New("invalid playlist id format")
	}

	coll := p.db.Collection(p.collection)
	_, err = coll.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return false, fmt.Errorf("delete failed: %w", err)
	}

	return true, nil
}

// 更新播放列表基本信息
func (p *playlistRepository) UpdatePlaylistInfo(ctx context.Context, playlistId string, playlist scene_audio_route_models.PlaylistMetadata) (*scene_audio_route_models.PlaylistMetadata, error) {
	objID, err := primitive.ObjectIDFromHex(playlistId)
	if err != nil {
		return nil, errors.New("invalid playlist id format")
	}

	// 添加名称唯一性检查
	filter := bson.M{
		"name": playlist.Name,
		"_id":  bson.M{"$ne": objID},
	}
	count, err := p.db.Collection(p.collection).CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("name check failed: %w", err)
	}
	if count > 0 {
		return nil, errors.New("playlist name already exists")
	}

	update := bson.M{
		"$set": bson.M{
			"name":       playlist.Name,
			"comment":    playlist.Comment,
			"updated_at": time.Now().UTC(),
		},
	}

	coll := p.db.Collection(p.collection)
	// 执行更新操作
	result, err := coll.UpdateByID(ctx, objID, update)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}
	if result.MatchedCount == 0 {
		return nil, errors.New("document not found")
	}

	// 新增：查询更新后的文档
	var updatedDoc scene_audio_db_models.PlaylistMetadata
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedDoc)
	if err != nil {
		return nil, fmt.Errorf("fetch updated document failed: %w", err)
	}

	return convertToRouteModel(updatedDoc), nil
}

// 数据库模型转换
func convertToDBModel(routeModel scene_audio_route_models.PlaylistMetadata) scene_audio_db_models.PlaylistMetadata {
	return scene_audio_db_models.PlaylistMetadata{
		ID:        routeModel.ID,
		Name:      routeModel.Name,
		Comment:   routeModel.Comment,
		CreatedAt: routeModel.CreatedAt,
		UpdatedAt: routeModel.UpdatedAt,
	}
}

// 路由模型转换
func convertToRouteModel(dbModel scene_audio_db_models.PlaylistMetadata) *scene_audio_route_models.PlaylistMetadata {
	return &scene_audio_route_models.PlaylistMetadata{
		ID:        dbModel.ID,
		Name:      dbModel.Name,
		Comment:   dbModel.Comment,
		Duration:  dbModel.Duration,
		SongCount: dbModel.SongCount,
		CreatedAt: dbModel.CreatedAt,
		UpdatedAt: dbModel.UpdatedAt,
		Path:      dbModel.Path,
		Size:      dbModel.Size,
	}
}

// 批量转换路由模型
func convertToRouteModels(dbModels []scene_audio_db_models.PlaylistMetadata) []scene_audio_route_models.PlaylistMetadata {
	routeModels := make([]scene_audio_route_models.PlaylistMetadata, 0, len(dbModels))
	for _, m := range dbModels {
		routeModels = append(routeModels, *convertToRouteModel(m))
	}
	return routeModels
}

// PlaylistStats 播放列表统计数据
type PlaylistStats struct {
	SongCount float64
	Duration  float64
}

// calculatePlaylistStats 计算单个播放列表的统计数据
func (p *playlistRepository) calculatePlaylistStats(ctx context.Context, playlistId string) (*PlaylistStats, error) {
	objID, err := primitive.ObjectIDFromHex(playlistId)
	if err != nil {
		return nil, errors.New("invalid playlist id format")
	}

	playlistTrackColl := p.db.Collection(domain.CollectionFileEntityAudioScenePlaylistTrack)

	// 构建聚合管道
	pipeline := []bson.D{
		// 匹配指定播放列表的关联记录
		{
			{Key: "$match", Value: bson.D{
				{Key: "playlist_id", Value: objID},
			}},
		},
		// 关联媒体文件表
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneMediaFile},
				{Key: "localField", Value: "media_file_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "media_file"},
			}},
		},
		// 展开媒体文件数组
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$media_file"},
				{Key: "preserveNullAndEmptyArrays", Value: false},
			}},
		},
		// 分组统计
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$playlist_id"},
				{Key: "song_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "duration", Value: bson.D{{Key: "$sum", Value: "$media_file.duration"}}},
			}},
		},
	}

	cursor, err := playlistTrackColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate failed: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		ID        primitive.ObjectID `bson:"_id"`
		SongCount int                `bson:"song_count"`
		Duration  float64            `bson:"duration"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}
		return &PlaylistStats{
			SongCount: float64(result.SongCount),
			Duration:  result.Duration,
		}, nil
	}

	// 如果没有找到记录，返回0
	return &PlaylistStats{
		SongCount: 0,
		Duration:  0,
	}, nil
}

// calculatePlaylistsStatsBatch 批量计算所有播放列表的统计数据（优化版本，避免N+1查询）
func (p *playlistRepository) calculatePlaylistsStatsBatch(ctx context.Context, playlists []scene_audio_route_models.PlaylistMetadata) (map[string]*PlaylistStats, error) {
	if len(playlists) == 0 {
		return make(map[string]*PlaylistStats), nil
	}

	playlistTrackColl := p.db.Collection(domain.CollectionFileEntityAudioScenePlaylistTrack)

	// 收集所有播放列表ID
	playlistIDs := make([]primitive.ObjectID, 0, len(playlists))
	for _, playlist := range playlists {
		playlistIDs = append(playlistIDs, playlist.ID)
	}

	// 构建聚合管道
	pipeline := []bson.D{
		// 匹配所有相关播放列表的关联记录
		{
			{Key: "$match", Value: bson.D{
				{Key: "playlist_id", Value: bson.D{
					{Key: "$in", Value: playlistIDs},
				}},
			}},
		},
		// 关联媒体文件表
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneMediaFile},
				{Key: "localField", Value: "media_file_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "media_file"},
			}},
		},
		// 展开媒体文件数组
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$media_file"},
				{Key: "preserveNullAndEmptyArrays", Value: false},
			}},
		},
		// 按播放列表ID分组统计
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$playlist_id"},
				{Key: "song_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "duration", Value: bson.D{{Key: "$sum", Value: "$media_file.duration"}}},
			}},
		},
	}

	cursor, err := playlistTrackColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate failed: %w", err)
	}
	defer cursor.Close(ctx)

	// 构建结果映射
	statsMap := make(map[string]*PlaylistStats)

	var result struct {
		ID        primitive.ObjectID `bson:"_id"`
		SongCount int                `bson:"song_count"`
		Duration  float64            `bson:"duration"`
	}

	for cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}
		statsMap[result.ID.Hex()] = &PlaylistStats{
			SongCount: float64(result.SongCount),
			Duration:  result.Duration,
		}
	}

	return statsMap, nil
}

// GetFirstTrackCoverImage 获取播放列表第一首歌的封面图片路径
func (p *playlistRepository) GetFirstTrackCoverImage(ctx context.Context, playlistId string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(playlistId)
	if err != nil {
		return "", errors.New("invalid playlist id format")
	}

	playlistTrackColl := p.db.Collection(domain.CollectionFileEntityAudioScenePlaylistTrack)

	// 构建聚合管道，获取第一首歌（按index排序）
	pipeline := []bson.D{
		// 匹配指定播放列表
		{
			{Key: "$match", Value: bson.D{
				{Key: "playlist_id", Value: objID},
			}},
		},
		// 关联媒体文件表
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneMediaFile},
				{Key: "localField", Value: "media_file_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "media_file"},
			}},
		},
		// 展开媒体文件数组
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$media_file"},
				{Key: "preserveNullAndEmptyArrays", Value: false},
			}},
		},
		// 按index排序
		{
			{Key: "$sort", Value: bson.D{
				{Key: "index", Value: 1},
			}},
		},
		// 只取第一个
		{
			{Key: "$limit", Value: 1},
		},
		// 只返回封面URL字段
		{
			{Key: "$project", Value: bson.D{
				{Key: "high_image_url", Value: "$media_file.high_image_url"},
				{Key: "medium_image_url", Value: "$media_file.medium_image_url"},
			}},
		},
	}

	cursor, err := playlistTrackColl.Aggregate(ctx, pipeline)
	if err != nil {
		return "", fmt.Errorf("aggregate failed: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		HighImageURL   string `bson:"high_image_url"`
		MediumImageURL string `bson:"medium_image_url"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return "", fmt.Errorf("decode failed: %w", err)
		}
		// 优先返回高清封面，如果没有则返回中等分辨率封面
		if result.HighImageURL != "" {
			return result.HighImageURL, nil
		}
		return result.MediumImageURL, nil
	}

	// 如果没有找到，返回空字符串
	return "", nil
}

// GetTopThreeTrackCoverImages 获取播放列表前三个媒体项的封面图片路径
func (p *playlistRepository) GetTopThreeTrackCoverImages(ctx context.Context, playlistId string) ([]string, error) {
	objID, err := primitive.ObjectIDFromHex(playlistId)
	if err != nil {
		return nil, errors.New("invalid playlist id format")
	}

	playlistTrackColl := p.db.Collection(domain.CollectionFileEntityAudioScenePlaylistTrack)

	// 构建聚合管道，获取前三个媒体项（按index排序）
	pipeline := []bson.D{
		// 匹配指定播放列表
		{
			{Key: "$match", Value: bson.D{
				{Key: "playlist_id", Value: objID},
			}},
		},
		// 关联媒体文件表
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: domain.CollectionFileEntityAudioSceneMediaFile},
				{Key: "localField", Value: "media_file_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "media_file"},
			}},
		},
		// 展开媒体文件数组
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$media_file"},
				{Key: "preserveNullAndEmptyArrays", Value: false},
			}},
		},
		// 按index排序
		{
			{Key: "$sort", Value: bson.D{
				{Key: "index", Value: 1},
			}},
		},
		// 只取前三个
		{
			{Key: "$limit", Value: 3},
		},
		// 只返回封面URL字段
		{
			{Key: "$project", Value: bson.D{
				{Key: "high_image_url", Value: "$media_file.high_image_url"},
				{Key: "medium_image_url", Value: "$media_file.medium_image_url"},
			}},
		},
	}

	cursor, err := playlistTrackColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate failed: %w", err)
	}
	defer cursor.Close(ctx)

	var covers []string
	var result struct {
		HighImageURL   string `bson:"high_image_url"`
		MediumImageURL string `bson:"medium_image_url"`
	}

	for cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}
		// 优先使用高清封面，如果没有则使用中等分辨率封面
		coverPath := result.HighImageURL
		if coverPath == "" {
			coverPath = result.MediumImageURL
		}
		// 只添加非空的封面路径
		if coverPath != "" {
			covers = append(covers, coverPath)
		}
	}

	return covers, nil
}
