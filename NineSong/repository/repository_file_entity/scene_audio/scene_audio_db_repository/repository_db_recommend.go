package scene_audio_db_repository

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DataFetcher 定义数据获取接口
type DataFetcher interface {
	GetAnnotations(ctx context.Context, itemType string, algorithmType string) ([]scene_audio_db_models.AnnotationMetadata, error)
	GetItems(ctx context.Context, targetCollection string, annotations []scene_audio_db_models.AnnotationMetadata) ([]bson.M, error)
}

// TagExtractor 定义标签提取接口
type TagExtractor interface {
	ExtractTags(ctx context.Context, items []bson.M) ([]string, map[string]int, error)
}

// SimilarityCalculator 定义相似度计算接口
type SimilarityCalculator interface {
	CalculateSimilarity(itemText string, tagName string) bool
}

// ResultRanker 定义结果排序接口
type ResultRanker interface {
	RankResults(items []bson.M, tagNameToCount map[string]int, algorithmType string) []bson.M
}

type recommendRepository struct {
	db         mongo.Database
	collection string

	// 缓存相关字段
	wordCloudCache map[string][]scene_audio_db_models.WordCloudMetadata
	cacheMutex     sync.RWMutex
	cacheExpiry    time.Time
	logShow        bool // 控制是否输出日志
}

// RecommendationPipeline 推荐流程管道
type RecommendationPipeline struct {
	dataFetcher    DataFetcher
	tagExtractor   TagExtractor
	similarityCalc SimilarityCalculator
	ranker         ResultRanker
}

func NewRecommendRepository(db mongo.Database, collection string) scene_audio_route_interface.RecommendRouteRepository {
	return &recommendRepository{
		db:         db,
		collection: collection,
		logShow:    true, // 默认输出日志
	}
}

// GetGeneralRecommendations - 通用推荐接口
func (r *recommendRepository) GetGeneralRecommendations(
	ctx context.Context,
	recommendType string,
	limit int,
	randomSeed string,
	recommendOffset string,
	logShow bool,
) ([]interface{}, error) {
	// 设置日志输出控制
	r.logShow = logShow

	// 验证参数
	if limit <= 0 {
		return nil, fmt.Errorf("limit参数必须大于0")
	}

	// 解析参数
	recommendOffsetInt, err := strconv.Atoi(recommendOffset)
	if err != nil {
		return nil, fmt.Errorf("无效的offset参数: %w", err)
	}

	// 设置随机种子
	seed, err := strconv.ParseInt(randomSeed, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的randomSeed参数: %w", err)
	}
	rand.Seed(seed)

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

	// 使用统一的推荐流程
	results, err := r.getUnifiedRecommendationWorkflow(ctx, itemType, targetCollection, recommendType, limit, recommendOffsetInt, "general", seed, logShow)
	if err != nil {
		return nil, fmt.Errorf("获取推荐数据失败: %w", err)
	}

	// 如果结果为空，返回错误
	if len(results) == 0 {
		return nil, fmt.Errorf("未找到推荐数据")
	}

	return results, nil
}

// GetPersonalizedRecommendations - 个性化推荐接口
func (r *recommendRepository) GetPersonalizedRecommendations(
	ctx context.Context,
	userId string,
	recommendType string,
	limit int,
	logShow bool,
) ([]interface{}, error) {
	// 设置日志输出控制
	r.logShow = logShow

	// 验证参数
	if recommendType == "" {
		return nil, fmt.Errorf("recommendType参数是必需的")
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit参数必须大于0")
	}

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

	// 使用当前时间作为随机种子
	seed := time.Now().UnixNano()

	// 使用统一的推荐流程，但可以加入个性化策略
	results, err := r.getUnifiedRecommendationWorkflow(ctx, itemType, targetCollection, recommendType, limit, 0, "personalized", seed, logShow)
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

// GetPopularRecommendations - 热门推荐接口
func (r *recommendRepository) GetPopularRecommendations(
	ctx context.Context,
	recommendType string,
	limit int,
	logShow bool,
) ([]interface{}, error) {
	// 设置日志输出控制
	r.logShow = logShow

	// 验证参数
	if recommendType == "" {
		return nil, fmt.Errorf("recommendType参数是必需的")
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit参数必须大于0")
	}

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

	// 使用当前时间作为随机种子
	seed := time.Now().UnixNano()

	// 使用统一的推荐流程，但可以加入热门度策略
	results, err := r.getUnifiedRecommendationWorkflow(ctx, itemType, targetCollection, recommendType, limit, 0, "popular", seed, logShow)
	if err != nil {
		return nil, fmt.Errorf("获取热门推荐失败: %w", err)
	}

	// 如果结果为空，返回错误
	if len(results) == 0 {
		return nil, fmt.Errorf("未找到热门推荐数据")
	}

	return results, nil
}

// 统一的推荐流程
func (r *recommendRepository) getUnifiedRecommendationWorkflow(
	ctx context.Context,
	itemType string,
	targetCollection string,
	recommendType string,
	limit int,
	recommendOffset int,
	algorithmType string,
	randomSeed int64,
	logShow bool,
) ([]interface{}, error) {

	if r.logShow {
		fmt.Printf("开始统一推荐流程，itemType=%s, algorithmType=%s\n", itemType, algorithmType)
	}

	// 创建推荐流程管道
	pipeline := &RecommendationPipeline{
		dataFetcher:    r,
		tagExtractor:   r,
		similarityCalc: r,
		ranker:         r,
	}

	// 步骤1: 从annotation集合中获取用户行为数据
	annotations, err := pipeline.dataFetcher.GetAnnotations(ctx, itemType, algorithmType)
	if err != nil {
		return nil, fmt.Errorf("获取注释数据失败: %w", err)
	}

	// 步骤2: 根据annotation的item_id和item_type寻找对应的项
	items, err := pipeline.dataFetcher.GetItems(ctx, targetCollection, annotations)
	if err != nil {
		return nil, fmt.Errorf("获取项目数据失败: %w", err)
	}

	if r.logShow {
		fmt.Printf("找到%d个项目数据\n", len(items))
	}

	// 如果没有项目数据，使用降级策略
	if len(items) == 0 {
		if r.logShow {
			fmt.Printf("没有找到项目数据，使用降级策略\n")
		}
		return r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, recommendOffset, randomSeed)
	}

	// 步骤3: 从项目中提取标签信息
	allTagNames, _, err := pipeline.tagExtractor.ExtractTags(ctx, items)
	if err != nil {
		return nil, fmt.Errorf("提取标签失败: %w", err)
	}

	// 如果没有标签，使用降级策略
	if len(allTagNames) == 0 {
		if r.logShow {
			fmt.Printf("没有从项目中提取到标签，使用降级策略\n")
		}
		return r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, recommendOffset, randomSeed)
	}

	// 步骤4: 在词云数据表中查找相似标签
	// 将itemType映射到词云数据中的type值
	wordCloudType := itemType
	if itemType == "media" {
		wordCloudType = "media_file"
	} else if itemType == "media_cue" {
		wordCloudType = "media_file_cue"
	}

	// 添加调试信息，检查实际的词云类型
	if r.logShow {
		fmt.Printf("映射后的词云类型: %s\n", wordCloudType)
	}

	// 获取词云集合
	wordCloudColl := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFileWordCloud)

	// 添加调试信息，检查数据库中是否有词云数据
	if count, err := wordCloudColl.CountDocuments(ctx, bson.M{}); err != nil {
		if r.logShow {
			fmt.Printf("词云数据计数查询失败: %v\n", err)
		}
	} else {
		if r.logShow {
			fmt.Printf("词云数据总数量: %d\n", count)
		}
	}

	// 检查特定类型的词云数据数量
	if typeCount, err := wordCloudColl.CountDocuments(ctx, bson.M{"type": wordCloudType}); err != nil {
		if r.logShow {
			fmt.Printf("特定类型词云数据计数查询失败: %v\n", err)
		}
	} else {
		if r.logShow {
			fmt.Printf("类型为%s的词云数据数量: %d\n", wordCloudType, typeCount)
		}
	}

	// 初始化词云标签数组
	var wordCloudTags []scene_audio_db_models.WordCloudMetadata

	// 首先尝试使用缓存获取特定类型的词云标签
	cachedWordCloudTags, err := r.getCachedWordCloudTagsByType(ctx, wordCloudType)
	if err != nil {
		if r.logShow {
			fmt.Printf("获取缓存的词云标签失败: %v\n", err)
		}
		// 如果缓存获取失败，使用原始查询方式
		wordCloudPipeline := []bson.D{
			{{"$match", bson.D{
				{"name", bson.D{{"$in", allTagNames}}},
				{"type", wordCloudType},
			}}},
			{{"$sort", bson.D{{"count", -1}}}},
			{{"$limit", 100}},
		}

		// 添加调试信息
		if r.logShow {
			fmt.Printf("词云查询条件: 标签数量=%d, 类型=%s\n", len(allTagNames), wordCloudType)
		}
		if len(allTagNames) > 0 {
			if r.logShow {
				fmt.Printf("第一个标签: %s\n", allTagNames[0])
			}
		}

		// 添加调试信息
		if r.logShow {
			fmt.Printf("词云查询Pipeline: %+v\n", wordCloudPipeline)
		}

		wordCloudCursor1, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
		if err := r.handleError("词云数据查询", err); err != nil {
			r.logDebug("查询Pipeline: %+v", wordCloudPipeline)
			return nil, err
		}
		defer wordCloudCursor1.Close(ctx)

		if err := wordCloudCursor1.All(ctx, &wordCloudTags); err != nil {
			if err := r.handleError("解析词云数据", err); err != nil {
				return nil, err
			}
		}
	} else {
		// 在缓存中查找匹配的标签
		var matchedTags []scene_audio_db_models.WordCloudMetadata
		tagNameSet := make(map[string]bool)
		for _, tagName := range allTagNames {
			tagNameSet[tagName] = true
		}

		for _, wordCloudTag := range cachedWordCloudTags {
			if _, exists := tagNameSet[wordCloudTag.Name]; exists {
				matchedTags = append(matchedTags, wordCloudTag)
			}
		}

		// 如果在缓存中找到了匹配的标签，使用这些标签
		if len(matchedTags) > 0 {
			// 按count排序
			sort.Slice(matchedTags, func(i, j int) bool {
				return matchedTags[i].Count > matchedTags[j].Count
			})

			// 限制返回数量
			if len(matchedTags) > 100 {
				matchedTags = matchedTags[:100]
			}

			wordCloudTags = matchedTags
		} else {
			// 如果缓存中没有找到匹配的标签，使用原始查询方式
			wordCloudPipeline := []bson.D{
				{{"$match", bson.D{
					{"name", bson.D{{"$in", allTagNames}}},
					{"type", wordCloudType},
				}}},
				{{"$sort", bson.D{{"count", -1}}}},
				{{"$limit", 100}},
			}

			// 添加调试信息
			if r.logShow {
				fmt.Printf("词云查询条件: 标签数量=%d, 类型=%s\n", len(allTagNames), wordCloudType)
			}
			if len(allTagNames) > 0 {
				if r.logShow {
					fmt.Printf("第一个标签: %s\n", allTagNames[0])
				}
			}

			// 添加调试信息
			if r.logShow {
				fmt.Printf("词云查询Pipeline: %+v\n", wordCloudPipeline)
			}

			wordCloudCursor1, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
			if err != nil {
				if r.logShow {
					fmt.Printf("词云数据查询失败: %v\n", err)
				}
				if r.logShow {
					fmt.Printf("查询Pipeline: %+v\n", wordCloudPipeline)
				}
				return nil, fmt.Errorf("词云数据查询失败: %w", err)
			}
			defer wordCloudCursor1.Close(ctx)

			if err := wordCloudCursor1.All(ctx, &wordCloudTags); err != nil {
				return nil, fmt.Errorf("解析词云数据失败: %w", err)
			}
		}
	}

	// 如果没有找到匹配类型和标签的词云数据，则尝试只匹配标签名称
	if len(wordCloudTags) == 0 {
		if r.logShow {
			fmt.Printf("没有找到匹配类型和标签的词云数据，尝试只匹配标签名称\n")
		}

		// 使用缓存的默认词云标签
		cachedAllWordCloudTags, err := r.getCachedWordCloudTags(ctx)
		if err != nil {
			if r.logShow {
				fmt.Printf("获取缓存的所有词云标签失败: %v\n", err)
			}
			// 如果缓存获取失败，使用原始查询方式
			wordCloudPipeline := []bson.D{
				{{"$match", bson.D{{"name", bson.D{{"$in", allTagNames}}}}}},
				{{"$sort", bson.D{{"count", -1}}}},
				{{"$limit", 100}},
			}

			wordCloudCursor2, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
			if err != nil {
				if r.logShow {
					fmt.Printf("词云数据查询失败: %v\n", err)
				}
				if r.logShow {
					fmt.Printf("查询Pipeline: %+v\n", wordCloudPipeline)
				}
				return nil, fmt.Errorf("词云数据查询失败: %w", err)
			}
			defer wordCloudCursor2.Close(ctx)

			if err := wordCloudCursor2.All(ctx, &wordCloudTags); err != nil {
				return nil, fmt.Errorf("解析词云数据失败: %w", err)
			}
		} else {
			// 在缓存中查找匹配的标签
			var matchedTags []scene_audio_db_models.WordCloudMetadata
			tagNameSet := make(map[string]bool)
			for _, tagName := range allTagNames {
				tagNameSet[tagName] = true
			}

			for _, wordCloudTag := range cachedAllWordCloudTags {
				if _, exists := tagNameSet[wordCloudTag.Name]; exists {
					matchedTags = append(matchedTags, wordCloudTag)
				}
			}

			// 如果在缓存中找到了匹配的标签，使用这些标签
			if len(matchedTags) > 0 {
				// 按count排序
				sort.Slice(matchedTags, func(i, j int) bool {
					return matchedTags[i].Count > matchedTags[j].Count
				})

				// 限制返回数量
				if len(matchedTags) > 100 {
					matchedTags = matchedTags[:100]
				}

				wordCloudTags = matchedTags
			} else {
				// 如果缓存中没有找到匹配的标签，使用原始查询方式
				wordCloudPipeline := []bson.D{
					{{"$match", bson.D{{"name", bson.D{{"$in", allTagNames}}}}}},
					{{"$sort", bson.D{{"count", -1}}}},
					{{"$limit", 100}},
				}

				wordCloudCursor2, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
				if err != nil {
					if r.logShow {
						fmt.Printf("词云数据查询失败: %v\n", err)
					}
					if r.logShow {
						fmt.Printf("查询Pipeline: %+v\n", wordCloudPipeline)
					}
					return nil, fmt.Errorf("词云数据查询失败: %w", err)
				}
				defer wordCloudCursor2.Close(ctx)

				if err := wordCloudCursor2.All(ctx, &wordCloudTags); err != nil {
					return nil, fmt.Errorf("解析词云数据失败: %w", err)
				}
			}
		}
	}

	// 添加调试信息
	if r.logShow {
		fmt.Printf("找到%d个词云标签\n", len(wordCloudTags))
	}

	// 如果没有词云标签，尝试匹配genre标签
	if len(wordCloudTags) == 0 {
		if r.logShow {
			fmt.Printf("没有找到词云标签，尝试匹配genre标签\n")
		}

		// 构建genre匹配查询
		genrePipeline := []bson.D{
			{{"$match", bson.D{
				{"genre", bson.D{{"$in", allTagNames}}},
			}}},
			{{"$sample", bson.D{{"size", limit * 2}}}},
			{{"$limit", limit}},
		}

		// 获取集合引用
		itemColl := r.db.Collection(targetCollection)

		genreCursor, err := itemColl.Aggregate(ctx, genrePipeline)
		if err != nil {
			if r.logShow {
				fmt.Printf("genre标签查询失败: %v\n", err)
			}
			// 如果genre查询也失败，使用降级策略
			return r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, recommendOffset, randomSeed)
		}
		defer genreCursor.Close(ctx)

		var genreResults []interface{}
		for genreCursor.Next(ctx) {
			var itemDoc bson.M
			if err := genreCursor.Decode(&itemDoc); err != nil {
				continue
			}

			// 创建推荐结果
			score := 0.5 + rand.Float64()*0.3 // 0.5-0.8的随机分数
			result, err := r.createRecommendationResult(itemDoc, recommendType, score, "基于genre标签推荐", 0, 0, false, "GenreBasedAlgorithm", map[string]string{"limit": strconv.Itoa(limit), "offset": strconv.Itoa(recommendOffset)}, []string{"genre"}, []scene_audio_db_models.AnnotationMetadata{}, []scene_audio_db_models.WordCloudMetadata{}, []scene_audio_db_models.WordCloudRecommendation{})
			if err != nil {
				continue
			}
			genreResults = append(genreResults, result)
		}

		// 如果genre匹配成功，返回结果
		if len(genreResults) > 0 {
			if r.logShow {
				fmt.Printf("通过genre标签找到%d个推荐项\n", len(genreResults))
			}
			return genreResults, nil
		}

		// 如果genre匹配也失败，使用降级策略
		if r.logShow {
			fmt.Printf("没有找到genre标签匹配项，使用降级策略\n")
		}
		return r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, recommendOffset, randomSeed)
	}

	// 4. 使用这些tag在对应的数据库表中寻找推荐项并返回数据
	// 构建推荐标签列表（基于相关性动态选择标签）
	var recommendTagNames []string
	tagNameToCount := make(map[string]int)

	// 计算每个标签的相关性分数
	type TagScore struct {
		Name  string
		Score float64
		Count int
	}

	var tagScores []TagScore

	// 为每个词云标签计算相关性分数
	for _, tag := range wordCloudTags {
		// 基础分数为词云中的出现次数
		baseScore := float64(tag.Count)

		// 计算与项目标签的匹配度
		matchScore := 0.0
		for _, projectTagName := range allTagNames {
			if strings.EqualFold(tag.Name, projectTagName) {
				matchScore += 1.0
			} else if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(projectTagName)) ||
				strings.Contains(strings.ToLower(projectTagName), strings.ToLower(tag.Name)) {
				matchScore += 0.5
			}
		}

		// 综合分数 = 基础分数 * 匹配度权重
		finalScore := baseScore * (1.0 + matchScore*0.5)

		tagScores = append(tagScores, TagScore{
			Name:  tag.Name,
			Score: finalScore,
			Count: tag.Count,
		})
	}

	// 按相关性分数排序
	sort.Slice(tagScores, func(i, j int) bool {
		return tagScores[i].Score > tagScores[j].Score
	})

	// 根据相关性动态选择标签数量
	// 最多选择30个标签，最少选择5个标签
	maxTags := 30
	minTags := 5
	selectedTags := 0

	// 先选择高相关性的标签
	for _, tagScore := range tagScores {
		if selectedTags >= maxTags {
			break
		}

		recommendTagNames = append(recommendTagNames, tagScore.Name)
		tagNameToCount[tagScore.Name] = tagScore.Count
		selectedTags++
	}

	// 如果选择的标签数量少于最小值，补充一些genre标签
	if len(recommendTagNames) < minTags && len(allTagNames) > len(recommendTagNames) {
		additionalTagsNeeded := minTags - len(recommendTagNames)
		for _, tagName := range allTagNames {
			// 检查标签是否已经添加
			alreadyAdded := false
			for _, existingTag := range recommendTagNames {
				if existingTag == tagName {
					alreadyAdded = true
					break
				}
			}

			if !alreadyAdded && additionalTagsNeeded > 0 {
				recommendTagNames = append(recommendTagNames, tagName)
				// 对于genre标签，我们给一个默认的计数
				if _, exists := tagNameToCount[tagName]; !exists {
					tagNameToCount[tagName] = 10
				}
				additionalTagsNeeded--
			}
		}
	}

	// 添加调试信息
	if r.logShow {
		fmt.Printf("使用%d个标签进行推荐\n", len(recommendTagNames))
	}

	// 获取集合引用
	annotationColl := r.db.Collection(domain.CollectionFileEntityAudioSceneAnnotation)
	itemColl := r.db.Collection(targetCollection)

	// 获取已存在于annotation中的项目ID，用于排除
	annotatedItemsPipeline := []bson.D{
		{{"$match", bson.D{{"item_type", itemType}}}},
		{{"$group", bson.D{
			{"_id", nil},
			{"itemIds", bson.D{{"$addToSet", "$item_id"}}},
		}}},
	}

	annotatedCursor, err := annotationColl.Aggregate(ctx, annotatedItemsPipeline)
	if err := r.handleError("获取已标注项目", err); err != nil {
		return nil, err
	}
	defer annotatedCursor.Close(ctx)

	var annotatedResult []struct {
		ItemIds []string `bson:"itemIds"`
	}
	if err := annotatedCursor.All(ctx, &annotatedResult); err != nil {
		if err := r.handleError("解析已标注项目数据", err); err != nil {
			return nil, err
		}
	}

	// 将字符串类型的item_id转换为ObjectID
	var annotatedItemObjectIds []primitive.ObjectID
	if len(annotatedResult) > 0 {
		for _, itemId := range annotatedResult[0].ItemIds {
			if objectId, err := primitive.ObjectIDFromHex(itemId); err == nil {
				annotatedItemObjectIds = append(annotatedItemObjectIds, objectId)
			}
		}
	}

	// 添加调试信息
	if r.logShow {
		fmt.Printf("找到%d个已标注项目\n", len(annotatedItemObjectIds))
	}

	// 构建最终推荐查询
	// 构建匹配条件 - 先尝试宽松的条件
	recommendMatchCondition := bson.D{
		{"$or", []bson.D{
			{{"tags", bson.D{{"$in", recommendTagNames}}}},
			{{"genre", bson.D{{"$in", recommendTagNames}}}},
		}},
	}

	// 添加调试信息
	if r.logShow {
		fmt.Printf("推荐标签数量: %d\n", len(recommendTagNames))
		if len(recommendTagNames) > 0 {
			fmt.Printf("前5个推荐标签: %v\n", recommendTagNames[:int(math.Min(5, float64(len(recommendTagNames))))])
		}
		fmt.Printf("已标注项目数量: %d\n", len(annotatedItemObjectIds))
	}

	// 计算符合条件的项目数量（不排除已标注项目）
	matchCount, err := itemColl.CountDocuments(ctx, recommendMatchCondition)
	if err != nil {
		r.logInfo("计算符合条件的项目数量失败: %v", err)
	} else {
		r.logInfo("符合条件的项目数量（不排除已标注项目）: %d", matchCount)
	}

	// 如果没有符合条件的项目，尝试更宽松的条件
	if matchCount == 0 {
		// 尝试只匹配genre字段
		genreMatchCondition := bson.D{
			{"genre", bson.D{{"$in", recommendTagNames}}},
		}

		genreMatchCount, err := itemColl.CountDocuments(ctx, genreMatchCondition)
		if err != nil {
			r.logInfo("计算genre匹配的项目数量失败: %v", err)
		} else {
			r.logInfo("genre匹配的项目数量: %d", genreMatchCount)
			if genreMatchCount > 0 {
				recommendMatchCondition = genreMatchCondition
				matchCount = genreMatchCount
			}
		}
	}

	// 如果仍然没有符合条件的项目，尝试匹配文件名
	if matchCount == 0 {
		// 构建正则表达式来匹配文件名
		var regexConditions []bson.D
		for _, tagName := range recommendTagNames {
			if len(tagName) > 1 { // 只使用长度大于1的标签
				regexConditions = append(regexConditions, bson.D{{"file_name", bson.D{{"$regex", tagName}, {"$options", "i"}}}})
			}
		}

		if len(regexConditions) > 0 {
			fileNameMatchCondition := bson.D{
				{"$or", regexConditions},
			}

			fileNameMatchCount, err := itemColl.CountDocuments(ctx, fileNameMatchCondition)
			if err != nil {
				r.logInfo("计算文件名匹配的项目数量失败: %v", err)
			} else {
				r.logInfo("文件名匹配的项目数量: %d", fileNameMatchCount)
				if fileNameMatchCount > 0 {
					recommendMatchCondition = fileNameMatchCondition
					matchCount = fileNameMatchCount
				}
			}
		}
	}

	// 构建最终的推荐查询条件（添加排除已标注项目的条件）
	finalRecommendMatchCondition := recommendMatchCondition
	if len(annotatedItemObjectIds) > 0 {
		finalRecommendMatchCondition = bson.D{
			{"_id", bson.D{{"$nin", annotatedItemObjectIds}}}, // 排除已交互的项目
			{"$and", []bson.D{recommendMatchCondition}},
		}
	}

	recommendPipeline := []bson.D{
		// 匹配包含推荐标签的项目（排除用户已经交互过的项目）
		{{"$match", finalRecommendMatchCondition}},
		// 添加随机排序
		{{"$sample", bson.D{{"size", limit * 2}}}},
		// 添加偏移量处理
		{{"$skip", recommendOffset}},
		// 限制返回数量
		{{"$limit", limit}},
	}

	recommendCursor, err := itemColl.Aggregate(ctx, recommendPipeline)
	if err := r.handleError("推荐项目查询", err); err != nil {
		return nil, err
	}
	defer recommendCursor.Close(ctx)

	var results []interface{}
	for recommendCursor.Next(ctx) {
		var itemDoc bson.M
		if err := recommendCursor.Decode(&itemDoc); err != nil {
			continue
		}

		// 从注释数据中查找对应的播放日期
		var playDate *time.Time
		for _, annotation := range annotations {
			// 检查item_id是否匹配
			if itemID, ok := itemDoc["_id"]; ok {
				if objectID, ok := itemID.(primitive.ObjectID); ok {
					if objectID.Hex() == annotation.ItemID {
						playDate = &annotation.PlayDate
						break
					}
				}
			}
		}

		// 计算推荐分数，传入播放日期用于时间衰减计算
		score := r.calculateRecommendationScore(itemDoc, tagNameToCount, algorithmType, playDate)

		// 设置推荐理由和算法
		reason := "基于内容相似性推荐"
		algorithm := "ContentSimilarityAlgorithm"
		basis := []string{"tag_similarity"}

		switch algorithmType {
		case "personalized":
			reason = "基于用户行为推荐"
			algorithm = "UserBehaviorAlgorithm"
			basis = []string{"play_count", "rating", "starred", "play_complete_count", "play_date"}
		case "popular":
			reason = "基于热门度推荐"
			algorithm = "PopularityAlgorithm"
			basis = []string{"play_count"}
		}

		// 关键修复：确保在调用推荐依据函数之前，数据是有效的
		// 收集推荐依据信息
		annotationBasis := r.getAnnotationBasisFromAnnotations(annotations, 5)
		tagBasis := r.getTagBasisFromWordCloud(wordCloudTags, 10)
		relatedItems := r.getRelatedItemsFromWordCloud(wordCloudTags, 5)

		// 创建推荐结果 - 确保传入正确的依据数据
		result, err := r.createRecommendationResult(
			itemDoc,
			recommendType,
			score,
			reason,
			0, 0, false,
			algorithm,
			map[string]string{
				"limit":          strconv.Itoa(limit),
				"offset":         strconv.Itoa(recommendOffset),
				"algorithm_type": algorithmType,
			},
			basis, // 根据算法类型设置不同的 basis 参数
			annotationBasis,
			tagBasis,
			relatedItems,
		)
		if err != nil {
			continue
		}

		results = append(results, result)
	}

	if r.logShow {
		fmt.Printf("生成了%d个推荐结果\n", len(results))
	}

	// 如果没有生成推荐结果，使用降级策略
	if len(results) == 0 {
		if r.logShow {
			fmt.Printf("没有生成推荐结果，使用降级策略\n")
		}
		return r.getItemsWithoutAnnotations(ctx, targetCollection, recommendType, 0, limit, recommendOffset, randomSeed)
	}

	return results, nil
}

// 计算推荐分数
func (r *recommendRepository) calculateRecommendationScore(
	itemDoc bson.M,
	tagNameToCount map[string]int,
	algorithmType string,
	playDate *time.Time, // 添加播放日期参数用于时间衰减计算
) float64 {
	score := 0.5 // 基础分数

	// 计算标签匹配度
	matchedTagCount := 0
	if tags, ok := itemDoc["tags"]; ok {
		if tagArray, ok := tags.(primitive.A); ok {
			for _, tag := range tagArray {
				if tagName, ok := tag.(string); ok {
					if count, exists := tagNameToCount[tagName]; exists {
						matchedTagCount++
						// 根据标签的频率调整分数，频率越高分数越高
						score += float64(count) * 0.001
					}
				}
			}
		}
	}

	// 检查genre字段
	if genre, ok := itemDoc["genre"]; ok {
		if genreName, ok := genre.(string); ok {
			if count, exists := tagNameToCount[genreName]; exists {
				matchedTagCount++
				// genre匹配给予更高的权重
				score += float64(count) * 0.002
			}
		}
	}

	// 根据匹配的标签数量调整分数
	if matchedTagCount > 0 {
		score += float64(matchedTagCount) * 0.1
	}

	// 根据算法类型调整分数策略
	switch algorithmType {
	case "personalized":
		score += 0.1 // 个性化推荐额外加分
	case "popular":
		score += 0.05 // 热门推荐稍微加分
	}

	// 添加随机因素
	score += rand.Float64() * 0.1

	// 添加时间衰减因子：近期播放行为权重更高
	if playDate != nil {
		// 计算距离现在的时间差（天数）
		timeDiff := time.Since(*playDate).Hours() / 24.0

		// 使用指数衰减函数：越近的播放行为权重越高
		// 衰减因子：exp(-timeDiff / halfLife)，halfLife为半衰期（天数）
		// 这里设置半衰期为30天，即30天前的播放行为权重为当前的一半
		halfLife := 30.0
		timeDecayFactor := math.Exp(-timeDiff / halfLife)

		// 将时间衰减因子应用到分数上，最多增加30%的权重
		score *= (1.0 + timeDecayFactor*0.3)
	}

	// 确保分数在合理范围内
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// 从注释数据中获取推荐依据
func (r *recommendRepository) getAnnotationBasisFromAnnotations(
	annotations []scene_audio_db_models.AnnotationMetadata,
	limit int,
) []scene_audio_db_models.AnnotationMetadata {
	if len(annotations) == 0 {
		return []scene_audio_db_models.AnnotationMetadata{}
	}

	if len(annotations) > limit {
		return annotations[:limit]
	}

	return annotations
}

// 从词云数据中获取标签依据
func (r *recommendRepository) getTagBasisFromWordCloud(
	wordCloudTags []scene_audio_db_models.WordCloudMetadata,
	limit int,
) []scene_audio_db_models.WordCloudMetadata {
	if len(wordCloudTags) == 0 {
		return []scene_audio_db_models.WordCloudMetadata{}
	}

	if len(wordCloudTags) > limit {
		return wordCloudTags[:limit]
	}

	return wordCloudTags
}

// 从词云数据中获取相关项目信息
func (r *recommendRepository) getRelatedItemsFromWordCloud(
	wordCloudTags []scene_audio_db_models.WordCloudMetadata,
	limit int,
) []scene_audio_db_models.WordCloudRecommendation {
	var relatedItems []scene_audio_db_models.WordCloudRecommendation

	for i, tag := range wordCloudTags {
		if i >= limit {
			break
		}
		relatedItems = append(relatedItems, scene_audio_db_models.WordCloudRecommendation{
			ID:    tag.ID,
			Type:  tag.Type,
			Name:  tag.Name,
			Score: float64(tag.Count) * 0.01,
		})
	}

	return relatedItems
}

// 创建推荐结果
func (r *recommendRepository) createRecommendationResult(
	itemInfo bson.M,
	recommendType string,
	score float64,
	reason string,
	playCount int,
	rating int,
	starred bool,
	algorithm string,
	parameters map[string]string,
	basis []string,
	annotationBasis []scene_audio_db_models.AnnotationMetadata,
	tagBasis []scene_audio_db_models.WordCloudMetadata,
	relatedItems []scene_audio_db_models.WordCloudRecommendation,
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
			Score           float64                                         `json:"score"`
			Reason          string                                          `json:"reason"`
			PlayCount       int                                             `json:"play_count"`
			Rating          int                                             `json:"rating"`
			Starred         bool                                            `json:"starred"`
			Algorithm       string                                          `json:"algorithm"`
			Parameters      map[string]string                               `json:"parameters"`
			Basis           []string                                        `json:"basis"`
			AnnotationBasis []scene_audio_db_models.AnnotationMetadata      `json:"annotation_basis"`
			TagBasis        []scene_audio_db_models.WordCloudMetadata       `json:"tag_basis"`
			RelatedItems    []scene_audio_db_models.WordCloudRecommendation `json:"related_items"`
		}{
			ArtistMetadata:  artist,
			Score:           score,
			Reason:          reason,
			PlayCount:       playCount,
			Rating:          rating,
			Starred:         starred,
			Algorithm:       algorithm,
			Parameters:      parameters,
			Basis:           basis,
			AnnotationBasis: annotationBasis,
			TagBasis:        tagBasis,
			RelatedItems:    relatedItems,
		}
		return result, nil

	case "album":
		var album scene_audio_route_models.AlbumMetadata
		bsonBytes, _ := bson.Marshal(itemInfo)
		bson.Unmarshal(bsonBytes, &album)

		// 添加推荐元数据
		result := struct {
			scene_audio_route_models.AlbumMetadata
			Score           float64                                         `json:"score"`
			Reason          string                                          `json:"reason"`
			PlayCount       int                                             `json:"play_count"`
			Rating          int                                             `json:"rating"`
			Starred         bool                                            `json:"starred"`
			Algorithm       string                                          `json:"algorithm"`
			Parameters      map[string]string                               `json:"parameters"`
			Basis           []string                                        `json:"basis"`
			AnnotationBasis []scene_audio_db_models.AnnotationMetadata      `json:"annotation_basis"`
			TagBasis        []scene_audio_db_models.WordCloudMetadata       `json:"tag_basis"`
			RelatedItems    []scene_audio_db_models.WordCloudRecommendation `json:"related_items"`
		}{
			AlbumMetadata:   album,
			Score:           score,
			Reason:          reason,
			PlayCount:       playCount,
			Rating:          rating,
			Starred:         starred,
			Algorithm:       algorithm,
			Parameters:      parameters,
			Basis:           basis,
			AnnotationBasis: annotationBasis,
			TagBasis:        tagBasis,
			RelatedItems:    relatedItems,
		}
		return result, nil

	case "media":
		var mediaFile scene_audio_route_models.MediaFileMetadata
		bsonBytes, _ := bson.Marshal(itemInfo)
		bson.Unmarshal(bsonBytes, &mediaFile)

		// 添加推荐元数据
		result := struct {
			scene_audio_route_models.MediaFileMetadata
			Score           float64                                         `json:"score"`
			Reason          string                                          `json:"reason"`
			PlayCount       int                                             `json:"play_count"`
			Rating          int                                             `json:"rating"`
			Starred         bool                                            `json:"starred"`
			Algorithm       string                                          `json:"algorithm"`
			Parameters      map[string]string                               `json:"parameters"`
			Basis           []string                                        `json:"basis"`
			AnnotationBasis []scene_audio_db_models.AnnotationMetadata      `json:"annotation_basis"`
			TagBasis        []scene_audio_db_models.WordCloudMetadata       `json:"tag_basis"`
			RelatedItems    []scene_audio_db_models.WordCloudRecommendation `json:"related_items"`
		}{
			MediaFileMetadata: mediaFile,
			Score:             score,
			Reason:            reason,
			PlayCount:         playCount,
			Rating:            rating,
			Starred:           starred,
			Algorithm:         algorithm,
			Parameters:        parameters,
			Basis:             basis,
			AnnotationBasis:   annotationBasis,
			TagBasis:          tagBasis,
			RelatedItems:      relatedItems,
		}
		return result, nil

	case "media_cue":
		var mediaFileCue scene_audio_route_models.MediaFileCueMetadata
		bsonBytes, _ := bson.Marshal(itemInfo)
		bson.Unmarshal(bsonBytes, &mediaFileCue)

		// 添加推荐元数据
		result := struct {
			scene_audio_route_models.MediaFileCueMetadata
			Score           float64                                         `json:"score"`
			Reason          string                                          `json:"reason"`
			PlayCount       int                                             `json:"play_count"`
			Rating          int                                             `json:"rating"`
			Starred         bool                                            `json:"starred"`
			Algorithm       string                                          `json:"algorithm"`
			Parameters      map[string]string                               `json:"parameters"`
			Basis           []string                                        `json:"basis"`
			AnnotationBasis []scene_audio_db_models.AnnotationMetadata      `json:"annotation_basis"`
			TagBasis        []scene_audio_db_models.WordCloudMetadata       `json:"tag_basis"`
			RelatedItems    []scene_audio_db_models.WordCloudRecommendation `json:"related_items"`
		}{
			MediaFileCueMetadata: mediaFileCue,
			Score:                score,
			Reason:               reason,
			PlayCount:            playCount,
			Rating:               rating,
			Starred:              starred,
			Algorithm:            algorithm,
			Parameters:           parameters,
			Basis:                basis,
			AnnotationBasis:      annotationBasis,
			TagBasis:             tagBasis,
			RelatedItems:         relatedItems,
		}
		return result, nil
	}

	return nil, fmt.Errorf("不支持的推荐类型: %s", recommendType)
}

// 当没有注释数据时，直接从目标集合获取数据
func (r *recommendRepository) getItemsWithoutAnnotations(
	ctx context.Context,
	targetCollection string,
	recommendType string,
	startInt int,
	endInt int,
	recommendOffsetInt int,
	randomSeed int64,
) ([]interface{}, error) {
	// 设置随机种子
	rand.Seed(randomSeed)

	// 直接从目标集合获取数据，使用随机推荐
	itemColl := r.db.Collection(targetCollection)

	// 构建聚合管道
	pipeline := []bson.D{
		{{"$sample", bson.D{{"size", endInt * 2}}}},
		{{"$skip", recommendOffsetInt}},
		{{"$limit", endInt - startInt}},
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
		score := 0.3 + rand.Float64()*0.4

		// 创建推荐结果（降级策略中没有播放日期信息，传入nil）
		result, err := r.createRecommendationResult(itemDoc, recommendType, score, "热门推荐", 0, 0, false,
			"FallbackAlgorithm",
			map[string]string{
				"start":  strconv.Itoa(startInt),
				"end":    strconv.Itoa(endInt),
				"offset": strconv.Itoa(recommendOffsetInt),
			},
			[]string{"random"},
			[]scene_audio_db_models.AnnotationMetadata{},
			[]scene_audio_db_models.WordCloudMetadata{},
			[]scene_audio_db_models.WordCloudRecommendation{})

		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// getCachedWordCloudTags 获取带缓存的词云标签
func (r *recommendRepository) getCachedWordCloudTags(ctx context.Context) ([]scene_audio_db_models.WordCloudMetadata, error) {
	// 检查缓存是否有效
	r.cacheMutex.RLock()
	if r.wordCloudCache != nil && time.Now().Before(r.cacheExpiry) {
		// 返回默认类型的词云标签缓存
		if tags, exists := r.wordCloudCache["default"]; exists {
			r.cacheMutex.RUnlock()
			r.logDebug("使用缓存的词云标签，数量: %d", len(tags))
			return tags, nil
		}
	}
	r.cacheMutex.RUnlock()

	// 缓存失效或不存在，重新获取数据
	r.logInfo("词云标签缓存失效或不存在，重新获取数据")
	wordCloudColl := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFileWordCloud)

	// 使用单次查询获取所有词云标签
	wordCloudPipeline := []bson.D{
		{{"$sort", bson.D{{"count", -1}}}},
		{{"$limit", 1000}}, // 获取1000个最常见的词云标签
	}

	wordCloudCursor, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
	if err != nil {
		return nil, fmt.Errorf("获取词云标签失败: %w", err)
	}
	defer wordCloudCursor.Close(ctx)

	var allWordCloudTags []scene_audio_db_models.WordCloudMetadata
	if err = wordCloudCursor.All(ctx, &allWordCloudTags); err != nil {
		return nil, fmt.Errorf("解析词云标签失败: %w", err)
	}

	r.logDebug("成功获取到%d个词云标签", len(allWordCloudTags))

	// 更新缓存
	r.cacheMutex.Lock()
	// 初始化缓存映射（如果为nil）
	if r.wordCloudCache == nil {
		r.wordCloudCache = make(map[string][]scene_audio_db_models.WordCloudMetadata)
	}
	r.wordCloudCache["default"] = allWordCloudTags
	r.cacheExpiry = time.Now().Add(5 * time.Minute) // 缓存5分钟
	r.cacheMutex.Unlock()

	return allWordCloudTags, nil
}

// getCachedWordCloudTagsByType 根据类型获取带缓存的词云标签
func (r *recommendRepository) getCachedWordCloudTagsByType(ctx context.Context, wordCloudType string) ([]scene_audio_db_models.WordCloudMetadata, error) {
	// 检查缓存是否有效
	cacheKey := "type_" + wordCloudType
	r.cacheMutex.RLock()
	if r.wordCloudCache != nil && time.Now().Before(r.cacheExpiry) {
		if tags, exists := r.wordCloudCache[cacheKey]; exists {
			r.cacheMutex.RUnlock()
			r.logDebug("使用缓存的%s类型词云标签，数量: %d", wordCloudType, len(tags))
			return tags, nil
		}
	}
	r.cacheMutex.RUnlock()

	// 缓存失效或不存在，重新获取数据
	r.logInfo("%s类型词云标签缓存失效或不存在，重新获取数据", wordCloudType)
	wordCloudColl := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFileWordCloud)

	// 使用单次查询获取指定类型的词云标签
	wordCloudPipeline := []bson.D{
		{{"$match", bson.D{{"type", wordCloudType}}}},
		{{"$sort", bson.D{{"count", -1}}}},
		{{"$limit", 500}}, // 获取500个最常见的词云标签
	}

	wordCloudCursor, err := wordCloudColl.Aggregate(ctx, wordCloudPipeline)
	if err != nil {
		return nil, fmt.Errorf("获取词云标签失败: %w", err)
	}
	defer wordCloudCursor.Close(ctx)

	var wordCloudTags []scene_audio_db_models.WordCloudMetadata
	if err = wordCloudCursor.All(ctx, &wordCloudTags); err != nil {
		return nil, fmt.Errorf("解析词云标签失败: %w", err)
	}

	r.logDebug("成功获取到%d个%s类型词云标签", len(wordCloudTags), wordCloudType)

	// 更新缓存
	r.cacheMutex.Lock()
	// 初始化缓存映射（如果为nil）
	if r.wordCloudCache == nil {
		r.wordCloudCache = make(map[string][]scene_audio_db_models.WordCloudMetadata)
	}
	r.wordCloudCache[cacheKey] = wordCloudTags
	r.cacheExpiry = time.Now().Add(5 * time.Minute) // 缓存5分钟
	r.cacheMutex.Unlock()

	return wordCloudTags, nil
}

// ExtractTags 从项目中提取标签
func (r *recommendRepository) ExtractTags(ctx context.Context, items []bson.M) ([]string, map[string]int, error) {
	// 步骤3: 从项目中提取标签信息
	var allTagNames []string
	tagSet := make(map[string]bool)        // 用于去重
	tagSourceCount := make(map[string]int) // 统计标签来源

	r.logDebug("开始从%d个项目中提取标签", len(items))

	// 批量获取所有词云标签用于匹配（限制数量以提高性能）
	allWordCloudTags, err := r.getCachedWordCloudTags(ctx)
	if err != nil {
		r.logInfo("获取词云标签失败: %v", err)
		return nil, nil, fmt.Errorf("获取词云标签失败: %w", err)
	}

	r.logDebug("获取到%d个词云标签用于匹配", len(allWordCloudTags))

	// 批量处理项目标签提取
	// 创建一个映射来存储所有项目文本，以便一次性处理
	projectTexts := make([]string, len(items))

	for i, item := range items {
		// 收集项目的所有文本内容用于匹配
		var itemTexts []string
		if album, ok := item["album"]; ok {
			if albumStr, ok := album.(string); ok && albumStr != "" {
				itemTexts = append(itemTexts, strings.ToLower(albumStr))
			}
		}
		if artist, ok := item["artist"]; ok {
			if artistStr, ok := artist.(string); ok && artistStr != "" {
				itemTexts = append(itemTexts, strings.ToLower(artistStr))
			}
		}
		if fileName, ok := item["file_name"]; ok {
			if fileNameStr, ok := fileName.(string); ok && fileNameStr != "" {
				itemTexts = append(itemTexts, strings.ToLower(fileNameStr))
			}
		}
		if lyrics, ok := item["lyrics"]; ok {
			if lyricsStr, ok := lyrics.(string); ok && lyricsStr != "" {
				// 只取歌词的前1000个字符用于匹配，避免过长的文本影响性能
				if len(lyricsStr) > 1000 {
					lyricsStr = lyricsStr[:1000]
				}
				itemTexts = append(itemTexts, strings.ToLower(lyricsStr))
			}
		}

		// 将项目文本连接成一个字符串用于匹配
		fullItemText := strings.ToLower(strings.Join(itemTexts, " "))
		projectTexts[i] = fullItemText

		// 添加调试信息
		if i < 3 {
			r.logDebug("项目%d的文本内容长度: %d", i, len(fullItemText))
		}
	}

	// 批量处理标签匹配
	for i, fullItemText := range projectTexts {
		// 与词云标签进行模糊匹配
		for _, wordCloudTag := range allWordCloudTags {
			tagName := strings.ToLower(wordCloudTag.Name)
			// 使用更智能的匹配算法
			if r.CalculateSimilarity(fullItemText, tagName) {
				// 去重并过滤空标签
				if tagName != "" && !tagSet[tagName] {
					tagSet[tagName] = true
					allTagNames = append(allTagNames, wordCloudTag.Name) // 保持原始大小写
					tagSourceCount["word_cloud"]++
				}
			}
		}

		// 提取genre字段（单独添加）
		if genre, ok := items[i]["genre"]; ok {
			if genreName, ok := genre.(string); ok && genreName != "" {
				// 去重并过滤空标签
				if !tagSet[genreName] {
					tagSet[genreName] = true
					allTagNames = append(allTagNames, genreName)
					tagSourceCount["genre"]++
				}
			}
		}
	}

	// 添加调试信息
	r.logInfo("标签来源统计: word_cloud=%d, genre=%d", tagSourceCount["word_cloud"], tagSourceCount["genre"])
	r.logInfo("从项目中提取了%d个标签", len(allTagNames))

	// 打印前几个提取的标签
	if len(allTagNames) > 0 {
		count := len(allTagNames)
		if count > 5 {
			count = 5
		}
		r.logDebug("前%d个标签: %v", count, allTagNames[:count])
	}

	return allTagNames, tagSourceCount, nil
}

// CalculateSimilarity 智能标签匹配函数
func (r *recommendRepository) CalculateSimilarity(itemText, tagName string) bool {
	// 精确匹配
	if strings.Contains(itemText, tagName) {
		return true
	}

	// 反向匹配
	if strings.Contains(tagName, itemText) {
		return true
	}

	// 分词匹配
	itemWords := strings.Fields(itemText)
	tagWords := strings.Fields(tagName)

	// 检查是否有共同的词
	for _, itemWord := range itemWords {
		for _, tagWord := range tagWords {
			// 如果词长度大于2且包含关系
			if len(itemWord) > 2 && len(tagWord) > 2 &&
				(strings.Contains(itemWord, tagWord) || strings.Contains(tagWord, itemWord)) {
				return true
			}
		}
	}

	return false
}

// handleError 统一错误处理函数
func (r *recommendRepository) handleError(operation string, err error) error {
	if err != nil {
		// 记录错误日志
		fmt.Printf("[%s] 错误: %v\n", operation, err)
		// 返回包装后的错误
		return fmt.Errorf("%s失败: %w", operation, err)
	}
	return nil
}

// logInfo 统一信息日志函数
func (r *recommendRepository) logInfo(message string, args ...interface{}) {
	if r.logShow {
		fmt.Printf("[INFO] "+message+"\n", args...)
	}
}

// RankResults 对推荐结果进行排序
func (r *recommendRepository) RankResults(items []bson.M, tagNameToCount map[string]int, algorithmType string) []bson.M {
	// 这里可以实现更复杂的排序逻辑
	// 目前我们只是返回原始项目列表
	return items
}

// logDebug 统一调试日志函数
func (r *recommendRepository) logDebug(message string, args ...interface{}) {
	// 在生产环境中可能需要根据日志级别来决定是否输出
	if r.logShow {
		fmt.Printf("[DEBUG] "+message+"\n", args...)
	}
}

// 确保recommendRepository实现所有接口
var _ DataFetcher = (*recommendRepository)(nil)
var _ TagExtractor = (*recommendRepository)(nil)
var _ SimilarityCalculator = (*recommendRepository)(nil)
var _ ResultRanker = (*recommendRepository)(nil)

// GetAnnotations 根据算法类型获取注释数据
func (r *recommendRepository) GetAnnotations(
	ctx context.Context,
	itemType string,
	algorithmType string,
) ([]scene_audio_db_models.AnnotationMetadata, error) {
	annotationColl := r.db.Collection(domain.CollectionFileEntityAudioSceneAnnotation)

	// 记录开始处理的日志
	r.logDebug("开始获取%s类型的注释数据，算法类型: %s", itemType, algorithmType)

	// 构建查询条件，根据算法类型调整策略
	var matchCondition bson.D
	switch algorithmType {
	case "personalized":
		matchCondition = bson.D{{"item_type", itemType}}
	case "popular":
		matchCondition = bson.D{
			{"item_type", itemType},
			{"play_count", bson.D{{"$gt", 0}}},
		}
	default:
		matchCondition = bson.D{{"item_type", itemType}}
	}

	// 获取用户行为数据
	// 根据算法类型构建不同的聚合管道
	var annotationPipeline []bson.D

	// 首先检查播放次数为30次以上的数据数量，以动态调整limit值
	countMatchCondition := bson.D{
		{"$and", []bson.D{
			matchCondition,
			{{"play_count", bson.D{{"$gte", 30}}}},
		}},
	}

	// 计算播放次数>=30的文档数量
	count, err := annotationColl.CountDocuments(ctx, countMatchCondition)
	if err != nil {
		return nil, fmt.Errorf("计算播放次数>=30的数据失败: %w", err)
	}

	// 根据播放次数>=30的数据数量动态设置limit值
	limitValue := 200
	if count > 200 {
		limitValue = int(count)
	}

	r.logDebug("根据播放次数动态调整limit值: %d (原始计数: %d)", limitValue, count)

	switch algorithmType {
	case "personalized":
		// 个性化推荐：根据播放次数、最近播放时间、喜欢状态、收藏星级综合排序
		annotationPipeline = []bson.D{
			{{"$match", matchCondition}},
			{{"$addFields", bson.D{
				{"score", bson.D{
					{"$add", []interface{}{
						// 播放次数占25%
						bson.D{{"$multiply", []interface{}{"$play_count", 0.25}}},
						// 评分占20%
						bson.D{{"$multiply", []interface{}{"$rating", 0.2}}},
						// 收藏状态占25%
						bson.D{{"$cond", []interface{}{"$starred", 0.25, 0}}},
						// 完整播放次数占10%
						bson.D{{"$cond", []interface{}{
							bson.D{{"$gte", []interface{}{"$play_complete_count", 1}}},
							0.1,
							0,
						}}},
						// 最近播放时间占20% (越近分数越高)
						bson.D{{"$multiply", []interface{}{
							bson.D{{"$divide", []interface{}{
								bson.D{{"$subtract", []interface{}{"$$NOW", "$play_date"}}},
								1000 * 60 * 60 * 24, // 转换为天数
							}}},
							0.2, // 越近分数越高，权重提高到20%
						}}},
					}},
				}},
			}}},
			{{"$sort", bson.D{{"score", -1}}}},
			{{"$limit", limitValue}}, // 根据播放次数>=30的数据数量动态设置limit值
		}
	case "popular":
		// 热门推荐：主要根据播放次数排序
		annotationPipeline = []bson.D{
			{{"$match", matchCondition}},
			{{"$sort", bson.D{{"play_count", -1}}}},
			{{"$limit", limitValue}}, // 根据播放次数>=30的数据数量动态设置limit值
		}
	default:
		// 通用推荐：结合随机性和用户行为数据
		// 30%随机性 + 70%基于用户行为的智能选取
		annotationPipeline = []bson.D{
			{{"$match", matchCondition}},
			{{"$addFields", bson.D{
				{"random_value", bson.D{{"$rand", bson.A{}}}}, // 添加随机值字段
			}}},
			{{"$addFields", bson.D{
				{"score", bson.D{
					{"$add", []interface{}{
						// 随机因素占30%
						bson.D{{"$multiply", []interface{}{"$random_value", 0.3}}},
						// 播放次数占25%
						bson.D{{"$multiply", []interface{}{"$play_count", 0.25}}},
						// 最近播放时间占20% (越近分数越高)
						bson.D{{"$multiply", []interface{}{
							bson.D{{"$divide", []interface{}{
								bson.D{{"$subtract", []interface{}{"$$NOW", "$play_date"}}},
								1000 * 60 * 60 * 24, // 转换为天数
							}}},
							0.2, // 越近分数越高
						}}},
						// 喜欢状态占15%
						bson.D{{"$cond", []interface{}{"$starred", 0.15, 0}}},
						// 收藏星级占10%
						bson.D{{"$multiply", []interface{}{"$rating", 0.1}}},
					}},
				}},
			}}},
			{{"$sort", bson.D{{"score", -1}}}},
			{{"$limit", limitValue}}, // 根据播放次数>=30的数据数量动态设置limit值
		}
	}

	annotationCursor, err := annotationColl.Aggregate(ctx, annotationPipeline)
	if err := r.handleError("注释数据查询", err); err != nil {
		return nil, err
	}
	defer annotationCursor.Close(ctx)

	var annotations []scene_audio_db_models.AnnotationMetadata
	if err := annotationCursor.All(ctx, &annotations); err != nil {
		if err := r.handleError("解析注释数据", err); err != nil {
			return nil, err
		}
	}

	r.logDebug("成功获取到%d个注释数据", len(annotations))
	return annotations, nil
}

// GetItems 根据注释数据获取对应的项目
func (r *recommendRepository) GetItems(
	ctx context.Context,
	targetCollection string,
	annotations []scene_audio_db_models.AnnotationMetadata,
) ([]bson.M, error) {
	// 记录开始处理的日志
	r.logDebug("开始处理%d个注释数据", len(annotations))

	// 收集item_id用于后续查询
	var itemIds []string
	itemIdToAnnotation := make(map[string]scene_audio_db_models.AnnotationMetadata)
	for _, annotation := range annotations {
		itemIds = append(itemIds, annotation.ItemID)
		itemIdToAnnotation[annotation.ItemID] = annotation
	}

	// 将字符串类型的item_id转换为ObjectID
	var itemObjectIds []primitive.ObjectID
	var invalidItemIds []string
	for _, itemId := range itemIds {
		if objectId, err := primitive.ObjectIDFromHex(itemId); err == nil {
			itemObjectIds = append(itemObjectIds, objectId)
		} else {
			invalidItemIds = append(invalidItemIds, itemId)
		}
	}

	// 输出无效item_id的调试信息
	if len(invalidItemIds) > 0 {
		endIdx := len(invalidItemIds)
		if endIdx > 5 {
			endIdx = 5
		}
		r.logInfo("发现%d个无效的item_id: %v", len(invalidItemIds), invalidItemIds[:endIdx])
	}

	// 查询对应的项目信息
	itemColl := r.db.Collection(targetCollection)
	r.logDebug("准备从集合%s查询%d个项目", targetCollection, len(itemObjectIds))

	itemPipeline := []bson.D{
		// 匹配项目ID
		{{"$match", bson.D{{"_id", bson.D{{"$in", itemObjectIds}}}}}},
		// 添加随机排序
		{{"$sample", bson.D{{"size", len(itemObjectIds)}}}},
		// 限制数量
		{{"$limit", 200}}, // 增加到200个以获取更多样化的数据
	}

	itemCursor, err := itemColl.Aggregate(ctx, itemPipeline)
	if err := r.handleError("项目数据查询", err); err != nil {
		return nil, err
	}
	defer itemCursor.Close(ctx)

	var items []bson.M
	if err := itemCursor.All(ctx, &items); err != nil {
		if err := r.handleError("解析项目数据", err); err != nil {
			return nil, err
		}
	}

	r.logDebug("成功获取到%d个项目数据", len(items))
	return items, nil
}
