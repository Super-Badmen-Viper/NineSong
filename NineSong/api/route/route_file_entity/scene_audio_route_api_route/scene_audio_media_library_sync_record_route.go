package scene_audio_route_api_route

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	dbMongo "github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	scene_audio_repository "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_db_repository"
	scene_audio_usecase "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase"
	"github.com/gin-gonic/gin"
)

// NewMediaLibrarySyncRecordRouter 创建媒体库同步记录路由
func NewMediaLibrarySyncRecordRouter(timeout time.Duration, db dbMongo.Database, protectedRouter *gin.RouterGroup) {
	// 初始化仓库
	mediaLibrarySyncRecordRepo := scene_audio_repository.NewMediaLibrarySyncRecordRepository(db, domain.CollectionMediaLibrarySyncRecord)

	// 初始化UseCase
	mediaLibrarySyncRecordUseCase := scene_audio_usecase.NewMediaLibrarySyncRecordUseCase(
		mediaLibrarySyncRecordRepo,
		timeout,
	)

	// 创建控制器
	ctrl := scene_audio_route_api_controller.NewMediaLibrarySyncRecordController(mediaLibrarySyncRecordUseCase)

	// 设置路由
	audioGroup := protectedRouter.Group("/media-library-sync-record")
	{
		// 创建同步记录
		audioGroup.POST("/create", func(c *gin.Context) {
			ctrl.CreateSyncRecordAdapter(c)
		})

		// 根据ID获取同步记录
		audioGroup.GET("/id/:id", func(c *gin.Context) {
			ctrl.GetSyncRecordByIDAdapter(c)
		})

		// 根据媒体文件ID获取同步记录
		audioGroup.GET("/media-file-id/:media_file_id", func(c *gin.Context) {
			ctrl.GetSyncRecordByMediaFileIDAdapter(c)
		})

		// 根据媒体库音频文件ID获取同步记录
		audioGroup.GET("/media-library-audio-id/:media_library_audio_id", func(c *gin.Context) {
			ctrl.GetSyncRecordByMediaLibraryAudioIDAdapter(c)
		})

		// 根据同步状态获取同步记录列表
		audioGroup.GET("/status/:status", func(c *gin.Context) {
			ctrl.GetSyncRecordsByStatusAdapter(c)
		})

		// 根据同步类型获取同步记录列表
		audioGroup.GET("/type/:type", func(c *gin.Context) {
			ctrl.GetSyncRecordsByTypeAdapter(c)
		})

		// 更新同步状态
		audioGroup.PUT("/status/:id", func(c *gin.Context) {
			ctrl.UpdateSyncStatusAdapter(c)
		})

		// 更新同步进度
		audioGroup.PUT("/progress/:id", func(c *gin.Context) {
			ctrl.UpdateSyncProgressAdapter(c)
		})

		// 更新同步结果
		audioGroup.PUT("/result/:id", func(c *gin.Context) {
			ctrl.UpdateSyncResultAdapter(c)
		})

		// 根据ID删除同步记录
		audioGroup.DELETE("/id/:id", func(c *gin.Context) {
			ctrl.DeleteSyncRecordByIDAdapter(c)
		})

		// 根据媒体文件ID删除同步记录
		audioGroup.DELETE("/media-file-id/:media_file_id", func(c *gin.Context) {
			ctrl.DeleteSyncRecordByMediaFileIDAdapter(c)
		})

		// 根据媒体库音频文件ID删除同步记录
		audioGroup.DELETE("/media-library-audio-id/:media_library_audio_id", func(c *gin.Context) {
			ctrl.DeleteSyncRecordByMediaLibraryAudioIDAdapter(c)
		})

		// 获取同步统计信息
		audioGroup.GET("/statistics", func(c *gin.Context) {
			ctrl.GetSyncStatisticsAdapter(c)
		})

		// 检查是否已存在关联指定媒体文件ID的同步记录
		audioGroup.GET("/exists/media-file-id/:media_file_id", func(c *gin.Context) {
			ctrl.ExistsByMediaFileIDAdapter(c)
		})

		// 检查是否已存在关联指定媒体库音频文件ID的同步记录
		audioGroup.GET("/exists/media-library-audio-id/:media_library_audio_id", func(c *gin.Context) {
			ctrl.ExistsByMediaLibraryAudioIDAdapter(c)
		})
	}
}