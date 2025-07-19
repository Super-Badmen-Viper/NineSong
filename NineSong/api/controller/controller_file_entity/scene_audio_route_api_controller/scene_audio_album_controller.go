package scene_audio_route_api_controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type AlbumController struct {
	AlbumUsecase scene_audio_route_interface.AlbumRepository
}

func NewAlbumController(uc scene_audio_route_interface.AlbumRepository) *AlbumController {
	return &AlbumController{AlbumUsecase: uc}
}

func (c *AlbumController) GetAlbumItems(ctx *gin.Context) {
	params := struct {
		Start    string `form:"start" binding:"required"`
		End      string `form:"end" binding:"required"`
		Sort     string `form:"sort"`
		Order    string `form:"order"`
		Search   string `form:"search"`
		Starred  string `form:"starred"`
		ArtistID string `form:"artist_id"`
		MinYear  string `form:"min_year"`
		MaxYear  string `form:"max_year"`
	}{
		Start:    ctx.Query("start"),
		End:      ctx.Query("end"),
		Sort:     ctx.DefaultQuery("sort", "name"),
		Order:    ctx.DefaultQuery("order", "asc"),
		Search:   ctx.Query("search"),
		Starred:  ctx.Query("starred"),
		ArtistID: ctx.Query("artist_id"),
		MinYear:  ctx.Query("min_year"),
		MaxYear:  ctx.Query("max_year"),
	}

	if params.Start == "" || params.End == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMS", "必须提供start和end参数")
		return
	}

	albums, err := c.AlbumUsecase.GetAlbumItems(
		ctx.Request.Context(),
		params.Start,
		params.End,
		params.Sort,
		params.Order,
		params.Search,
		params.Starred,
		params.ArtistID,
		params.MinYear,
		params.MaxYear,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "albums", albums, len(albums))
}

func (c *AlbumController) GetAlbumItemsMultipleSorting(ctx *gin.Context) {
	params := struct {
		Start    string   `form:"start" binding:"required"`
		End      string   `form:"end" binding:"required"`
		Search   string   `form:"search"`
		Starred  string   `form:"starred"`
		ArtistID string   `form:"artist_id"`
		MinYear  string   `form:"min_year"`
		MaxYear  string   `form:"max_year"`
		Sort     []string `form:"sort"` // 格式: "field:order"
	}{
		Start:    ctx.Query("start"),
		End:      ctx.Query("end"),
		Search:   ctx.Query("search"),
		Starred:  ctx.Query("starred"),
		ArtistID: ctx.Query("artist_id"),
		MinYear:  ctx.Query("min_year"),
		MaxYear:  ctx.Query("max_year"),
		Sort:     ctx.QueryArray("sort"),
	}

	// 解析排序参数
	sortOrders := make([]domain_util.SortOrder, 0, len(params.Sort))
	for _, s := range params.Sort {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_SORT_FORMAT", "sort参数格式应为field:order")
			return
		}
		sortOrders = append(sortOrders, domain_util.SortOrder{
			Sort:  parts[0],
			Order: parts[1],
		})
	}

	albums, err := c.AlbumUsecase.GetAlbumItemsMultipleSorting(
		ctx.Request.Context(),
		params.Start,
		params.End,
		sortOrders,
		params.Search,
		params.Starred,
		params.ArtistID,
		params.MinYear,
		params.MaxYear,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "albums", albums, len(albums))
}

func (c *AlbumController) GetAlbumFilterCounts(ctx *gin.Context) {
	counts, err := c.AlbumUsecase.GetAlbumFilterItemsCount(ctx.Request.Context())

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "albums", counts, 1)
}
