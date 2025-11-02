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

func NewRecommendController(repo scene_audio_route_interface.RecommendRouteRepository) *RecommendController {
	return &RecommendController{
		RecommendUsecase: repo,
	}
}

func (c *RecommendController) GetGeneralRecommendations(ctx *gin.Context) {
	params := struct {
		RecommendType string `form:"recommend_type" binding:"required,oneof=artist album media media_cue"`
		Limit         string `form:"limit" binding:"required"`
		RandomSeed    string `form:"random_seed" binding:"required"`
		Offset        string `form:"recommend_offset" binding:"required"`
		LogShow       string `form:"log_show"` // 控制是否输出日志
		Refresh       string `form:"refresh"`  // 控制是否跳过缓存
	}{
		RecommendType: ctx.Query("recommend_type"),
		Limit:         ctx.Query("limit"),
		RandomSeed:    ctx.Query("random_seed"),
		Offset:        ctx.Query("recommend_offset"),
		LogShow:       ctx.Query("log_show"),
		Refresh:       ctx.Query("refresh"),
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

	// 验证offset参数
	_, err = strconv.Atoi(params.Offset)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_OFFSET", "offset参数必须是数字")
		return
	}

	// 解析logShow参数，默认为true（输出日志）
	logShow := true
	if params.LogShow == "false" {
		logShow = false
	}

	// 解析refresh参数，默认为false（使用缓存）
	refresh := false
	if params.Refresh == "true" {
		refresh = true
	}

	recommendResults, err := c.RecommendUsecase.GetGeneralRecommendations(
		ctx.Request.Context(),
		params.RecommendType,
		limit,
		params.RandomSeed,
		params.Offset,
		logShow,
		refresh,
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
		LogShow       string `form:"log_show"` // 控制是否输出日志
		Refresh       string `form:"refresh"`  // 控制是否跳过缓存
	}{
		UserId:        ctx.Query("user_id"),
		RecommendType: ctx.Query("recommend_type"),
		Limit:         ctx.Query("limit"),
		LogShow:       ctx.Query("log_show"),
		Refresh:       ctx.Query("refresh"),
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

	// 解析logShow参数，默认为true（输出日志）
	logShow := true
	if params.LogShow == "false" {
		logShow = false
	}

	// 解析refresh参数，默认为false（使用缓存）
	refresh := false
	if params.Refresh == "true" {
		refresh = true
	}

	recommendResults, err := c.RecommendUsecase.GetPersonalizedRecommendations(
		ctx.Request.Context(),
		params.UserId,
		params.RecommendType,
		limit,
		logShow,
		refresh,
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
		LogShow       string `form:"log_show"` // 控制是否输出日志
		Refresh       string `form:"refresh"`  // 控制是否跳过缓存
	}{
		RecommendType: ctx.Query("recommend_type"),
		Limit:         ctx.Query("limit"),
		LogShow:       ctx.Query("log_show"),
		Refresh:       ctx.Query("refresh"),
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

	// 解析logShow参数，默认为true（输出日志）
	logShow := true
	if params.LogShow == "false" {
		logShow = false
	}

	// 解析refresh参数，默认为false（使用缓存）
	refresh := false
	if params.Refresh == "true" {
		refresh = true
	}

	recommendResults, err := c.RecommendUsecase.GetPopularRecommendations(
		ctx.Request.Context(),
		params.RecommendType,
		limit,
		logShow,
		refresh,
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
