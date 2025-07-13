package scene_audio_route_api_controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"net/http"
	"strings"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
)

type MediaFileController struct {
	MediaFileUsecase scene_audio_route_interface.MediaFileRepository
}

func NewMediaFileController(uc scene_audio_route_interface.MediaFileRepository) *MediaFileController {
	return &MediaFileController{MediaFileUsecase: uc}
}

func (c *MediaFileController) GetMediaFiles(ctx *gin.Context) {
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

	mediaFiles, err := c.MediaFileUsecase.GetMediaFileItems(
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

	controller.SuccessResponse(ctx, "mediaFiles", mediaFiles, len(mediaFiles))
}

func (c *MediaFileController) GetMediaFileIds(ctx *gin.Context) {
	params := struct {
		Ids string `form:"ids" binding:"required"`
	}{
		Ids: ctx.Query("ids"),
	}

	mediaFiles, err := c.MediaFileUsecase.GetMediaFileItemsIds(
		ctx.Request.Context(),
		strings.Split(params.Ids, ","),
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "mediaFiles", mediaFiles, len(mediaFiles))
}

func (c *MediaFileController) GetMediaFilesMultipleSorting(ctx *gin.Context) {
	params := struct {
		Start    string   `form:"start" binding:"required"`
		End      string   `form:"end" binding:"required"`
		Search   string   `form:"search"`
		Starred  string   `form:"starred"`
		AlbumID  string   `form:"album_id"`
		ArtistID string   `form:"artist_id"`
		Year     string   `form:"year"`
		Sort     []string `form:"sort"` // 格式: "field:order"
	}{
		Start:    ctx.Query("start"),
		End:      ctx.Query("end"),
		Search:   ctx.Query("search"),
		Starred:  ctx.Query("starred"),
		AlbumID:  ctx.Query("album_id"),
		ArtistID: ctx.Query("artist_id"),
		Year:     ctx.Query("year"),
		Sort:     ctx.QueryArray("sort"),
	}

	// 解析排序参数
	sortOrders := make([]domain.SortOrder, 0, len(params.Sort))
	for _, s := range params.Sort {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_SORT_FORMAT", "sort parameter must be in field:order format")
			return
		}
		sortOrders = append(sortOrders, domain.SortOrder{
			Sort:  parts[0],
			Order: parts[1],
		})
	}

	mediaFiles, err := c.MediaFileUsecase.GetMediaFileItemsMultipleSorting(
		ctx.Request.Context(),
		params.Start,
		params.End,
		sortOrders,
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

	controller.SuccessResponse(ctx, "mediaFiles", mediaFiles, len(mediaFiles))
}

func (c *MediaFileController) GetMediaFilterCounts(ctx *gin.Context) {
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

	counts, err := c.MediaFileUsecase.GetMediaFileFilterItemsCount(
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

	controller.SuccessResponse(ctx, "mediaFiles", counts, 1)
}
