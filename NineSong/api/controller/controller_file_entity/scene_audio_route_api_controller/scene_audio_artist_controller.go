package scene_audio_route_api_controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type ArtistController struct {
	ArtistUsecase scene_audio_route_interface.ArtistRepository
}

func NewArtistController(uc scene_audio_route_interface.ArtistRepository) *ArtistController {
	return &ArtistController{ArtistUsecase: uc}
}

func (c *ArtistController) GetArtists(ctx *gin.Context) {
	params := struct {
		Start   string `form:"start" binding:"required"`
		End     string `form:"end" binding:"required"`
		Sort    string `form:"sort"`
		Order   string `form:"order"`
		Search  string `form:"search"`
		Starred string `form:"starred"`
	}{
		Start:   ctx.Query("start"),
		End:     ctx.Query("end"),
		Sort:    ctx.DefaultQuery("sort", "name"),
		Order:   ctx.DefaultQuery("order", "asc"),
		Search:  ctx.Query("search"),
		Starred: ctx.Query("starred"),
	}

	artists, err := c.ArtistUsecase.GetArtistItems(
		ctx.Request.Context(),
		params.Start,
		params.End,
		params.Sort,
		params.Order,
		params.Search,
		params.Starred,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "artists", artists, len(artists))
}

func (c *ArtistController) GetArtistsMultipleSorting(ctx *gin.Context) {
	params := struct {
		Start   string   `form:"start" binding:"required"`
		End     string   `form:"end" binding:"required"`
		Search  string   `form:"search"`
		Starred string   `form:"starred"`
		Sort    []string `form:"sort"` // 格式: "field:order"
	}{
		Start:   ctx.Query("start"),
		End:     ctx.Query("end"),
		Search:  ctx.Query("search"),
		Starred: ctx.Query("starred"),
		Sort:    ctx.QueryArray("sort"),
	}

	// 解析排序参数
	sortOrders := make([]domain.SortOrder, 0, len(params.Sort))
	for _, s := range params.Sort {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_SORT_FORMAT", "sort参数格式应为field:order")
			return
		}
		sortOrders = append(sortOrders, domain.SortOrder{
			Sort:  parts[0],
			Order: parts[1],
		})
	}

	artists, err := c.ArtistUsecase.GetArtistItemsMultipleSorting(
		ctx.Request.Context(),
		params.Start,
		params.End,
		sortOrders,
		params.Search,
		params.Starred,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "artists", artists, len(artists))
}

func (c *ArtistController) GetArtistFilterCounts(ctx *gin.Context) {
	params := struct {
		Search  string `form:"search"`
		Starred string `form:"starred"`
	}{
		Search:  ctx.Query("search"),
		Starred: ctx.Query("starred"),
	}

	counts, err := c.ArtistUsecase.GetArtistFilterItemsCount(
		ctx.Request.Context(),
		params.Search,
		params.Starred,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "artists", counts, 1)
}
