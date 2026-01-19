package scene_audio_route_api_route

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	dbMongo "github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_db_repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase"

	"github.com/gin-gonic/gin"
)

func NewMediaLibraryAudioRouter(
	timeout time.Duration,
	db dbMongo.Database,
	protectedRouter *gin.RouterGroup,
	basicFileUsecase *usecase_file_entity.FileUsecase,
) {
	// 创建仓库实例
	repo := scene_audio_db_repository.NewMediaLibraryAudioRepository(db, domain.CollectionSceneAudioDbMediaLibraryAudio)
	chunkedUploadSessionRepo := scene_audio_db_repository.NewChunkedUploadSessionRepository(db, domain.CollectionChunkedUploadSession)

	// 创建用例实例
	usecase := scene_audio_route_usecase.NewMediaLibraryAudioUsecase(repo, chunkedUploadSessionRepo, basicFileUsecase, timeout)

	// 创建同步记录仓库和用例实例
	syncRecordRepo := scene_audio_db_repository.NewMediaLibrarySyncRecordRepository(db, domain.CollectionMediaLibrarySyncRecord)
	syncRecordUsecase := scene_audio_route_usecase.NewMediaLibrarySyncRecordUseCase(syncRecordRepo, timeout)

	// 创建控制器实例
	ctrl := scene_audio_route_api_controller.NewMediaLibraryAudioController(usecase, syncRecordUsecase)

	// 定义路由组
	audioGroup := protectedRouter.Group("/media-library-audio")
	{
		// 上传音频文件
		audioGroup.POST("/upload", ctrl.UploadHandler)

		// 下载音频文件
		audioGroup.GET("/download/:file_id", ctrl.DownloadHandler)

		// 获取音频文件信息
		audioGroup.GET("/info/:file_id", ctrl.GetFileInfoHandler)

		// 获取指定媒体库的所有音频文件
		audioGroup.GET("/library/:library_id", ctrl.GetFilesByLibraryHandler)

		// 获取指定用户上传的所有音频文件
		audioGroup.GET("/uploader/:uploader_id", ctrl.GetFilesByUploaderHandler)

		// 删除音频文件
		audioGroup.DELETE("/:file_id", ctrl.DeleteHandler)

		// 获取上传进度
		audioGroup.GET("/progress/upload/:upload_id", ctrl.GetUploadProgressHandler)

		// 获取下载进度
		audioGroup.GET("/progress/download/:file_id", ctrl.GetDownloadProgressHandler)
	}
}
