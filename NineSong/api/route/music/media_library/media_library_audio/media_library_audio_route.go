package media_library_audio

import (
	"time"

	controller "github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/music/media_library/media_library_audio"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	dbMongo "github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo"
	repoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/media_library/media_library_audio/core"
	syncRecordRepoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/media_library/sync_record/core"
	fileUsecaseCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/music/file_entity/file/core"
	usecaseCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/music/media_library/media_library_audio/core"
	syncRecordUsecaseCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/music/media_library/sync_record/core"

	"github.com/gin-gonic/gin"
)

func NewMediaLibraryAudioRouter(
	timeout time.Duration,
	db dbMongo.Database,
	protectedRouter *gin.RouterGroup,
	basicFileUsecase *fileUsecaseCore.FileUsecase,
) {
	// 创建仓库实例
	repo := repoCore.NewMediaLibraryAudioRepository(db, shared.CollectionSceneAudioDbMediaLibraryAudio)
	chunkedUploadSessionRepo := repoCore.NewChunkedUploadSessionRepository(db, shared.CollectionChunkedUploadSession)

	// 创建用例实例
	usecase := usecaseCore.NewMediaLibraryAudioUsecase(repo, chunkedUploadSessionRepo, basicFileUsecase, timeout)

	// 创建同步记录仓库和用例实例
	syncRecordRepo := syncRecordRepoCore.NewMediaLibrarySyncRecordRepository(db, shared.CollectionMediaLibrarySyncRecord)
	syncRecordUsecase := syncRecordUsecaseCore.NewMediaLibrarySyncRecordUseCase(syncRecordRepo, timeout)

	// 创建控制器实例
	ctrl := controller.NewMediaLibraryAudioController(usecase, syncRecordUsecase)

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
