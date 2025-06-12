package scene_audio_db_api_controller

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"time"
)

type FileController struct {
	usecase *usecase_file_entity.FileUsecase
}

func NewFileController(uc *usecase_file_entity.FileUsecase) *FileController {
	return &FileController{usecase: uc}
}

func (ctrl *FileController) ScanDirectory(c *gin.Context) {
	var req struct {
		FolderPath string `form:"folder_path"`
		FolderType int    `form:"folder_type" binding:"required"`
		ScanModel  int    `form:"scan_model" binding:"oneof=0 1 2 3"`
	}

	if err := c.ShouldBind(&req); err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "无效的请求格式: "+err.Error())
		return
	}

	if req.FolderPath != "" {
		if _, err := os.Stat(req.FolderPath); err != nil {
			if os.IsNotExist(err) {
				controller.ErrorResponse(c, http.StatusBadRequest, "DIRECTORY_NOT_FOUND",
					fmt.Sprintf("指定的目录不存在: %s", req.FolderPath))
				return
			}
			controller.ErrorResponse(c, http.StatusInternalServerError, "DIRECTORY_ACCESS_ERROR",
				fmt.Sprintf("无法访问目录: %s (%v)", req.FolderPath, err))
			return
		}

		fileInfo, err := os.Stat(req.FolderPath)
		if err != nil {
			controller.ErrorResponse(c, http.StatusInternalServerError, "DIRECTORY_STAT_ERROR",
				fmt.Sprintf("无法获取目录信息: %s (%v)", req.FolderPath, err))
			return
		}
		if !fileInfo.IsDir() {
			controller.ErrorResponse(c, http.StatusBadRequest, "NOT_A_DIRECTORY",
				fmt.Sprintf("路径不是目录: %s", req.FolderPath))
			return
		}
	}

	if req.ScanModel == 0 || req.ScanModel == 3 {
		if len(req.FolderPath) == 0 {
			controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "扫描模式为新建或删除时，必须提供目录路径")
			return
		}
	}

	var dirPaths []string
	if req.FolderPath != "" {
		dirPaths = append(dirPaths, req.FolderPath)
	}

	bgCtx := context.Background()
	go func() {
		if err := ctrl.usecase.ProcessDirectory(bgCtx, dirPaths, req.FolderType, req.ScanModel); err != nil {
			log.Printf("扫描失败 %s: %v", req.FolderPath, err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"ninesong-response": gin.H{
			"status":        "ok",
			"version":       controller.APIVersion,
			"type":          controller.ServiceType,
			"serverVersion": controller.ServerVersion,
			"message":       "后台处理已启动",
		},
	})
}

func (ctrl *FileController) GetScanProgress(c *gin.Context) {
	progress, startTime := ctrl.usecase.GetScanProgress()

	c.JSON(http.StatusOK, gin.H{
		"progress":   progress,
		"start_time": startTime.Format(time.RFC3339),
	})
}
