package scene_audio_route_api_controller

import (
	"errors"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"strconv"
	"strings"
)

type WordCloudController struct {
	WordCloudUsecase scene_audio_route_interface.WordCloudMetadata
}

func NewWordCloudController(uc scene_audio_route_interface.WordCloudMetadata) *WordCloudController {
	return &WordCloudController{WordCloudUsecase: uc}
}

func (c *WordCloudController) GetAllWordCloudHandler(ctx *gin.Context) {
	wordClouds, err := c.WordCloudUsecase.GetAllWordCloudSearch(ctx.Request.Context())
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	controller.SuccessResponse(ctx, "word_cloud", wordClouds, len(wordClouds))
}

func (c *WordCloudController) GetHighFrequencyWordCloudHandler(ctx *gin.Context) {
	params := struct {
		Limit int `form:"limit" binding:"required,numeric,min=1,max=100"`
	}{}

	if err := ctx.ShouldBindQuery(&params); err != nil {
		switch {
		case errors.Is(err, strconv.ErrRange):
			controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAM", "limit超出有效范围(1-100)")
		case errors.Is(err, strconv.ErrSyntax):
			controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAM", "limit必须是整数")
		default:
			controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAM", err.Error())
		}
		return
	}

	highWordCloud, err := c.WordCloudUsecase.GetHighFrequencyWordCloudSearch(
		ctx.Request.Context(),
		params.Limit,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	controller.SuccessResponse(ctx, "word_cloud", highWordCloud, 1)
}

func (c *WordCloudController) GetRecommendedWordCloudHandler(ctx *gin.Context) {
	var req struct {
		Keywords string `form:"keywords" binding:"required"`
	}

	if err := ctx.ShouldBindWith(&req, binding.Form); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "Missing keywords parameter")
		return
	}

	recommendWordCloud, err := c.WordCloudUsecase.GetRecommendedWordCloudSearch(
		ctx.Request.Context(),
		strings.Split(req.Keywords, ","),
	)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	controller.SuccessResponse(ctx, "word_cloud", recommendWordCloud, 1)
}
