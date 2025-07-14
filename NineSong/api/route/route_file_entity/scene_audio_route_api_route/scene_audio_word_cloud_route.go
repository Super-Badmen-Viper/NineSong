package scene_audio_route_api_route

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_db_repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase"
	"github.com/gin-gonic/gin"
	"time"
)

func NewWordCloudRouter(
	timeout time.Duration,
	db mongo.Database,
	group *gin.RouterGroup,
) {
	repoMediaFileDB := scene_audio_db_repository.NewMediaFileRepository(db, domain.CollectionFileEntityAudioSceneMediaFile)
	repoMediaFileWordCloud := scene_audio_db_repository.NewWordCloudRepository(db, domain.CollectionFileEntityAudioSceneMediaFileWordCloud)
	usecase := scene_audio_route_usecase.NewWordCloudUsecase(repoMediaFileDB, repoMediaFileWordCloud, timeout)
	ctrl := scene_audio_route_api_controller.NewWordCloudController(usecase)

	wordCloudGroup := group.Group("/word_cloud")
	{
		wordCloudGroup.GET("", ctrl.GetAllWordCloudHandler)
		wordCloudGroup.GET("genre", ctrl.GetAllGenreHandler)
		wordCloudGroup.GET("high", ctrl.GetHighFrequencyWordCloudHandler)
		wordCloudGroup.POST("recommend", ctrl.GetRecommendedWordCloudHandler)
	}
}
