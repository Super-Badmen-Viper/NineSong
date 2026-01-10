package scene_audio_db_api_route

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_db_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity"
	"github.com/gin-gonic/gin"
)

func NewFolderEntityRouter(timeout time.Duration, db mongo.Database, group *gin.RouterGroup, fileUsecase *usecase_file_entity.FileUsecase) {
	folderRepo := repository_file_entity.NewFolderRepo(db, domain.CollectionFileEntityFolderInfo)

	libraryUsecase := usecase_file_entity.NewLibraryUsecase(folderRepo)
	libraryUsecase.SetFileUsecase(fileUsecase)

	libCtrl := scene_audio_db_api_controller.NewLibraryController(libraryUsecase)

	group.GET("/folders", libCtrl.BrowseFolders)
	group.POST("/libraries", libCtrl.CreateLibrary)
	group.PUT("/libraries", libCtrl.UpdateLibrary)
	group.DELETE("/libraries", libCtrl.DeleteLibrary)
	group.GET("/libraries", libCtrl.GetLibraries)
}
