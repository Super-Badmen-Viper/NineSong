package scene_audio_db_api_controller

import (
	"errors"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type LibraryController struct {
	uc usecase_file_entity.LibraryUsecase
}

func NewLibraryController(uc usecase_file_entity.LibraryUsecase) *LibraryController {
	return &LibraryController{uc: uc}
}

func (ctrl *LibraryController) BrowseFolders(c *gin.Context) {
	path := c.Query("path")

	folders, err := ctrl.uc.BrowseFolders(c.Request.Context(), path)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "FOLDER_BROWSE_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(c, "folders", folders, len(folders))
}

func (ctrl *LibraryController) CreateLibrary(c *gin.Context) {
	var req struct {
		Name       string `form:"name" binding:"required"`
		Path       string `form:"path" binding:"required"`
		FolderType int    `form:"folder_type" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.FolderType < 1 {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_TYPE", "无效的媒体库类型")
	}

	library, err := ctrl.uc.CreateLibrary(c.Request.Context(), req.Name, req.Path, req.FolderType)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain_file_entity.ErrLibraryDuplicate) {
			status = http.StatusConflict
		}
		controller.ErrorResponse(c, status, "LIBRARY_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"library": library,
	})
}

func (ctrl *LibraryController) UpdateLibrary(c *gin.Context) {
	var params struct {
		ID            string `form:"id" binding:"required"`
		NewName       string `form:"newName" binding:"required"`
		NewFolderPath string `form:"newFolderPath" binding:"required"`
	}

	// 使用ShouldBind处理表单数据[1,7](@ref)
	if err := c.ShouldBind(&params); err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// 转换ID格式
	objID, err := primitive.ObjectIDFromHex(params.ID)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "无效的媒体库ID格式")
		return
	}

	// 调用业务层执行更新
	if err := ctrl.uc.UpdateLibrary(c.Request.Context(), objID, params.NewName, params.NewFolderPath); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain_file_entity.ErrLibraryNotFound):
			status = http.StatusNotFound
		case errors.Is(err, domain_file_entity.ErrLibraryDuplicate):
			status = http.StatusConflict
		case errors.Is(err, domain_file_entity.ErrLibraryInUse):
			status = http.StatusLocked // 423 Locked更准确
		}
		controller.ErrorResponse(c, status, "LIBRARY_ERROR", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (ctrl *LibraryController) DeleteLibrary(c *gin.Context) {
	id := c.Query("id")

	if err := ctrl.uc.DeleteLibrary(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain_file_entity.ErrLibraryNotFound):
			status = http.StatusNotFound
		case errors.Is(err, domain_file_entity.ErrLibraryInUse):
			status = http.StatusConflict
		}
		controller.ErrorResponse(c, status, "LIBRARY_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted":   true,
		"libraryId": id,
	})
}

func (ctrl *LibraryController) GetLibraries(c *gin.Context) {
	libraries, err := ctrl.uc.GetLibraries(c.Request.Context())
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "LIBRARY_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(c, "libraries", libraries, len(libraries))
}
