package scene_audio_route_api_controller

import (
	"net/http"
	"strconv"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
)

type RecommendController struct {
	RecommendUsecase scene_audio_route_interface.RecommendRouteRepository
}

func NewRecommendController(uc scene_audio_route_interface.RecommendRouteRepository) *RecommendController {
	return &RecommendController{RecommendUsecase: uc}
}

func (c *RecommendController) GetRecommendAnnotationWordCloudItems(ctx *gin.Context) {
	params := struct {
		Start         string `form:"start" binding:"required"`
		End           string `form:"end" binding:"required"`
		RecommendType string `form:"recommend_type" binding:"required,oneof=artist album media media_cue"`
		RandomSeed    string `form:"random_seed" binding:"required"`
		Offset        string `form:"offset" binding:"required"`
	}{
		Start:         ctx.Query("start"),
		End:           ctx.Query("end"),
		RecommendType: ctx.Query("recommend_type"),
		RandomSeed:    ctx.Query("random_seed"),
		Offset:        ctx.Query("offset"),
	}

	// 验证start和end参数
	start, err := strconv.Atoi(params.Start)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_START", "start参数必须是数字")
		return
	}

	end, err := strconv.Atoi(params.End)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_END", "end参数必须是数字")
		return
	}

	// 验证start < end
	if start >= end {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_RANGE", "start必须小于end")
		return
	}

	// 验证offset参数
	_, err = strconv.Atoi(params.Offset)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_OFFSET", "offset参数必须是数字")
		return
	}

	recommendResults, err := c.RecommendUsecase.GetRecommendAnnotationWordCloudItems(
		ctx.Request.Context(),
		params.Start,
		params.End,
		params.RecommendType,
		params.RandomSeed,
		params.Offset,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	// 检查返回结果是否为空
	if recommendResults == nil {
		controller.ErrorResponse(ctx, http.StatusNotFound, "NO_DATA", "未找到推荐数据")
		return
	}

	controller.SuccessResponse(ctx, "recommendations", recommendResults, len(recommendResults))
}

func (c *RecommendController) GetPersonalizedRecommendations(ctx *gin.Context) {
	params := struct {
		UserId        string `form:"user_id"`
		RecommendType string `form:"recommend_type" binding:"required,oneof=artist album media media_cue"`
		Limit         string `form:"limit" binding:"required"`
	}{
		UserId:        ctx.Query("user_id"),
		RecommendType: ctx.Query("recommend_type"),
		Limit:         ctx.Query("limit"),
	}

	// 验证limit参数
	limit, err := strconv.Atoi(params.Limit)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_LIMIT", "limit参数必须是数字")
		return
	}

	if limit <= 0 {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_LIMIT", "limit参数必须大于0")
		return
	}

	recommendResults, err := c.RecommendUsecase.GetPersonalizedRecommendations(
		ctx.Request.Context(),
		params.UserId,
		params.RecommendType,
		limit,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	// 检查返回结果是否为空
	if recommendResults == nil {
		controller.ErrorResponse(ctx, http.StatusNotFound, "NO_DATA", "未找到推荐数据")
		return
	}

	controller.SuccessResponse(ctx, "recommendations", recommendResults, len(recommendResults))
}

func (c *RecommendController) GetPopularRecommendations(ctx *gin.Context) {
	params := struct {
		RecommendType string `form:"recommend_type" binding:"required,oneof=artist album media media_cue"`
		Limit         string `form:"limit" binding:"required"`
	}{
		RecommendType: ctx.Query("recommend_type"),
		Limit:         ctx.Query("limit"),
	}

	// 验证limit参数
	limit, err := strconv.Atoi(params.Limit)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_LIMIT", "limit参数必须是数字")
		return
	}

	if limit <= 0 {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_LIMIT", "limit参数必须大于0")
		return
	}

	recommendResults, err := c.RecommendUsecase.GetPopularRecommendations(
		ctx.Request.Context(),
		params.RecommendType,
		limit,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	// 检查返回结果是否为空
	if recommendResults == nil {
		controller.ErrorResponse(ctx, http.StatusNotFound, "NO_DATA", "未找到推荐数据")
		return
	}

	controller.SuccessResponse(ctx, "recommendations", recommendResults, len(recommendResults))
}
