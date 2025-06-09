package scene_audio_db_api_controller

import (
	"errors"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LibraryController struct {
	uc usecase_file_entity.LibraryUsecase
}

func NewLibraryController(uc usecase_file_entity.LibraryUsecase) *LibraryController {
	return &LibraryController{uc: uc}
}

func (ctrl *LibraryController) CreateLibrary(c *gin.Context) {
	var req struct {
		Name      string                          `json:"name" binding:"required"`
		Path      string                          `json:"path" binding:"required"`
		FileTypes []domain_file_entity.FileTypeNo `json:"file_types" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		scene_audio_route_api_controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	library, err := ctrl.uc.CreateLibrary(c.Request.Context(), req.Name, req.Path, req.FileTypes)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain_file_entity.ErrLibraryDuplicate) {
			status = http.StatusConflict
		}
		scene_audio_route_api_controller.ErrorResponse(c, status, "LIBRARY_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"library": library,
	})
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
		scene_audio_route_api_controller.ErrorResponse(c, status, "LIBRARY_ERROR", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (ctrl *LibraryController) GetLibraries(c *gin.Context) {
	libraries, err := ctrl.uc.GetLibraries(c.Request.Context())
	if err != nil {
		scene_audio_route_api_controller.ErrorResponse(c, http.StatusInternalServerError, "LIBRARY_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"libraries": libraries,
	})
}
