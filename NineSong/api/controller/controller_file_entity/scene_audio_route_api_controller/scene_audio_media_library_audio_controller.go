package scene_audio_route_api_controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaLibraryAudioController struct {
	MediaLibraryAudioUsecase      scene_audio_route_interface.MediaLibraryAudioUsecase
	MediaLibrarySyncRecordUsecase scene_audio_route_interface.MediaLibrarySyncRecordUsecase
}

func NewMediaLibraryAudioController(uc scene_audio_route_interface.MediaLibraryAudioUsecase, syncRecordUsecase scene_audio_route_interface.MediaLibrarySyncRecordUsecase) *MediaLibraryAudioController {
	return &MediaLibraryAudioController{
		MediaLibraryAudioUsecase:      uc,
		MediaLibrarySyncRecordUsecase: syncRecordUsecase,
	}
}

// UploadHandler 处理音频文件上传
func (c *MediaLibraryAudioController) UploadHandler(ctx *gin.Context) {
	// 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "FILE_UPLOAD_ERROR", "无法获取上传的文件")
		return
	}

	// 获取其他参数
	libraryID := ctx.PostForm("library_id")
	uploaderID := ctx.PostForm("uploader_id")

	// 验证必需参数
	if libraryID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少library_id参数")
		return
	}

	if uploaderID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少uploader_id参数")
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "FILE_OPEN_ERROR", "无法打开上传的文件")
		return
	}
	defer src.Close()

	// 创建同步记录
	syncRecord := &scene_audio_db_models.MediaLibrarySyncRecord{
		MediaFileID:         primitive.NilObjectID, // Will be updated after upload
		MediaLibraryAudioID: primitive.NilObjectID, // Will be updated after upload
		SyncStatus:          "uploading",
		SyncType:            "upload",
		FilePath:            file.Filename,
		CreatedAt:           primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:           primitive.NewDateTimeFromTime(time.Now()),
	}

	// 创建初始同步记录
	createdSyncRecord, err := c.MediaLibrarySyncRecordUsecase.CreateSyncRecord(ctx.Request.Context(), syncRecord)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SYNC_RECORD_CREATE_ERROR", "无法创建同步记录: "+err.Error())
		return
	}

	// 调用用例处理上传
	result, err := c.MediaLibraryAudioUsecase.UploadAudioFile(ctx.Request.Context(), src, file.Filename, libraryID, uploaderID)
	if err != nil {
		// 更新同步记录状态为失败
		updatedSyncRecord := &scene_audio_db_models.MediaLibrarySyncRecord{
			ID:                  createdSyncRecord.ID,
			MediaFileID:         createdSyncRecord.MediaFileID,
			MediaLibraryAudioID: createdSyncRecord.MediaLibraryAudioID,
			SyncStatus:          "failed",
			SyncType:            "upload",
			LastSyncTime:        primitive.NewDateTimeFromTime(time.Now()),
			CreatedAt:           createdSyncRecord.CreatedAt,
			UpdatedAt:           primitive.NewDateTimeFromTime(time.Now()),
			ErrorMessage:        err.Error(),
			Progress:            0.0,
			FileSize:            createdSyncRecord.FileSize,
			FilePath:            createdSyncRecord.FilePath,
			FileName:            createdSyncRecord.FileName,
			Checksum:            createdSyncRecord.Checksum,
		}

		_, updateErr := c.MediaLibrarySyncRecordUsecase.UpdateSyncRecord(ctx.Request.Context(), createdSyncRecord.ID.Hex(), updatedSyncRecord)
		if updateErr != nil {
			// Log the error but don't change the primary response
			fmt.Printf("Failed to update sync record status to failed: %v\n", updateErr)
		}

		controller.ErrorResponse(ctx, http.StatusInternalServerError, "UPLOAD_FAILED", err.Error())
		return
	}

	// 在媒体库音频场景中，MediaLibraryAudioID 和 MediaFileID 可能是同一个实体
	// 根据业务需求，我们可以将 MediaLibraryAudioID 设为结果中的 ID
	updatedSyncRecord := &scene_audio_db_models.MediaLibrarySyncRecord{
		ID:                  createdSyncRecord.ID,
		MediaFileID:         createdSyncRecord.MediaFileID, // Could be updated with result.ID if they represent the same entity
		MediaLibraryAudioID: result.ID,                     // 使用上传后生成的ID作为MediaLibraryAudioID
		SyncStatus:          "completed",
		SyncType:            "upload",
		LastSyncTime:        primitive.NewDateTimeFromTime(time.Now()),
		CreatedAt:           createdSyncRecord.CreatedAt,
		UpdatedAt:           primitive.NewDateTimeFromTime(time.Now()),
		ErrorMessage:        "",
		Progress:            100.0,
		FileSize:            result.FileSize,
		FilePath:            result.FilePath,
		FileName:            result.FileName,
		Checksum:            "", // Could be calculated if needed
	}

	_, updateErr := c.MediaLibrarySyncRecordUsecase.UpdateSyncRecord(ctx.Request.Context(), updatedSyncRecord.ID.Hex(), updatedSyncRecord)
	if updateErr != nil {
		// Log the error but don't change the primary response
		fmt.Printf("Failed to update sync record after successful upload: %v\n", updateErr)
	}

	controller.SuccessResponse(ctx, "result", result, 1)
}

// DownloadHandler 处理音频文件下载
func (c *MediaLibraryAudioController) DownloadHandler(ctx *gin.Context) {
	fileID := ctx.Param("file_id")
	if fileID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少file_id参数")
		return
	}

	// 创建同步记录
	syncRecord := &scene_audio_db_models.MediaLibrarySyncRecord{
		MediaFileID:         primitive.NilObjectID, // Will be updated after download
		MediaLibraryAudioID: primitive.NilObjectID, // Will be updated after download
		SyncStatus:          "downloading",
		SyncType:            "download",
		FilePath:            fileID, // Using file ID as identifier
		CreatedAt:           primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:           primitive.NewDateTimeFromTime(time.Now()),
	}

	// 创建初始同步记录
	createdSyncRecord, err := c.MediaLibrarySyncRecordUsecase.CreateSyncRecord(ctx.Request.Context(), syncRecord)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "SYNC_RECORD_CREATE_ERROR", "无法创建同步记录: "+err.Error())
		return
	}

	// 调用用例获取文件信息
	_, filePath, err := c.MediaLibraryAudioUsecase.DownloadAudioFile(ctx.Request.Context(), fileID)
	if err != nil {
		// 更新同步记录状态为失败
		updatedSyncRecord := &scene_audio_db_models.MediaLibrarySyncRecord{
			ID:                  createdSyncRecord.ID,
			MediaFileID:         createdSyncRecord.MediaFileID,
			MediaLibraryAudioID: createdSyncRecord.MediaLibraryAudioID,
			SyncStatus:          "failed",
			SyncType:            "download",
			LastSyncTime:        primitive.NewDateTimeFromTime(time.Now()),
			CreatedAt:           createdSyncRecord.CreatedAt,
			UpdatedAt:           primitive.NewDateTimeFromTime(time.Now()),
			ErrorMessage:        err.Error(),
			Progress:            0.0,
			FileSize:            createdSyncRecord.FileSize,
			FilePath:            createdSyncRecord.FilePath,
			FileName:            createdSyncRecord.FileName,
			Checksum:            createdSyncRecord.Checksum,
		}

		_, updateErr := c.MediaLibrarySyncRecordUsecase.UpdateSyncRecord(ctx.Request.Context(), createdSyncRecord.ID.Hex(), updatedSyncRecord)
		if updateErr != nil {
			// Log the error but don't change the primary response
			fmt.Printf("Failed to update sync record status to failed: %v\n", updateErr)
		}

		controller.ErrorResponse(ctx, http.StatusNotFound, "FILE_NOT_FOUND", "音频文件不存在")
		return
	}

	// 获取音频文件信息
	audioFile, err := c.MediaLibraryAudioUsecase.GetAudioFileByID(ctx.Request.Context(), fileID)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusNotFound, "FILE_NOT_FOUND", "音频文件不存在")
		return
	}

	// 更新同步记录，关联到实际的库音频ID
	updatedSyncRecord := &scene_audio_db_models.MediaLibrarySyncRecord{
		ID:                  createdSyncRecord.ID,
		MediaFileID:         createdSyncRecord.MediaFileID, // Could be updated with audioFile.ID if they represent the same entity
		MediaLibraryAudioID: audioFile.ID,                  // Use the downloaded file's ID
		SyncStatus:          "completed",
		SyncType:            "download",
		LastSyncTime:        primitive.NewDateTimeFromTime(time.Now()),
		CreatedAt:           createdSyncRecord.CreatedAt,
		UpdatedAt:           primitive.NewDateTimeFromTime(time.Now()),
		ErrorMessage:        "",
		Progress:            100.0,
		FileSize:            audioFile.FileSize,
		FilePath:            audioFile.FilePath,
		FileName:            audioFile.FileName,
		Checksum:            "", // Could be calculated if needed
	}

	_, updateErr := c.MediaLibrarySyncRecordUsecase.UpdateSyncRecord(ctx.Request.Context(), updatedSyncRecord.ID.Hex(), updatedSyncRecord)
	if updateErr != nil {
		// Log the error but don't change the primary response
		fmt.Printf("Failed to update sync record after successful download: %v\n", updateErr)
	}

	// 设置响应头
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename="+audioFile.FileName)
	ctx.Header("Content-Type", "application/octet-stream")

	// 提供文件下载
	ctx.File(filePath)
}

// GetFileInfoHandler 获取音频文件信息
func (c *MediaLibraryAudioController) GetFileInfoHandler(ctx *gin.Context) {
	fileID := ctx.Param("file_id")
	if fileID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少file_id参数")
		return
	}

	// 调用用例获取文件信息
	audioFile, err := c.MediaLibraryAudioUsecase.GetAudioFileByID(ctx.Request.Context(), fileID)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusNotFound, "FILE_NOT_FOUND", "音频文件不存在")
		return
	}

	controller.SuccessResponse(ctx, "result", gin.H{
		"id":          audioFile.ID.Hex(),
		"file_name":   audioFile.FileName,
		"file_size":   audioFile.FileSize,
		"file_path":   audioFile.FilePath,
		"file_type":   audioFile.FileType,
		"checksum":    audioFile.Checksum,
		"upload_time": audioFile.UploadTime.Time().Format("2006-01-02T15:04:05Z"),
		"created_at":  audioFile.CreatedAt.Time().Format("2006-01-02T15:04:05Z"),
		"updated_at":  audioFile.UpdatedAt.Time().Format("2006-01-02T15:04:05Z"),
		"library_id":  audioFile.LibraryID.Hex(),
		"uploader_id": audioFile.UploaderID.Hex(),
	}, 1)
}

// GetFilesByLibraryHandler 获取指定媒体库的所有音频文件
func (c *MediaLibraryAudioController) GetFilesByLibraryHandler(ctx *gin.Context) {
	libraryID := ctx.Param("library_id")
	if libraryID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少library_id参数")
		return
	}

	// 调用用例获取文件列表
	audioFiles, err := c.MediaLibraryAudioUsecase.GetAudioFilesByLibrary(ctx.Request.Context(), libraryID)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "GET_FILES_FAILED", err.Error())
		return
	}

	var results []gin.H
	for _, file := range audioFiles {
		results = append(results, gin.H{
			"id":          file.ID.Hex(),
			"file_name":   file.FileName,
			"file_size":   file.FileSize,
			"file_path":   file.FilePath,
			"created_at":  file.CreatedAt.Time().Format("2006-01-02T15:04:05Z"),
			"updated_at":  file.UpdatedAt.Time().Format("2006-01-02T15:04:05Z"),
			"library_id":  file.LibraryID.Hex(),
			"uploader_id": file.UploaderID.Hex(),
		})
	}

	controller.SuccessResponse(ctx, "results", results, len(results))
}

// GetFilesByUploaderHandler 获取指定用户上传的所有音频文件
func (c *MediaLibraryAudioController) GetFilesByUploaderHandler(ctx *gin.Context) {
	uploaderID := ctx.Param("uploader_id")
	if uploaderID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少uploader_id参数")
		return
	}

	// 调用用例获取文件列表
	audioFiles, err := c.MediaLibraryAudioUsecase.GetAudioFilesByUploader(ctx.Request.Context(), uploaderID)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "GET_FILES_FAILED", err.Error())
		return
	}

	var results []gin.H
	for _, file := range audioFiles {
		results = append(results, gin.H{
			"id":          file.ID.Hex(),
			"file_name":   file.FileName,
			"file_size":   file.FileSize,
			"file_path":   file.FilePath,
			"created_at":  file.CreatedAt.Time().Format("2006-01-02T15:04:05Z"),
			"updated_at":  file.UpdatedAt.Time().Format("2006-01-02T15:04:05Z"),
			"library_id":  file.LibraryID.Hex(),
			"uploader_id": file.UploaderID.Hex(),
		})
	}

	controller.SuccessResponse(ctx, "results", results, len(results))
}

// DeleteHandler 删除音频文件
func (c *MediaLibraryAudioController) DeleteHandler(ctx *gin.Context) {
	fileID := ctx.Param("file_id")
	if fileID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少file_id参数")
		return
	}

	// 调用用例删除文件
	err := c.MediaLibraryAudioUsecase.DeleteAudioFile(ctx.Request.Context(), fileID)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(ctx, "result", gin.H{
		"message": "音频文件删除成功",
	}, 1)
}

// GetUploadProgressHandler 获取上传进度
func (c *MediaLibraryAudioController) GetUploadProgressHandler(ctx *gin.Context) {
	uploadID := ctx.Param("upload_id")
	if uploadID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少upload_id参数")
		return
	}

	// 调用用例获取上传进度
	progress, err := c.MediaLibraryAudioUsecase.GetUploadProgress(ctx.Request.Context(), uploadID)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusNotFound, "PROGRESS_NOT_FOUND", "上传进度不存在或已过期")
		return
	}

	controller.SuccessResponse(ctx, "result", gin.H{
		"id":        progress.ID.Hex(),
		"file_name": progress.FileName,
		"status":    progress.Status,
		"progress":  progress.Progress,
	}, 1)
}

// GetDownloadProgressHandler 获取下载进度
func (c *MediaLibraryAudioController) GetDownloadProgressHandler(ctx *gin.Context) {
	fileID := ctx.Param("file_id")
	if fileID == "" {
		controller.ErrorResponse(ctx, http.StatusBadRequest, "MISSING_PARAMETER", "缺少file_id参数")
		return
	}

	// 调用用例获取下载进度
	progress, err := c.MediaLibraryAudioUsecase.GetDownloadProgress(ctx.Request.Context(), fileID)
	if err != nil {
		controller.ErrorResponse(ctx, http.StatusNotFound, "PROGRESS_NOT_FOUND", "下载进度不存在或已过期")
		return
	}

	controller.SuccessResponse(ctx, "result", gin.H{
		"id":        progress.ID.Hex(),
		"file_name": progress.FileName,
		"status":    progress.Status,
		"progress":  progress.Progress,
	}, 1)
}
