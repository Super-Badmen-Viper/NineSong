package scene_audio_route_usecase

import (
	"context"
	"fmt"

	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
)

type RecommendUsecase struct {
	repoAnnotation         scene_audio_route_interface.AnnotationRepository
	repoWordCloud          scene_audio_db_interface.WordCloudDBRepository
	repoMediaFileRecommend scene_audio_route_interface.RecommendRouteRepository
	db                     mongo.Database
	timeout                time.Duration
}

func NewRecommendUsecase(
	repoAnnotation scene_audio_route_interface.AnnotationRepository,
	repoWordCloud scene_audio_db_interface.WordCloudDBRepository,
	repoMediaFileRecommend scene_audio_route_interface.RecommendRouteRepository,
	db mongo.Database,
	timeout time.Duration,
) scene_audio_route_interface.RecommendRouteRepository {
	return &RecommendUsecase{
		repoAnnotation:         repoAnnotation,
		repoWordCloud:          repoWordCloud,
		repoMediaFileRecommend: repoMediaFileRecommend,
		db:                     db,
		timeout:                timeout,
	}
}

func (uc *RecommendUsecase) GetGeneralRecommendations(
	ctx context.Context,
	recommendType string,
	limit int,
	randomSeed string,
	recommendOffset string,
) ([]interface{}, error) {
	// 验证参数
	if recommendType == "" || randomSeed == "" || recommendOffset == "" {
		return nil, fmt.Errorf("所有参数都是必需的")
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit参数必须大于0")
	}

	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	results, err := uc.repoMediaFileRecommend.GetGeneralRecommendations(
		ctx,
		recommendType,
		limit,
		randomSeed,
		recommendOffset,
	)

	if err != nil {
		return nil, fmt.Errorf("获取推荐数据失败: %w", err)
	}

	// 检查结果是否为空
	if results == nil || len(results) == 0 {
		return nil, fmt.Errorf("未找到推荐数据")
	}

	return results, nil
}

func (uc *RecommendUsecase) GetPersonalizedRecommendations(
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

	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	results, err := uc.repoMediaFileRecommend.GetPersonalizedRecommendations(
		ctx,
		userId,
		recommendType,
		limit,
	)

	if err != nil {
		return nil, fmt.Errorf("获取个性化推荐数据失败: %w", err)
	}

	// 检查结果是否为空
	if results == nil || len(results) == 0 {
		return nil, fmt.Errorf("未找到个性化推荐数据")
	}

	return results, nil
}

func (uc *RecommendUsecase) GetPopularRecommendations(
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

	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	results, err := uc.repoMediaFileRecommend.GetPopularRecommendations(
		ctx,
		recommendType,
		limit,
	)

	if err != nil {
		return nil, fmt.Errorf("获取热门推荐数据失败: %w", err)
	}

	// 检查结果是否为空
	if results == nil || len(results) == 0 {
		return nil, fmt.Errorf("未找到热门推荐数据")
	}

	return results, nil
}
