package scene_audio_route_api_controller

import (
	"net/http"
	"strconv"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
)

type AnnotationController struct {
	usecase scene_audio_route_interface.AnnotationRepository
}

func NewAnnotationController(uc scene_audio_route_interface.AnnotationRepository) *AnnotationController {
	return &AnnotationController{usecase: uc}
}

type BaseAnnotationRequest struct {
	ItemID   string `form:"item_id" binding:"required"`
	ItemType string `form:"item_type" binding:"required,oneof=artist album media media_cue"`
}

type UpdateRatingRequest struct {
	BaseAnnotationRequest
	Rating int `form:"rating" json:"rating" binding:"required,min=0,max=5"`
}

func (c *AnnotationController) UpdateStarred(ctx *gin.Context) {
	var req BaseAnnotationRequest
	if err := ctx.ShouldBind(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
		return
	}

	result, err := c.usecase.UpdateStarred(ctx, req.ItemID, req.ItemType)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", result, 1)
}

func (c *AnnotationController) UpdateUnStarred(ctx *gin.Context) {
	var req BaseAnnotationRequest
	if err := ctx.ShouldBind(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
		return
	}

	result, err := c.usecase.UpdateUnStarred(ctx, req.ItemID, req.ItemType)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", result, 1)
}

func (c *AnnotationController) UpdateRating(ctx *gin.Context) {
	// 从查询参数和表单数据中获取值
	itemID := ctx.Query("item_id")
	itemType := ctx.Query("item_type")
	ratingStr := ctx.Query("rating")

	// 验证必需参数
	if itemID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", "item_id is required")
		return
	}
	if itemType == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", "item_type is required")
		return
	}
	if ratingStr == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", "rating is required")
		return
	}

	// 验证 item_type
	validTypes := map[string]bool{"artist": true, "album": true, "media": true, "media_cue": true}
	if !validTypes[itemType] {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", "invalid item_type, must be artist/album/media/media_cue")
		return
	}

	// 转换 rating 为整数
	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", "rating must be a valid integer")
		return
	}

	// 验证 rating 范围
	if rating < 0 || rating > 5 {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", "rating must be between 0-5")
		return
	}

	result, err := c.usecase.UpdateRating(ctx, itemID, itemType, rating)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", result, 1)
}

func (c *AnnotationController) UpdateScrobble(ctx *gin.Context) {
	var req BaseAnnotationRequest
	if err := ctx.ShouldBind(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
		return
	}

	result, err := c.usecase.UpdateScrobble(ctx, req.ItemID, req.ItemType)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", result, 1)
}

func (c *AnnotationController) UpdateCompleteScrobble(ctx *gin.Context) {
	var req BaseAnnotationRequest
	if err := ctx.ShouldBind(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
		return
	}

	result, err := c.usecase.UpdateCompleteScrobble(ctx, req.ItemID, req.ItemType)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", result, 1)
}

type UpdateTagSourceRequest struct {
	ItemID   string                            `json:"item_id" form:"item_id" binding:"required"`
	ItemType string                            `json:"item_type" form:"item_type" binding:"required,oneof=artist album media media_cue"`
	Tags     []scene_audio_db_models.TagSource `json:"tags" binding:"required"`
}

type UpdateWeightedTagRequest struct {
	ItemID   string                              `json:"item_id" form:"item_id" binding:"required"`
	ItemType string                              `json:"item_type" form:"item_type" binding:"required,oneof=artist album media media_cue"`
	Tags     []scene_audio_db_models.WeightedTag `json:"tags" binding:"required"`
}

func (c *AnnotationController) UpdateTagSource(ctx *gin.Context) {
	var req UpdateTagSourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
		return
	}

	result, err := c.usecase.UpdateTagSource(ctx, req.ItemID, req.ItemType, req.Tags)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "TAG_UPDATE_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", result, 1)
}

func (c *AnnotationController) UpdateWeightedTag(ctx *gin.Context) {
	var req UpdateWeightedTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
		return
	}

	result, err := c.usecase.UpdateWeightedTag(ctx, req.ItemID, req.ItemType, req.Tags)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "WEIGHT_UPDATE_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", result, 1)
}
