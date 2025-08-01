package scene_audio_route_api_controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type PlaylistTrackController struct {
	PlaylistTrackUsecase scene_audio_route_interface.PlaylistTrackRepository
}

func NewPlaylistTrackController(uc scene_audio_route_interface.PlaylistTrackRepository) *PlaylistTrackController {
	return &PlaylistTrackController{PlaylistTrackUsecase: uc}
}

func (c *PlaylistTrackController) GetPlaylistTracks(ctx *gin.Context) {
	var params struct {
		Start               string `form:"start" binding:"required"`
		End                 string `form:"end" binding:"required"`
		Sort                string `form:"sort" binding:"required"`
		Order               string `form:"order" binding:"required"`
		Search              string `form:"search"`
		Starred             string `form:"starred"`
		AlbumId             string `form:"albumId"`
		ArtistId            string `form:"artistId"`
		Year                string `form:"year"`
		Suffix              string `form:"suffix"`
		MinBitRate          string `form:"min_bitrate"`
		MaxBitRate          string `form:"max_bitrate"`
		FolderPath          string `form:"folder_path"`
		FolderPathSubFilter string `form:"folder_path_sub_filter"`
		PlaylistId          string `form:"playlistId" binding:"required"`
	}

	if err := ctx.ShouldBindQuery(&params); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "PARAMS_ERROR", parseBindingError(err))
		return
	}

	results, err := c.PlaylistTrackUsecase.GetPlaylistTrackItems(
		ctx.Request.Context(),
		params.Start,
		params.End,
		params.Sort,
		params.Order,
		params.Search,
		params.Starred,
		params.AlbumId,
		params.ArtistId,
		params.Year,
		params.PlaylistId,
		params.Suffix,
		params.MinBitRate,
		params.MaxBitRate,
		params.FolderPath,
		params.FolderPathSubFilter,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "mediaFiles", results, len(results))
}

func (c *PlaylistTrackController) GetPlaylistTracksMultipleSorting(ctx *gin.Context) {
	var params struct {
		Start               string   `form:"start" binding:"required"`
		End                 string   `form:"end" binding:"required"`
		Search              string   `form:"search"`
		Starred             string   `form:"starred"`
		AlbumId             string   `form:"albumId"`
		ArtistId            string   `form:"artistId"`
		Year                string   `form:"year"`
		Suffix              string   `form:"suffix"`
		MinBitRate          string   `form:"min_bitrate"`
		MaxBitRate          string   `form:"max_bitrate"`
		FolderPath          string   `form:"folder_path"`
		FolderPathSubFilter string   `form:"folder_path_sub_filter"`
		PlaylistId          string   `form:"playlistId" binding:"required"`
		Sort                []string `form:"sort"` // 格式: "field:order"
	}

	if err := ctx.ShouldBindQuery(&params); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "PARAMS_ERROR", parseBindingError(err))
		return
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

	results, err := c.PlaylistTrackUsecase.GetPlaylistTrackItemsMultipleSorting(
		ctx.Request.Context(),
		params.Start,
		params.End,
		sortOrders,
		params.Search,
		params.Starred,
		params.AlbumId,
		params.ArtistId,
		params.Year,
		params.PlaylistId,
		params.Suffix,
		params.MinBitRate,
		params.MaxBitRate,
		params.FolderPath,
		params.FolderPathSubFilter,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "mediaFiles", results, len(results))
}

func (c *PlaylistTrackController) GetPlaylistFilterCounts(ctx *gin.Context) {
	counts, err := c.PlaylistTrackUsecase.GetPlaylistTrackFilterItemsCount(ctx.Request.Context())

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "counts", counts, 1)
}

func (c *PlaylistTrackController) AddPlaylistTracks(ctx *gin.Context) {
	var req struct {
		PlaylistID   string `form:"playlist_id" binding:"required"`
		MediaFileIDs string `form:"media_file_ids" binding:"required"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "PARAMS_ERROR", parseBindingError(err))
		return
	}

	success, err := c.PlaylistTrackUsecase.AddPlaylistTrackItems(
		ctx.Request.Context(),
		req.PlaylistID,
		req.MediaFileIDs,
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		}
		controller.ErrorResponse(ctx, statusCode, "OPERATION_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", gin.H{"success": success}, 1)
}

func (c *PlaylistTrackController) RemovePlaylistTracks(ctx *gin.Context) {
	var req struct {
		PlaylistID   string `form:"playlist_id" binding:"required"`
		MediaFileIDs string `form:"media_file_ids" binding:"required"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "PARAMS_ERROR", parseBindingError(err))
		return
	}

	success, err := c.PlaylistTrackUsecase.RemovePlaylistTrackItems(
		ctx.Request.Context(),
		req.PlaylistID,
		req.MediaFileIDs,
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		}
		controller.ErrorResponse(ctx, statusCode, "OPERATION_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", gin.H{"success": success}, 1)
}

func (c *PlaylistTrackController) SortPlaylistTracks(ctx *gin.Context) {
	var req struct {
		PlaylistID   string `form:"playlist_id" binding:"required"`
		MediaFileIDs string `form:"media_file_ids" binding:"required"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "PARAMS_ERROR", parseBindingError(err))
		return
	}

	success, err := c.PlaylistTrackUsecase.SortPlaylistTrackItems(
		ctx.Request.Context(),
		req.PlaylistID,
		req.MediaFileIDs,
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		}
		controller.ErrorResponse(ctx, statusCode, "OPERATION_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", gin.H{"success": success}, 1)
}

func parseBindingError(err error) string {
	if err == nil {
		return ""
	}
	if strings.Contains(err.Error(), "required") {
		return "缺少必要参数"
	}
	if strings.Contains(err.Error(), "format") {
		return "参数格式错误"
	}
	return err.Error()
}
