package scene_audio_route_api_controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HomeController struct {
	usecase scene_audio_route_interface.HomeRepository
}

func NewHomeController(uc scene_audio_route_interface.HomeRepository) *HomeController {
	return &HomeController{usecase: uc}
}

func (c *HomeController) parsePagination(ctx *gin.Context) (start, end string) {
	start = ctx.DefaultQuery("start", "0")
	end = ctx.DefaultQuery("end", "50")
	return
}

func (c *HomeController) GetRandomArtistList(ctx *gin.Context) {
	start, end := c.parsePagination(ctx)

	artists, err := c.usecase.GetRandomArtistList(ctx, start, end)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "artists", artists, len(artists))
}

func (c *HomeController) GetRandomAlbumList(ctx *gin.Context) {
	start, end := c.parsePagination(ctx)

	albums, err := c.usecase.GetRandomAlbumList(ctx, start, end)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "albums", albums, len(albums))
}

func (c *HomeController) GetRandomMediaFileList(ctx *gin.Context) {
	start, end := c.parsePagination(ctx)

	mediaFiles, err := c.usecase.GetRandomMediaFileList(ctx, start, end)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "mediaFiles", mediaFiles, len(mediaFiles))
}
