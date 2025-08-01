package scene_audio_route_api_controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
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
	var params struct {
		Start               string `form:"start" binding:"required"`
		End                 string `form:"end" binding:"required"`
		Sort                string `form:"sort"`
		Order               string `form:"order"`
		Search              string `form:"search"`
		Starred             string `form:"starred"`
		AlbumID             string `form:"album_id"`
		ArtistID            string `form:"artist_id"`
		Year                string `form:"year"`
		Suffix              string `form:"suffix"`
		MinBitRate          string `form:"min_bitrate"`
		MaxBitRate          string `form:"max_bitrate"`
		FolderPath          string `form:"folder_path"`
		FolderPathSubFilter string `form:"folder_path_sub_filter"`
	}

	if err := ctx.ShouldBindQuery(&params); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "PARAMS_ERROR", parseBindingError(err))
		return
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
		params.Suffix,
		params.MinBitRate,
		params.MaxBitRate,
		params.FolderPath,
		params.FolderPathSubFilter,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "mediaFiles", mediaFiles, len(mediaFiles))
}

func (c *MediaFileController) GetMediaFileMetadatas(ctx *gin.Context) {
	var params struct {
		Start               string `form:"start" binding:"required"`
		End                 string `form:"end" binding:"required"`
		Sort                string `form:"sort"`
		Order               string `form:"order"`
		Search              string `form:"search"`
		Starred             string `form:"starred"`
		AlbumID             string `form:"album_id"`
		ArtistID            string `form:"artist_id"`
		Year                string `form:"year"`
		Suffix              string `form:"suffix"`
		MinBitRate          string `form:"min_bitrate"`
		MaxBitRate          string `form:"max_bitrate"`
		FolderPath          string `form:"folder_path"`
		FolderPathSubFilter string `form:"folder_path_sub_filter"`
	}

	if err := ctx.ShouldBindQuery(&params); err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "PARAMS_ERROR", parseBindingError(err))
		return
	}

	mediaFiles, err := c.MediaFileUsecase.GetMediaFileMetadataItems(
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
		params.Suffix,
		params.MinBitRate,
		params.MaxBitRate,
		params.FolderPath,
		params.FolderPathSubFilter,
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
	var params struct {
		Start               string   `form:"start" binding:"required"`
		End                 string   `form:"end" binding:"required"`
		Search              string   `form:"search"`
		Starred             string   `form:"starred"`
		AlbumID             string   `form:"album_id"`
		ArtistID            string   `form:"artist_id"`
		Year                string   `form:"year"`
		Suffix              string   `form:"suffix"`
		MinBitRate          string   `form:"min_bitrate"`
		MaxBitRate          string   `form:"max_bitrate"`
		FolderPath          string   `form:"folder_path"`
		FolderPathSubFilter string   `form:"folder_path_sub_filter"`
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
			controller.ErrorResponse(ctx, http.StatusBadRequest, "INVALID_SORT_FORMAT", "sort parameter must be in field:order format")
			return
		}
		sortOrders = append(sortOrders, domain_util.SortOrder{
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
		params.Suffix,
		params.MinBitRate,
		params.MaxBitRate,
		params.FolderPath,
		params.FolderPathSubFilter,
	)

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "mediaFiles", mediaFiles, len(mediaFiles))
}

func (c *MediaFileController) GetMediaFilterCounts(ctx *gin.Context) {
	counts, err := c.MediaFileUsecase.GetMediaFileFilterItemsCount(ctx.Request.Context())

	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "mediaFiles", counts, 1)
}
