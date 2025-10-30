package scene_audio_route_api_route

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_db_repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_route_repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/gin-gonic/gin"
)

func NewRecommendRouter(
	timeout time.Duration,
	db mongo.Database,
	group *gin.RouterGroup,
) {
	// 初始化repository
	annotationRepo := scene_audio_route_repository.NewAnnotationRepository(db)
	wordCloudRepo := scene_audio_db_repository.NewWordCloudRepository(db, domain.CollectionFileEntityAudioSceneMediaFileWordCloud)
	recommendRepo := scene_audio_db_repository.NewRecommendRepository(db, domain.CollectionFileEntityAudioSceneAnnotation)

	// 初始化usecase
	recommendUsecase := scene_audio_route_usecase.NewRecommendUsecase(annotationRepo, wordCloudRepo, recommendRepo, db, timeout)

	// 初始化controller
	recommendCtrl := scene_audio_route_api_controller.NewRecommendController(recommendUsecase)

	// 注册路由
	recommendGroup := group.Group("/recommend")
	{
		// 基于注释和词云的推荐
		// GET /recommend/annotation_wordcloud?start=0&end=30&recommend_type=media&random_seed=666&recommend_offset=0
		recommendGroup.GET("/annotation_wordcloud", recommendCtrl.GetRecommendAnnotationWordCloudItems)

		// 个性化推荐
		// GET /recommend/personalized?recommend_type=media&limit=30[&user_id=xxx]
		recommendGroup.GET("/personalized", recommendCtrl.GetPersonalizedRecommendations)

		// 热门推荐
		// GET /recommend/popular?recommend_type=media&limit=30
		recommendGroup.GET("/popular", recommendCtrl.GetPopularRecommendations)
	}
}
