package scene_audio_route_api_route

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_route_repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/gin-gonic/gin"
)

func NewArtistRouter(
	timeout time.Duration,
	db mongo.Database,
	group *gin.RouterGroup,
) {
	repo := scene_audio_route_repository.NewArtistRepository(db, domain.CollectionFileEntityAudioSceneArtist)

	usecase := scene_audio_route_usecase.NewArtistUsecase(repo, timeout)
	ctrl := scene_audio_route_api_controller.NewArtistController(usecase)

	artistGroup := group.Group("/artists")
	{
		artistGroup.GET("", ctrl.GetArtists)
		artistGroup.GET("/metadatas", ctrl.GetArtistMetadatas)
		artistGroup.GET("/sort", ctrl.GetArtistsMultipleSorting)
		artistGroup.GET("/filter_counts", ctrl.GetArtistFilterCounts)
		artistGroup.GET("/tree", ctrl.GetArtistTrees)
	}
}
