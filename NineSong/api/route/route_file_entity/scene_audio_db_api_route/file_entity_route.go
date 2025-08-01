package scene_audio_db_api_route

import (
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_db_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_db_repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity"
	"github.com/gin-gonic/gin"
	"time"
)

func NewFileEntityRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	// 初始化仓库
	fileRepo := repository_file_entity.NewFileRepo(db, domain.CollectionFileEntityFileInfo)
	folderRepo := repository_file_entity.NewFolderRepo(db, domain.CollectionFileEntityFolderInfo)
	detector := &domain_file_entity.FileDetectorImpl{}

	artistRepo := scene_audio_db_repository.NewArtistRepository(db, domain.CollectionFileEntityAudioSceneArtist)
	albumRepo := scene_audio_db_repository.NewAlbumRepository(db, domain.CollectionFileEntityAudioSceneAlbum)
	mediaRepo := scene_audio_db_repository.NewMediaFileRepository(db, domain.CollectionFileEntityAudioSceneMediaFile)
	tempRepo := scene_audio_db_repository.NewTempRepository(db, domain.CollectionFileEntityAudioSceneTempMetadata)
	mediaCueRepo := scene_audio_db_repository.NewMediaFileCueRepository(db, domain.CollectionFileEntityAudioSceneMediaFileCue)
	mediaWordCloudRepo := scene_audio_db_repository.NewWordCloudRepository(db, domain.CollectionFileEntityAudioSceneMediaFileWordCloud)
	lyricsFileRepo := scene_audio_db_repository.NewLyricsFileRepository(db, domain.CollectionFileEntityAudioSceneLyricsFile)
	// 构建用例（新增超时参数）
	uc := usecase_file_entity.NewFileUsecase(
		db,
		fileRepo,
		folderRepo,
		detector,
		0,
		artistRepo,
		albumRepo,
		mediaRepo,
		tempRepo,
		mediaCueRepo,
		mediaWordCloudRepo,
		lyricsFileRepo,
	)

	// 注册控制器
	ctrl := scene_audio_db_api_controller.NewFileController(uc)

	// 路由配置
	group.Use(requestLogger())
	group.POST("/scan", ctrl.ScanDirectory)
	group.GET("/scan_progress", ctrl.GetScanProgress)
}

func requestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d\n",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
		)
	})
}
