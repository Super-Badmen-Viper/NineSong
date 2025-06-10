package scene_audio_route_api_controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"net/http"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
)

type MediaFileCueController struct {
	MediaFileUsecase scene_audio_route_interface.MediaFileCueRepository
}

func NewMediaFileCueController(uc scene_audio_route_interface.MediaFileCueRepository) *MediaFileCueController {
	return &MediaFileCueController{MediaFileUsecase: uc}
}

func (c *MediaFileCueController) GetMediaFileCues(ctx *gin.Context) {
	params := struct {
		Start    string `form:"start" binding:"required"`
		End      string `form:"end" binding:"required"`
		Sort     string `form:"sort"`
		Order    string `form:"order"`
		Search   string `form:"search"`
		Starred  string `form:"starred"`
		AlbumID  string `form:"album_id"`
		ArtistID string `form:"artist_id"`
		Year     string `form:"year"`
	}{
		Start:    ctx.Query("start"),
		End:      ctx.Query("end"),
		Sort:     ctx.DefaultQuery("sort", "title"),
		Order:    ctx.DefaultQuery("order", "asc"),
		Search:   ctx.Query("search"),
		Starred:  ctx.Query("starred"),
		AlbumID:  ctx.Query("album_id"),
		ArtistID: ctx.Query("artist_id"),
		Year:     ctx.Query("year"),
	}

	cueFiles, err := c.MediaFileUsecase.GetMediaFileCueItems(
		ctx.Request.Context(),
		params.Start,
		params.End,
		params.Sort,
		params.Order,
		params.Search,
		params.Starred,
		params.AlbumID,
		params.ArtistID,
		params.Year,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "cueFiles", cueFiles, len(cueFiles))
}

func (c *MediaFileCueController) GetMediaCueFilterCounts(ctx *gin.Context) {
	params := struct {
		Search   string `form:"search"`
		Starred  string `form:"starred"`
		AlbumID  string `form:"album_id"`
		ArtistID string `form:"artist_id"`
		Year     string `form:"year"`
	}{
		Search:   ctx.Query("search"),
		Starred:  ctx.Query("starred"),
		AlbumID:  ctx.Query("album_id"),
		ArtistID: ctx.Query("artist_id"),
		Year:     ctx.Query("year"),
	}

	counts, err := c.MediaFileUsecase.GetMediaFileCueFilterItemsCount(
		ctx.Request.Context(),
		params.Search,
		params.Starred,
		params.AlbumID,
		params.ArtistID,
		params.Year,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "cueFiles", counts, 1)
}
