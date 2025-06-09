package scene_audio_db_api_controller

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type FileController struct {
	usecase *usecase_file_entity.FileUsecase
}

func NewFileController(uc *usecase_file_entity.FileUsecase) *FileController {
	return &FileController{usecase: uc}
}

func (ctrl *FileController) ScanDirectory(c *gin.Context) {
	var req struct {
		FolderPaths []string                        `json:"folder_paths" binding:"required"` // 修改为数组支持多个路径
		FileTypes   []domain_file_entity.FileTypeNo `json:"file_types" binding:"required"`
		ScanModel   int                             `json:"scan_model" binding:"oneof=0 1 2"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		scene_audio_route_api_controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "无效的请求格式: "+err.Error())
		return
	}

	// 根据 ScanModel 检查 FolderPaths 是否为空
	if req.ScanModel == 0 || req.ScanModel == 2 {
		if len(req.FolderPaths) == 0 { // 判断数组是否为空[1,4](@ref)
			scene_audio_route_api_controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "扫描模式为新建或覆盖时，必须提供目录路径")
			return
		}
	}
	// 当 ScanModel 为 1（修复模式）时，允许 FolderPaths 为空

	bgCtx := context.Background()
	go func() {
		if err := ctrl.usecase.ProcessDirectory(bgCtx, req.FolderPaths, req.FileTypes, req.ScanModel); err != nil {
			log.Printf("扫描失败 %s: %v", req.FolderPaths, err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"ninesong-response": gin.H{
			"status":        "ok",
			"version":       scene_audio_route_api_controller.APIVersion,
			"type":          scene_audio_route_api_controller.ServiceType,
			"serverVersion": scene_audio_route_api_controller.ServerVersion,
			"message":       "后台处理已启动",
		},
	})
}
