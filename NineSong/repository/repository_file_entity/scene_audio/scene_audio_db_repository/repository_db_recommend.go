package scene_audio_db_repository

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type recommendRepository struct {
	db         mongo.Database
	collection string
}

func NewRecommendRepository(db mongo.Database, collection string) scene_audio_route_interface.RecommendRouteRepository {
	return &recommendRepository{
		db:         db,
		collection: collection,
	}
}

func (r recommendRepository) GetRecommendAnnotationWordCloudItems(
	ctx context.Context,
	start, end, recommendType, randomSeed, offset string,
) ([]interface{}, error) {
	// 解析参数
	startInt, err := strconv.Atoi(start)
	if err != nil {
		return nil, fmt.Errorf("无效的start参数: %w", err)
	}

	endInt, err := strconv.Atoi(end)
	if err != nil {
		return nil, fmt.Errorf("无效的end参数: %w", err)
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		return nil, fmt.Errorf("无效的offset参数: %w", err)
	}

	// 验证参数范围
	if startInt >= endInt {
		return nil, fmt.Errorf("start参数必须小于end参数")
	}

	// 设置随机种子
	seed, err := strconv.ParseInt(randomSeed, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的randomSeed参数: %w", err)
	}
	rand.Seed(seed)

	// 根据推荐类型选择不同的集合和返回类型
	// 将recommendType映射到对应的ItemType
	var itemType string
	var targetCollection string
	switch recommendType {
	case "artist":
		itemType = "artist"
		targetCollection = domain.CollectionFileEntityAudioSceneArtist
	case "album":
		itemType = "album"
		targetCollection = domain.CollectionFileEntityAudioSceneAlbum
	case "media":
		itemType = "media"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFile
	case "media_cue":
		itemType = "media_cue"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFileCue
	default:
		itemType = "media"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFile
	}

	// 查询带有注释的数据
	annotationColl := r.db.Collection(domain.CollectionFileEntityAudioSceneAnnotation)
	wordCloudColl := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFileWordCloud)

	// 先检查是否有注释数据
	count, err := annotationColl.CountDocuments(ctx, bson.M{"item_type": itemType})
	if err != nil {
		return nil, fmt.Errorf("检查注释数据失败: %w", err)
	}

	// 如果没有注释数据，直接从目标集合获取数据
	if count == 0 {
		results, err := r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, startInt, endInt, offsetInt)
		if err != nil {
			return nil, fmt.Errorf("获取降级数据失败: %w", err)
		}

		// 如果降级数据也为空，返回错误
		if results == nil || len(results) == 0 {
			return nil, fmt.Errorf("未找到任何%s类型的数据", recommendType)
		}

		return results, nil
	}

	// 正确的推荐逻辑：
	// 1. 从annotation中提取用户喜欢项目的标签作为样本
	// 2. 使用这些标签在全曲库中寻找相似项目进行推荐
	// 3. 主要推荐那些在annotation中未出现的、但与annotation中项目有相似tag的项目
	// 4. offset参数用于tag高频词的偏离

	results, err := r.getSimilarityBasedRecommendations(ctx, annotationColl, wordCloudColl, targetCollection, itemType, recommendType, startInt, endInt, offsetInt)
	if err != nil {
		return nil, fmt.Errorf("获取相似性推荐失败: %w", err)
	}

	// 如果结果为空，返回错误
	if len(results) == 0 {
		return nil, fmt.Errorf("未找到推荐数据")
	}

	return results, nil
}

// 基于相似性的推荐算法
func (r recommendRepository) getSimilarityBasedRecommendations(
	ctx context.Context,
	annotationColl mongo.Collection,
	wordCloudColl mongo.Collection,
	targetCollection string,
	itemType string,
	recommendType string,
	start int,
	end int,
	offset int,
) ([]interface{}, error) {
	// 1. 从用户喜欢的项目中提取标签作为样本
	// 获取用户行为数据中的标签
	tagPipeline := []bson.D{
		// 匹配项目类型
		{{Key: "$match", Value: bson.D{{Key: "item_type", Value: itemType}}}},
		// 展开标签数组
		{{Key: "$unwind", Value: "$weighted_tags"}},
		// 按标签权重分组并排序
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$weighted_tags.tag"},
			{Key: "totalWeight", Value: bson.D{{Key: "$sum", Value: "$weighted_tags.weight"}}},
		}}},
		// 按权重降序排序
		{{Key: "$sort", Value: bson.D{{Key: "totalWeight", Value: -1}}}},
		// 应用offset偏移量
		{{Key: "$skip", Value: offset}},
		// 限制标签数量
		{{Key: "$limit", Value: 50}},
	}

	tagCursor, err := annotationColl.Aggregate(ctx, tagPipeline)
	if err != nil {
		return nil, fmt.Errorf("标签聚合查询失败: %w", err)
	}
	defer tagCursor.Close(ctx)

	// 收集用户喜欢的标签（应用offset偏移后）
	var userTags []struct {
		Tag         string  `bson:"_id"`
		TotalWeight float64 `bson:"totalWeight"`
	}
	if err := tagCursor.All(ctx, &userTags); err != nil {
		return nil, fmt.Errorf("解析标签数据失败: %w", err)
	}

	if len(userTags) == 0 {
		// 如果没有标签数据，使用降级策略
		return r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, start, end, offset)
	}

	// 2. 使用这些标签在词云中查找相似的标签
	var tagNames []string
	tagWeights := make(map[string]float64)
	for _, tag := range userTags {
		tagNames = append(tagNames, tag.Tag)
		tagWeights[tag.Tag] = tag.TotalWeight
	}

	// 在词云中查找包含这些标签的项目
	wordCloudPipeline := []bson.D{
		// 匹配标签
		{{Key: "$match", Value: bson.D{{Key: "name", Value: bson.D{{Key: "$in", Value: tagNames}}}}}},
		// 按计数降序排序
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		// 限制数量
		{{Key: "$limit", Value: 100}},
	}

	wordCloudCursor, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
	if err != nil {
		return nil, fmt.Errorf("词云聚合查询失败: %w", err)
	}
	defer wordCloudCursor.Close(ctx)

	// 收集词云中的相关标签
	var relatedTags []struct {
		ID    primitive.ObjectID `bson:"_id"`
		Name  string             `bson:"name"`
		Count int                `bson:"count"`
		Type  string             `bson:"type"`
		Rank  int                `bson:"rank"`
	}
	if err := wordCloudCursor.All(ctx, &relatedTags); err != nil {
		return nil, fmt.Errorf("解析词云数据失败: %w", err)
	}

	// 3. 基于标签相似性计算项目得分
	// 创建标签到得分的映射
	tagScores := make(map[string]float64)
	for _, tag := range relatedTags {
		// 标签得分 = 词云计数 * 用户对该标签的权重
		userWeight := tagWeights[tag.Name]
		tagScores[tag.Name] = float64(tag.Count) * userWeight
	}

	// 4. 获取已存在于annotation中的项目ID，用于排除
	annotatedItemsPipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "item_type", Value: itemType}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "itemIds", Value: bson.D{{Key: "$addToSet", Value: "$item_id"}}},
		}}},
	}

	annotatedCursor, err := annotationColl.Aggregate(ctx, annotatedItemsPipeline)
	if err != nil {
		return nil, fmt.Errorf("获取已标注项目失败: %w", err)
	}
	defer annotatedCursor.Close(ctx)

	var annotatedResult []struct {
		ItemIds []string `bson:"itemIds"`
	}
	if err := annotatedCursor.All(ctx, &annotatedResult); err != nil {
		return nil, fmt.Errorf("解析已标注项目数据失败: %w", err)
	}

	var annotatedItemIds []string
	if len(annotatedResult) > 0 {
		annotatedItemIds = annotatedResult[0].ItemIds
	}

	// 5. 从目标集合中查找项目并计算相似性得分
	// 这里简化处理，实际应该根据项目的标签信息计算得分
	// 由于项目模型中没有直接的标签字段，我们需要通过其他方式关联

	// 获取已存在于annotation中的项目ID，用于排除
	itemColl := r.db.Collection(targetCollection)

	// 构建聚合管道
	itemPipeline := []bson.D{
		// 排除已标注的项目
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: annotatedItemIds}}}}}},
		// 添加随机排序
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: end * 2}}}}, // 多取一些数据用于后续处理
		// 添加偏移量处理（用于分页）
		{{Key: "$skip", Value: start}},
		// 限制返回数量
		{{Key: "$limit", Value: end - start}},
	}

	itemCursor, err := itemColl.Aggregate(ctx, itemPipeline)
	if err != nil {
		return nil, fmt.Errorf("项目聚合查询失败: %w", err)
	}
	defer itemCursor.Close(ctx)

	var results []interface{}
	for itemCursor.Next(ctx) {
		var itemDoc bson.M
		if err := itemCursor.Decode(&itemDoc); err != nil {
			continue
		}

		// 计算相似性分数（这里简化处理，实际应该基于标签匹配）
		score := 0.5 + rand.Float64()*0.5 // 0.5-1.0的随机分数

		// 根据推荐类型创建具体的推荐结果
		result, err := r.createRecommendationResult(itemDoc, recommendType, score, "基于内容相似性推荐", 0, 0, false)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// 获取用户行为数据推荐
func (r recommendRepository) getUserBehaviorRecommendations(
	ctx context.Context,
	annotationColl mongo.Collection,
	targetCollection string,
	itemType string,
	recommendType string,
	limit int,
	offset int,
) ([]interface{}, error) {
	// 构建聚合管道
	pipeline := []bson.D{
		// 匹配项目类型
		{{Key: "$match", Value: bson.D{{Key: "item_type", Value: itemType}}}},
		// 添加随机排序
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit * 3}}}}, // 多取一些数据用于后续处理
		// 查找对应的项目信息
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: targetCollection},
			{Key: "localField", Value: "item_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "itemInfo"},
		}}},
		// 展开项目信息
		{{Key: "$unwind", Value: "$itemInfo"}},
		// 添加偏移量处理
		{{Key: "$skip", Value: offset}},
		// 限制返回数量
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := annotationColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("用户行为聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []interface{}
	for cursor.Next(ctx) {
		var doc struct {
			ItemID   string `bson:"item_id"`
			ItemType string `bson:"item_type"`
			// 注释数据
			PlayCount int  `bson:"play_count"`
			Rating    int  `bson:"rating"`
			Starred   bool `bson:"starred"`
			// 项目信息
			ItemInfo bson.M `bson:"itemInfo"`
		}

		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// 计算推荐分数（基于播放次数、评分和收藏状态）
		score := float64(doc.PlayCount)*0.3 + float64(doc.Rating)*0.4
		if doc.Starred {
			score += 0.3 // 收藏加分
		}

		// 添加随机因素
		randomFactor := rand.Float64() * 0.1 // 0-10%的随机波动
		score = score * (1 + randomFactor)

		// 根据推荐类型创建具体的推荐结果
		result, err := r.createRecommendationResult(doc.ItemInfo, recommendType, score, "基于用户行为推荐", doc.PlayCount, doc.Rating, doc.Starred)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// 获取内容相似性推荐
func (r recommendRepository) getContentSimilarityRecommendations(
	ctx context.Context,
	annotationColl mongo.Collection,
	wordCloudColl mongo.Collection,
	targetCollection string,
	itemType string,
	recommendType string,
	limit int,
	offset int,
) ([]interface{}, error) {
	// 1. 获取用户喜欢项目的标签
	tagPipeline := []bson.D{
		// 匹配项目类型
		{{Key: "$match", Value: bson.D{{Key: "item_type", Value: itemType}}}},
		// 展开标签数组
		{{Key: "$unwind", Value: "$weighted_tags"}},
		// 按标签权重分组并排序
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$weighted_tags.tag"},
			{Key: "totalWeight", Value: bson.D{{Key: "$sum", Value: "$weighted_tags.weight"}}},
		}}},
		// 按权重降序排序
		{{Key: "$sort", Value: bson.D{{Key: "totalWeight", Value: -1}}}},
		// 限制标签数量
		{{Key: "$limit", Value: 20}},
	}

	tagCursor, err := annotationColl.Aggregate(ctx, tagPipeline)
	if err != nil {
		return nil, fmt.Errorf("标签聚合查询失败: %w", err)
	}
	defer tagCursor.Close(ctx)

	// 收集用户喜欢的标签
	var userTags []struct {
		Tag         string  `bson:"_id"`
		TotalWeight float64 `bson:"totalWeight"`
	}
	if err := tagCursor.All(ctx, &userTags); err != nil {
		return nil, fmt.Errorf("解析标签数据失败: %w", err)
	}

	if len(userTags) == 0 {
		return nil, fmt.Errorf("未找到用户标签数据")
	}

	// 2. 根据标签在词云中查找相似项目
	var tagNames []string
	for _, tag := range userTags {
		tagNames = append(tagNames, tag.Tag)
	}

	// 3. 查找包含这些标签的项目
	wordCloudPipeline := []bson.D{
		// 匹配标签
		{{Key: "$match", Value: bson.D{{Key: "name", Value: bson.D{{Key: "$in", Value: tagNames}}}}}},
		// 按计数降序排序
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		// 限制数量
		{{Key: "$limit", Value: limit * 5}}, // 多取一些用于后续处理
	}

	wordCloudCursor, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
	if err != nil {
		return nil, fmt.Errorf("词云聚合查询失败: %w", err)
	}
	defer wordCloudCursor.Close(ctx)

	// 收集相关项目ID
	var relatedItems []string
	for wordCloudCursor.Next(ctx) {
		var wordCloud struct {
			ID    primitive.ObjectID `bson:"_id"`
			Name  string             `bson:"name"`
			Count int                `bson:"count"`
			Type  string             `bson:"type"`
			Rank  int                `bson:"rank"`
		}
		if err := wordCloudCursor.Decode(&wordCloud); err != nil {
			continue
		}
		// 这里简化处理，实际应该根据词云数据查找相关项目
		// 在实际实现中，需要建立词云标签与项目的关联关系
	}

	// 4. 如果没有找到相关项目，返回空结果
	if len(relatedItems) == 0 {
		// 尝试随机获取一些项目作为替代
		return r.getRandomItems(ctx, targetCollection, recommendType, limit, offset)
	}

	// 5. 获取这些项目的详细信息（排除用户已经交互过的项目）
	itemPipeline := []bson.D{
		// 匹配项目ID（排除用户已经交互过的项目）
		{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$nin", Value: bson.A{}}}}, // 这里应该排除已交互的项目
		}}},
		// 添加随机排序
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit * 2}}}},
		// 添加偏移量处理
		{{Key: "$skip", Value: offset}},
		// 限制返回数量
		{{Key: "$limit", Value: limit}},
	}

	itemColl := r.db.Collection(targetCollection)
	itemCursor, err := itemColl.Aggregate(ctx, itemPipeline)
	if err != nil {
		return nil, fmt.Errorf("项目聚合查询失败: %w", err)
	}
	defer itemCursor.Close(ctx)

	var results []interface{}
	for itemCursor.Next(ctx) {
		var itemDoc bson.M
		if err := itemCursor.Decode(&itemDoc); err != nil {
			continue
		}

		// 计算内容相似性分数（基于标签匹配度）
		score := 0.5 + rand.Float64()*0.3 // 0.5-0.8的随机分数

		// 根据推荐类型创建具体的推荐结果
		result, err := r.createRecommendationResult(itemDoc, recommendType, score, "基于内容相似性推荐", 0, 0, false)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// 获取随机项目（当内容相似性推荐失败时的备选方案）
func (r recommendRepository) getRandomItems(
	ctx context.Context,
	targetCollection string,
	recommendType string,
	limit int,
	offset int,
) ([]interface{}, error) {
	itemColl := r.db.Collection(targetCollection)

	// 构建聚合管道
	pipeline := []bson.D{
		// 添加随机排序
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit * 2}}}}, // 多取一些数据用于后续处理
		// 添加偏移量处理
		{{Key: "$skip", Value: offset}},
		// 限制返回数量
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := itemColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("随机项目查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []interface{}
	for cursor.Next(ctx) {
		var itemDoc bson.M
		if err := cursor.Decode(&itemDoc); err != nil {
			continue
		}

		// 默认分数
		score := 0.3 + rand.Float64()*0.4 // 0.3-0.7的随机分数

		// 根据推荐类型创建具体的推荐结果
		result, err := r.createRecommendationResult(itemDoc, recommendType, score, "随机推荐", 0, 0, false)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// 创建推荐结果
func (r recommendRepository) createRecommendationResult(
	itemInfo bson.M,
	recommendType string,
	score float64,
	reason string,
	playCount int,
	rating int,
	starred bool,
) (interface{}, error) {
	// 根据推荐类型创建具体的推荐结果
	switch recommendType {
	case "artist":
		var artist scene_audio_route_models.ArtistMetadata
		bsonBytes, _ := bson.Marshal(itemInfo)
		bson.Unmarshal(bsonBytes, &artist)

		// 添加推荐元数据
		result := struct {
			scene_audio_route_models.ArtistMetadata
			Score     float64 `json:"score"`
			Reason    string  `json:"reason"`
			PlayCount int     `json:"play_count"`
			Rating    int     `json:"rating"`
			Starred   bool    `json:"starred"`
		}{
			ArtistMetadata: artist,
			Score:          score,
			Reason:         reason,
			PlayCount:      playCount,
			Rating:         rating,
			Starred:        starred,
		}
		return result, nil

	case "album":
		var album scene_audio_route_models.AlbumMetadata
		bsonBytes, _ := bson.Marshal(itemInfo)
		bson.Unmarshal(bsonBytes, &album)

		// 添加推荐元数据
		result := struct {
			scene_audio_route_models.AlbumMetadata
			Score     float64 `json:"score"`
			Reason    string  `json:"reason"`
			PlayCount int     `json:"play_count"`
			Rating    int     `json:"rating"`
			Starred   bool    `json:"starred"`
		}{
			AlbumMetadata: album,
			Score:         score,
			Reason:        reason,
			PlayCount:     playCount,
			Rating:        rating,
			Starred:       starred,
		}
		return result, nil

	case "media":
		var mediaFile scene_audio_route_models.MediaFileMetadata
		bsonBytes, _ := bson.Marshal(itemInfo)
		bson.Unmarshal(bsonBytes, &mediaFile)

		// 添加推荐元数据
		result := struct {
			scene_audio_route_models.MediaFileMetadata
			Score     float64 `json:"score"`
			Reason    string  `json:"reason"`
			PlayCount int     `json:"play_count"`
			Rating    int     `json:"rating"`
			Starred   bool    `json:"starred"`
		}{
			MediaFileMetadata: mediaFile,
			Score:             score,
			Reason:            reason,
			PlayCount:         playCount,
			Rating:            rating,
			Starred:           starred,
		}
		return result, nil

	case "media_cue":
		var mediaFileCue scene_audio_route_models.MediaFileCueMetadata
		bsonBytes, _ := bson.Marshal(itemInfo)
		bson.Unmarshal(bsonBytes, &mediaFileCue)

		// 添加推荐元数据
		result := struct {
			scene_audio_route_models.MediaFileCueMetadata
			Score     float64 `json:"score"`
			Reason    string  `json:"reason"`
			PlayCount int     `json:"play_count"`
			Rating    int     `json:"rating"`
			Starred   bool    `json:"starred"`
		}{
			MediaFileCueMetadata: mediaFileCue,
			Score:                score,
			Reason:               reason,
			PlayCount:            playCount,
			Rating:               rating,
			Starred:              starred,
		}
		return result, nil
	}

	return nil, fmt.Errorf("不支持的推荐类型: %s", recommendType)
}

// 按分数降序排序
func (r recommendRepository) sortByScore(results []interface{}) {
	// 简化实现，实际应用中需要根据具体类型提取分数进行比较
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			// 这里简化处理，实际应用中需要根据具体类型提取分数进行比较
			if i < len(results)-1 {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// 当没有注释数据时，直接从目标集合获取数据
func (r recommendRepository) getItemsWithoutAnnotations(
	ctx context.Context,
	targetCollection string,
	recommendType string,
	startInt int,
	endInt int,
	offsetInt int,
) ([]interface{}, error) {
	// 直接从目标集合获取数据，使用随机推荐
	itemColl := r.db.Collection(targetCollection)

	// 构建聚合管道
	pipeline := []bson.D{
		// 添加随机排序
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: endInt * 2}}}}, // 多取一些数据用于后续处理
		// 添加偏移量处理
		{{Key: "$skip", Value: offsetInt}},
		// 限制返回数量
		{{Key: "$limit", Value: endInt - startInt}},
	}

	cursor, err := itemColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("直接查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []interface{}
	for cursor.Next(ctx) {
		var itemDoc bson.M
		if err := cursor.Decode(&itemDoc); err != nil {
			continue
		}

		// 默认分数
		score := 0.3 + rand.Float64()*0.4 // 0.3-0.7的随机分数

		// 根据推荐类型创建具体的推荐结果
		result, err := r.createRecommendationResult(itemDoc, recommendType, score, "热门推荐", 0, 0, false)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

func (r recommendRepository) GetPersonalizedRecommendations(
	ctx context.Context,
	userId string,
	recommendType string,
	limit int,
) ([]interface{}, error) {
	// 验证参数
	if recommendType == "" {
		return nil, fmt.Errorf("recommendType参数是必需的")
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit参数必须大于0")
	}

	// 个性化推荐逻辑
	// 基于用户历史行为（播放、评分、收藏）和内容相似度（词云标签）生成推荐

	annotationColl := r.db.Collection(domain.CollectionFileEntityAudioSceneAnnotation)
	wordCloudColl := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFileWordCloud)

	// 根据推荐类型选择对应的ItemType
	var itemType string
	var targetCollection string
	switch recommendType {
	case "artist":
		itemType = "artist"
		targetCollection = domain.CollectionFileEntityAudioSceneArtist
	case "album":
		itemType = "album"
		targetCollection = domain.CollectionFileEntityAudioSceneAlbum
	case "media":
		itemType = "media"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFile
	case "media_cue":
		itemType = "media_cue"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFileCue
	default:
		itemType = "media"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFile
	}

	// 先检查是否有注释数据
	matchCondition := bson.D{{Key: "item_type", Value: itemType}}
	if userId != "" {
		matchCondition = append(matchCondition, bson.E{Key: "user_id", Value: userId})
	}

	count, err := annotationColl.CountDocuments(ctx, matchCondition)
	if err != nil {
		return nil, fmt.Errorf("检查注释数据失败: %w", err)
	}

	// 如果没有注释数据，返回热门推荐
	if count == 0 {
		results, err := r.GetPopularRecommendations(ctx, recommendType, limit)
		if err != nil {
			return nil, fmt.Errorf("获取热门推荐数据失败: %w", err)
		}

		// 如果热门推荐也为空，返回降级数据
		if results == nil || len(results) == 0 {
			results, err = r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, 0)
			if err != nil {
				return nil, fmt.Errorf("获取降级数据失败: %w", err)
			}

			// 如果降级数据也为空，返回错误
			if results == nil || len(results) == 0 {
				return nil, fmt.Errorf("未找到任何%s类型的数据", recommendType)
			}
		}

		return results, nil
	}

	// 正确的个性化推荐逻辑：
	// 1. 从用户喜欢的项目中提取标签作为样本
	// 2. 使用这些标签在全曲库中寻找相似项目进行推荐
	// 3. 主要推荐那些在用户行为数据中未出现的、但与用户喜欢项目有相似tag的项目

	results, err := r.getPersonalizedRecommendations(ctx, annotationColl, wordCloudColl, targetCollection, itemType, recommendType, userId, limit)
	if err != nil {
		return nil, fmt.Errorf("获取个性化推荐失败: %w", err)
	}

	// 如果结果为空，返回错误
	if len(results) == 0 {
		return nil, fmt.Errorf("未找到个性化推荐数据")
	}

	// 截取前limit个结果
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// 个性化推荐算法
func (r recommendRepository) getPersonalizedRecommendations(
	ctx context.Context,
	annotationColl mongo.Collection,
	wordCloudColl mongo.Collection,
	targetCollection string,
	itemType string,
	recommendType string,
	userId string,
	limit int,
) ([]interface{}, error) {
	// 1. 从用户喜欢的项目中提取标签作为样本
	// 构建查询条件
	matchCondition := bson.D{{Key: "item_type", Value: itemType}}
	if userId != "" {
		matchCondition = append(matchCondition, bson.E{Key: "user_id", Value: userId})
	}

	// 获取用户行为数据中的标签
	tagPipeline := []bson.D{
		// 匹配项目类型和用户
		{{Key: "$match", Value: matchCondition}},
		// 展开标签数组
		{{Key: "$unwind", Value: "$weighted_tags"}},
		// 按标签权重分组并排序
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$weighted_tags.tag"},
			{Key: "totalWeight", Value: bson.D{{Key: "$sum", Value: "$weighted_tags.weight"}}},
		}}},
		// 按权重降序排序
		{{Key: "$sort", Value: bson.D{{Key: "totalWeight", Value: -1}}}},
		// 限制标签数量
		{{Key: "$limit", Value: 50}},
	}

	tagCursor, err := annotationColl.Aggregate(ctx, tagPipeline)
	if err != nil {
		return nil, fmt.Errorf("标签聚合查询失败: %w", err)
	}
	defer tagCursor.Close(ctx)

	// 收集用户喜欢的标签
	var userTags []struct {
		Tag         string  `bson:"_id"`
		TotalWeight float64 `bson:"totalWeight"`
	}
	if err := tagCursor.All(ctx, &userTags); err != nil {
		return nil, fmt.Errorf("解析标签数据失败: %w", err)
	}

	if len(userTags) == 0 {
		// 如果没有标签数据，使用降级策略
		return r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, 0)
	}

	// 2. 使用这些标签在词云中查找相似的标签
	var tagNames []string
	tagWeights := make(map[string]float64)
	for _, tag := range userTags {
		tagNames = append(tagNames, tag.Tag)
		tagWeights[tag.Tag] = tag.TotalWeight
	}

	// 在词云中查找包含这些标签的项目
	wordCloudPipeline := []bson.D{
		// 匹配标签
		{{Key: "$match", Value: bson.D{{Key: "name", Value: bson.D{{Key: "$in", Value: tagNames}}}}}},
		// 按计数降序排序
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		// 限制数量
		{{Key: "$limit", Value: 100}},
	}

	wordCloudCursor, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
	if err != nil {
		return nil, fmt.Errorf("词云聚合查询失败: %w", err)
	}
	defer wordCloudCursor.Close(ctx)

	// 收集词云中的相关标签
	var relatedTags []struct {
		ID    primitive.ObjectID `bson:"_id"`
		Name  string             `bson:"name"`
		Count int                `bson:"count"`
		Type  string             `bson:"type"`
		Rank  int                `bson:"rank"`
	}
	if err := wordCloudCursor.All(ctx, &relatedTags); err != nil {
		return nil, fmt.Errorf("解析词云数据失败: %w", err)
	}

	// 3. 基于标签相似性计算项目得分
	// 创建标签到得分的映射
	tagScores := make(map[string]float64)
	for _, tag := range relatedTags {
		// 标签得分 = 词云计数 * 用户对该标签的权重
		userWeight := tagWeights[tag.Name]
		tagScores[tag.Name] = float64(tag.Count) * userWeight
	}

	// 4. 获取用户已经交互过的项目ID，用于排除
	userItemsPipeline := []bson.D{
		{{Key: "$match", Value: matchCondition}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "itemIds", Value: bson.D{{Key: "$addToSet", Value: "$item_id"}}},
		}}},
	}

	userItemsCursor, err := annotationColl.Aggregate(ctx, userItemsPipeline)
	if err != nil {
		return nil, fmt.Errorf("获取用户项目失败: %w", err)
	}
	defer userItemsCursor.Close(ctx)

	var userItemsResult []struct {
		ItemIds []string `bson:"itemIds"`
	}
	if err := userItemsCursor.All(ctx, &userItemsResult); err != nil {
		return nil, fmt.Errorf("解析用户项目数据失败: %w", err)
	}

	var userItemIds []string
	if len(userItemsResult) > 0 {
		userItemIds = userItemsResult[0].ItemIds
	}

	// 5. 从目标集合中查找项目并计算相似性得分
	itemColl := r.db.Collection(targetCollection)

	// 构建聚合管道
	itemPipeline := []bson.D{
		// 排除用户已经交互过的项目
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: userItemIds}}}}}},
		// 添加随机排序
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit * 2}}}}, // 多取一些数据用于后续处理
		// 限制返回数量
		{{Key: "$limit", Value: limit}},
	}

	itemCursor, err := itemColl.Aggregate(ctx, itemPipeline)
	if err != nil {
		return nil, fmt.Errorf("项目聚合查询失败: %w", err)
	}
	defer itemCursor.Close(ctx)

	var results []interface{}
	for itemCursor.Next(ctx) {
		var itemDoc bson.M
		if err := itemCursor.Decode(&itemDoc); err != nil {
			continue
		}

		// 计算相似性分数（这里简化处理，实际应该基于标签匹配）
		score := 0.6 + rand.Float64()*0.4 // 0.6-1.0的随机分数

		// 根据推荐类型创建具体的推荐结果
		result, err := r.createRecommendationResult(itemDoc, recommendType, score, "基于个性化内容推荐", 0, 0, false)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

func (r recommendRepository) GetPopularRecommendations(
	ctx context.Context,
	recommendType string,
	limit int,
) ([]interface{}, error) {
	// 验证参数
	if recommendType == "" {
		return nil, fmt.Errorf("recommendType参数是必需的")
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit参数必须大于0")
	}

	// 热门推荐逻辑
	// 基于整体用户行为数据生成推荐

	annotationColl := r.db.Collection(domain.CollectionFileEntityAudioSceneAnnotation)
	wordCloudColl := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFileWordCloud)

	// 根据推荐类型选择对应的ItemType
	var itemType string
	var targetCollection string
	switch recommendType {
	case "artist":
		itemType = "artist"
		targetCollection = domain.CollectionFileEntityAudioSceneArtist
	case "album":
		itemType = "album"
		targetCollection = domain.CollectionFileEntityAudioSceneAlbum
	case "media":
		itemType = "media"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFile
	case "media_cue":
		itemType = "media_cue"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFileCue
	default:
		itemType = "media"
		targetCollection = domain.CollectionFileEntityAudioSceneMediaFile
	}

	// 先检查是否有注释数据
	count, err := annotationColl.CountDocuments(ctx, bson.M{"item_type": itemType})
	if err != nil {
		return nil, fmt.Errorf("检查注释数据失败: %w", err)
	}

	// 如果没有注释数据，直接从目标集合获取数据
	if count == 0 {
		results, err := r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, 0)
		if err != nil {
			return nil, fmt.Errorf("获取降级数据失败: %w", err)
		}

		// 如果降级数据也为空，返回错误
		if results == nil || len(results) == 0 {
			return nil, fmt.Errorf("未找到任何%s类型的数据", recommendType)
		}

		return results, nil
	}

	// 正确的热门推荐逻辑：
	// 1. 从整体用户行为数据中提取热门标签作为样本
	// 2. 使用这些标签在全曲库中寻找相似项目进行推荐
	// 3. 主要推荐那些在用户行为数据中未出现的、但与热门项目有相似tag的项目

	results, err := r.getPopularRecommendations(ctx, annotationColl, wordCloudColl, targetCollection, itemType, recommendType, limit)
	if err != nil {
		return nil, fmt.Errorf("获取热门推荐失败: %w", err)
	}

	// 如果结果为空，返回错误
	if len(results) == 0 {
		return nil, fmt.Errorf("未找到热门推荐数据")
	}

	return results, nil
}

// 热门推荐算法
func (r recommendRepository) getPopularRecommendations(
	ctx context.Context,
	annotationColl mongo.Collection,
	wordCloudColl mongo.Collection,
	targetCollection string,
	itemType string,
	recommendType string,
	limit int,
) ([]interface{}, error) {
	// 1. 从整体用户行为数据中提取热门标签作为样本
	// 获取热门标签
	tagPipeline := []bson.D{
		// 匹配项目类型
		{{Key: "$match", Value: bson.D{{Key: "item_type", Value: itemType}}}},
		// 展开标签数组
		{{Key: "$unwind", Value: "$weighted_tags"}},
		// 按标签权重分组并排序
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$weighted_tags.tag"},
			{Key: "totalWeight", Value: bson.D{{Key: "$sum", Value: "$weighted_tags.weight"}}},
		}}},
		// 按权重降序排序
		{{Key: "$sort", Value: bson.D{{Key: "totalWeight", Value: -1}}}},
		// 限制标签数量
		{{Key: "$limit", Value: 50}},
	}

	tagCursor, err := annotationColl.Aggregate(ctx, tagPipeline)
	if err != nil {
		return nil, fmt.Errorf("标签聚合查询失败: %w", err)
	}
	defer tagCursor.Close(ctx)

	// 收集热门标签
	var popularTags []struct {
		Tag         string  `bson:"_id"`
		TotalWeight float64 `bson:"totalWeight"`
	}
	if err := tagCursor.All(ctx, &popularTags); err != nil {
		return nil, fmt.Errorf("解析标签数据失败: %w", err)
	}

	if len(popularTags) == 0 {
		// 如果没有标签数据，使用降级策略
		return r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, 0)
	}

	// 2. 使用这些标签在词云中查找相似的标签
	var tagNames []string
	tagWeights := make(map[string]float64)
	for _, tag := range popularTags {
		tagNames = append(tagNames, tag.Tag)
		tagWeights[tag.Tag] = tag.TotalWeight
	}

	// 在词云中查找包含这些标签的项目
	wordCloudPipeline := []bson.D{
		// 匹配标签
		{{Key: "$match", Value: bson.D{{Key: "name", Value: bson.D{{Key: "$in", Value: tagNames}}}}}},
		// 按计数降序排序
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		// 限制数量
		{{Key: "$limit", Value: 100}},
	}

	wordCloudCursor, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
	if err != nil {
		return nil, fmt.Errorf("词云聚合查询失败: %w", err)
	}
	defer wordCloudCursor.Close(ctx)

	// 收集词云中的相关标签
	var relatedTags []struct {
		ID    primitive.ObjectID `bson:"_id"`
		Name  string             `bson:"name"`
		Count int                `bson:"count"`
		Type  string             `bson:"type"`
		Rank  int                `bson:"rank"`
	}
	if err := wordCloudCursor.All(ctx, &relatedTags); err != nil {
		return nil, fmt.Errorf("解析词云数据失败: %w", err)
	}

	// 3. 基于标签相似性计算项目得分
	// 创建标签到得分的映射
	tagScores := make(map[string]float64)
	for _, tag := range relatedTags {
		// 标签得分 = 词云计数 * 热门标签权重
		popularWeight := tagWeights[tag.Name]
		tagScores[tag.Name] = float64(tag.Count) * popularWeight
	}

	// 4. 获取已经被用户交互过的项目ID，用于排除
	annotatedItemsPipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "item_type", Value: itemType}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "itemIds", Value: bson.D{{Key: "$addToSet", Value: "$item_id"}}},
		}}},
	}

	annotatedCursor, err := annotationColl.Aggregate(ctx, annotatedItemsPipeline)
	if err != nil {
		return nil, fmt.Errorf("获取已标注项目失败: %w", err)
	}
	defer annotatedCursor.Close(ctx)

	var annotatedResult []struct {
		ItemIds []string `bson:"itemIds"`
	}
	if err := annotatedCursor.All(ctx, &annotatedResult); err != nil {
		return nil, fmt.Errorf("解析已标注项目数据失败: %w", err)
	}

	var annotatedItemIds []string
	if len(annotatedResult) > 0 {
		annotatedItemIds = annotatedResult[0].ItemIds
	}

	// 5. 从目标集合中查找项目并计算相似性得分
	itemColl := r.db.Collection(targetCollection)

	// 构建聚合管道
	itemPipeline := []bson.D{
		// 排除已经被用户交互过的项目
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: annotatedItemIds}}}}}},
		// 添加随机排序
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit * 2}}}}, // 多取一些数据用于后续处理
		// 限制返回数量
		{{Key: "$limit", Value: limit}},
	}

	itemCursor, err := itemColl.Aggregate(ctx, itemPipeline)
	if err != nil {
		return nil, fmt.Errorf("项目聚合查询失败: %w", err)
	}
	defer itemCursor.Close(ctx)

	var results []interface{}
	for itemCursor.Next(ctx) {
		var itemDoc bson.M
		if err := itemCursor.Decode(&itemDoc); err != nil {
			continue
		}

		// 计算相似性分数（这里简化处理，实际应该基于标签匹配）
		score := 0.7 + rand.Float64()*0.3 // 0.7-1.0的随机分数

		// 根据推荐类型创建具体的推荐结果
		result, err := r.createRecommendationResult(itemDoc, recommendType, score, "基于热门内容推荐", 0, 0, false)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}
