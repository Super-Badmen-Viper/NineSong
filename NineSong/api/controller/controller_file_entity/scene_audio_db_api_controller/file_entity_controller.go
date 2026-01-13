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
		// 增加日志输出，记录无效请求的详细信息
		log.Printf("[扫描请求] 无效请求参数: %v", err)
		log.Printf("[扫描请求] 请求原始数据: %+v", c.Request.Form)
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "无效的请求格式: "+err.Error())
		return
	}

	// 增加日志输出，记录有效的请求参数
	log.Printf("[扫描请求] 开始扫描 - FolderPath: '%s', FolderType: %d, ScanModel: %d",
		req.FolderPath, req.FolderType, req.ScanModel)

	if req.FolderPath != "" {
		if _, err := os.Stat(req.FolderPath); err != nil {
			if os.IsNotExist(err) {
				log.Printf("[扫描请求] 指定的目录不存在: %s", req.FolderPath)
				controller.ErrorResponse(c, http.StatusBadRequest, "DIRECTORY_NOT_FOUND",
					fmt.Sprintf("指定的目录不存在: %s", req.FolderPath))
				return
			}
			log.Printf("[扫描请求] 无法访问目录 %s: %v", req.FolderPath, err)
			controller.ErrorResponse(c, http.StatusInternalServerError, "DIRECTORY_ACCESS_ERROR",
				fmt.Sprintf("无法访问目录: %s (%v)", req.FolderPath, err))
			return
		}

		fileInfo, err := os.Stat(req.FolderPath)
		if err != nil {
			log.Printf("[扫描请求] 无法获取目录信息 %s: %v", req.FolderPath, err)
			controller.ErrorResponse(c, http.StatusInternalServerError, "DIRECTORY_STAT_ERROR",
				fmt.Sprintf("无法获取目录信息: %s (%v)", req.FolderPath, err))
			return
		}
		if !fileInfo.IsDir() {
			log.Printf("[扫描请求] 路径不是目录: %s", req.FolderPath)
			controller.ErrorResponse(c, http.StatusBadRequest, "NOT_A_DIRECTORY",
				fmt.Sprintf("路径不是目录: %s", req.FolderPath))
			return
		}
	}

	// 移除与修复逻辑相矛盾的验证 - 现在所有扫描模式都支持空FolderPath（默认扫描所有媒体库）
	// if req.ScanModel == 0 || req.ScanModel == 3 {
	// 	if len(req.FolderPath) == 0 {
	// 		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "扫描模式为新建或删除时，必须提供目录路径")
	// 		return
	// 	}
	// }

	var dirPaths []string
	if req.FolderPath != "" {
		dirPaths = append(dirPaths, req.FolderPath)
	}

	bgCtx := context.Background()
	go func() {
		if err := ctrl.usecase.ProcessDirectory(bgCtx, dirPaths, req.FolderType, req.ScanModel); err != nil {
			log.Printf(`扫描失败
				路径: %s
				错误类型: %T
				详细信息: %v`, req.FolderPath, err, err,
			)
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
	progress, startTime, activeScanCount, _ := ctrl.usecase.GetScanProgress()

	c.JSON(http.StatusOK, gin.H{
		"progress":          progress,
		"start_time":        startTime.Format(time.RFC3339),
		"active_scan_count": activeScanCount,
	})
}
