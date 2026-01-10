package scene_audio_route_api_route

//package scene_audio_route_api_route
//
//import (
//	"time"
//
//	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller"
//	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
//	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_route_repository"
//	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase"
//	"github.com/gin-gonic/gin"
//)
//
//func NewRetrievalRouter(
//	timeout time.Duration,
//	db mongo.Database,
//	group *gin.RouterGroup,
//) {
//	repo := scene_audio_route_repository.NewRetrievalRepository(db)
//	uc := scene_audio_route_usecase.NewRetrievalUsecase(repo, timeout)
//	ctrl := scene_audio_route_api_controller.NewRetrievalController(uc)
//
//	retrievalGroup := group.Group("/media")
//	{
//		retrievalGroup.GET("/stream", ctrl.FixedStreamHandler)
//		retrievalGroup.GET("/stream/real", ctrl.RealStreamHandler)
//		retrievalGroup.GET("/download", ctrl.DownloadHandler)
//		retrievalGroup.GET("/cover", ctrl.CoverArtIDHandler)
//		retrievalGroup.GET("/cover/path", ctrl.CoverArtPathHandler)
//		retrievalGroup.GET("/lyrics", ctrl.LyricsHandlerMetadata)
//		retrievalGroup.GET("/decode_info", ctrl.GetDecodeInfoHandler) // 添加解码信息路由
//	}
//}
