package scene_audio_route_usecase

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"sort"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
)

type wordCloudUsecase struct {
	repoMediaFile          scene_audio_db_interface.MediaFileRepository
	repoMediaFileWordCloud scene_audio_db_interface.WordCloudRepository
	timeout                time.Duration
}

func NewWordCloudUsecase(
	repoMediaFile scene_audio_db_interface.MediaFileRepository,
	repoMediaFileWordCloud scene_audio_db_interface.WordCloudRepository,
	timeout time.Duration,
) scene_audio_route_interface.WordCloudMetadata {
	return &wordCloudUsecase{
		repoMediaFile:          repoMediaFile,
		repoMediaFileWordCloud: repoMediaFileWordCloud,
		timeout:                timeout,
	}
}

func (w *wordCloudUsecase) GetAllWordCloudSearch(
	ctx context.Context,
) ([]scene_audio_db_models.WordCloudMetadata, error) {
	// 1. 获取所有词云数据（指针形式）
	ptrResults, err := w.repoMediaFileWordCloud.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("词云数据查询失败: %w", err)
	}

	// 2. 空数据检查
	if len(ptrResults) == 0 {
		return []scene_audio_db_models.WordCloudMetadata{}, nil
	}

	// 3. 转换为值切片并排序
	results := make([]scene_audio_db_models.WordCloudMetadata, len(ptrResults))
	for i, ptr := range ptrResults {
		if ptr == nil {
			// 空指针保护
			continue
		}
		results[i] = *ptr
	}

	// 4. 按词频降序排序（高频词优先）
	sort.Slice(results, func(i, j int) bool {
		if results[i].Count != results[j].Count {
			return results[i].Count > results[j].Count // 词频高的在前
		}
		return results[i].Name < results[j].Name // 词频相同时按字母排序
	})

	// 5. 添加实时排名（基于当前排序）
	for i := range results {
		results[i].Rank = i + 1
	}

	return results, nil
}

func (w *wordCloudUsecase) GetAllGenreSearch(ctx context.Context) ([]scene_audio_db_models.WordCloudMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	return w.repoMediaFile.GetAllGenre(ctx)
}

func (w *wordCloudUsecase) GetHighFrequencyWordCloudSearch(
	ctx context.Context, wordLimit int,
) ([]scene_audio_db_models.WordCloudMetadata, error) {
	// 1. 删除现有索引
	if err := w.repoMediaFileWordCloud.DropAllIndex(ctx); err != nil {
		// 非致命错误，记录日志但继续执行
		log.Printf("索引删除失败: %v", err)
	}

	// 2. 清空当前词云数据
	if err := w.repoMediaFileWordCloud.AllDelete(ctx); err != nil {
		return nil, fmt.Errorf("词云数据清空失败: %w", err)
	}

	// 3. 获取原始高频词数据
	allWords, err := w.repoMediaFile.GetHighFrequencyWords(ctx, wordLimit)
	if err != nil {
		return nil, fmt.Errorf("高频词获取失败: %w", err)
	}

	// 4. 数据处理流水线（排序+截取TopN）
	sort.Slice(allWords, func(i, j int) bool {
		return allWords[i].Count > allWords[j].Count
	})

	topN := 50
	if len(allWords) < topN {
		topN = len(allWords)
	}
	topWords := allWords[:topN]

	// 5. 构建最终结果集（使用指针切片）
	results := make([]*scene_audio_db_models.WordCloudMetadata, 0, topN)
	for i, word := range topWords {
		results = append(results, &scene_audio_db_models.WordCloudMetadata{
			ID:    primitive.NewObjectID(),
			Type:  word.Type,
			Name:  word.Name,
			Count: word.Count,
			Rank:  i + 1,
		})
	}

	// 6. 批量保存（直接传递指针切片）
	if _, err := w.repoMediaFileWordCloud.BulkUpsert(ctx, results); err != nil {
		return nil, fmt.Errorf("词云数据保存失败: %w", err)
	}

	// 7. 重建索引（补充unique参数）
	if err := w.repoMediaFileWordCloud.CreateIndex(ctx, "name", true); err != nil {
		log.Printf("索引重建警告: %v", err)
	}

	// 转换为值切片返回（适配函数签名）
	finalResults := make([]scene_audio_db_models.WordCloudMetadata, len(results))
	for i, ptr := range results {
		finalResults[i] = *ptr
	}
	return finalResults, nil
}

func (w *wordCloudUsecase) GetRecommendedWordCloudSearch(
	ctx context.Context, keywords []string,
) ([]scene_audio_db_models.Recommendation, error) {
	const recommendLimit = 30
	allRecommendations, err := w.repoMediaFile.GetRecommendedByKeywords(ctx, keywords, recommendLimit)
	if err != nil {
		return nil, fmt.Errorf("单曲推荐获取失败: %w", err)
	}

	// 合并并去重结果
	uniqueRecs := make(map[string]scene_audio_db_models.Recommendation)
	for _, rec := range allRecommendations {
		key := rec.Type + ":" + rec.ID.Hex()
		if existing, exists := uniqueRecs[key]; !exists || existing.Score < rec.Score {
			uniqueRecs[key] = rec
		}
	}

	// 按相关性分数排序
	finalResults := make([]scene_audio_db_models.Recommendation, 0, len(uniqueRecs))
	for _, rec := range uniqueRecs {
		finalResults = append(finalResults, rec)
	}
	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].Score > finalResults[j].Score
	})

	return finalResults, nil
}
