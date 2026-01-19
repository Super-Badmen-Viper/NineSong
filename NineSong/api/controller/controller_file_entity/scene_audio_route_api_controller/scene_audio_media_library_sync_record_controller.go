package scene_audio_route_api_controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
)

// MediaLibrarySyncRecordController 媒体库同步记录控制器
// 处理与媒体库同步记录相关的HTTP请求
type MediaLibrarySyncRecordController struct {
	mediaLibrarySyncRecordUseCase scene_audio_route_interface.MediaLibrarySyncRecordUsecase
}

// NewMediaLibrarySyncRecordController 创建新的媒体库同步记录控制器实例
func NewMediaLibrarySyncRecordController(mediaLibrarySyncRecordUseCase scene_audio_route_interface.MediaLibrarySyncRecordUsecase) *MediaLibrarySyncRecordController {
	return &MediaLibrarySyncRecordController{
		mediaLibrarySyncRecordUseCase: mediaLibrarySyncRecordUseCase,
	}
}

// CreateSyncRecord 创建同步记录
func (ctrl *MediaLibrarySyncRecordController) CreateSyncRecord(w http.ResponseWriter, r *http.Request) {
	var syncRecord scene_audio_db_models.MediaLibrarySyncRecord
	if err := json.NewDecoder(r.Body).Decode(&syncRecord); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	result, err := ctrl.mediaLibrarySyncRecordUseCase.CreateSyncRecord(ctx, &syncRecord)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetSyncRecordByID 根据ID获取同步记录
func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	result, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordByID(ctx, objectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == nil {
		http.Error(w, "Sync record not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetSyncRecordByMediaFileID 根据媒体文件ID获取同步记录
func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordByMediaFileID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("mediaFileId")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid media file ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	result, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordByMediaFileID(ctx, objectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == nil {
		http.Error(w, "Sync record not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetSyncRecordByMediaLibraryAudioID 根据媒体库音频文件ID获取同步记录
func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordByMediaLibraryAudioID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("mediaLibraryAudioId")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid media library audio ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	result, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordByMediaLibraryAudioID(ctx, objectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == nil {
		http.Error(w, "Sync record not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetSyncRecordsByStatus 根据同步状态获取同步记录列表
func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordsByStatus(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		http.Error(w, "Status parameter is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	results, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordsByStatus(ctx, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GetSyncRecordsByType 根据同步类型获取同步记录列表
func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordsByType(w http.ResponseWriter, r *http.Request) {
	syncType := r.URL.Query().Get("syncType")
	if syncType == "" {
		http.Error(w, "SyncType parameter is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	results, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordsByType(ctx, syncType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// UpdateSyncStatus 更新同步状态
func (ctrl *MediaLibrarySyncRecordController) UpdateSyncStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string  `json:"id"`
		Status       string  `json:"status"`
		Progress     float64 `json:"progress"`
		ErrorMessage string  `json:"errorMessage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = ctrl.mediaLibrarySyncRecordUseCase.UpdateSyncStatus(ctx, objectId, req.Status, req.Progress, req.ErrorMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sync status updated successfully"))
}

// UpdateSyncProgress 更新同步进度
func (ctrl *MediaLibrarySyncRecordController) UpdateSyncProgress(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID             string  `json:"id"`
		Progress       float64 `json:"progress"`
		LastUpdateTime int64   `json:"lastUpdateTime"` // Unix timestamp
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	lastUpdateTime := primitive.NewDateTimeFromTime(primitive.DateTime(req.LastUpdateTime).Time())

	ctx := context.Background()
	err = ctrl.mediaLibrarySyncRecordUseCase.UpdateSyncProgress(ctx, objectId, req.Progress, lastUpdateTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sync progress updated successfully"))
}

// UpdateSyncResult 更新同步结果
func (ctrl *MediaLibrarySyncRecordController) UpdateSyncResult(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string  `json:"id"`
		Status       string  `json:"status"`
		Progress     float64 `json:"progress"`
		ErrorMessage string  `json:"errorMessage"`
		FilePath     string  `json:"filePath"`
		FileSize     int64   `json:"fileSize"`
		Checksum     string  `json:"checksum"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = ctrl.mediaLibrarySyncRecordUseCase.UpdateSyncResult(ctx, objectId, req.Status, req.Progress, req.ErrorMessage, req.FilePath, req.FileSize, req.Checksum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sync result updated successfully"))
}

// DeleteSyncRecordByID 根据ID删除同步记录
func (ctrl *MediaLibrarySyncRecordController) DeleteSyncRecordByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = ctrl.mediaLibrarySyncRecordUseCase.DeleteSyncRecordByID(ctx, objectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sync record deleted successfully"))
}

// DeleteSyncRecordByMediaFileID 根据媒体文件ID删除同步记录
func (ctrl *MediaLibrarySyncRecordController) DeleteSyncRecordByMediaFileID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("mediaFileId")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid media file ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = ctrl.mediaLibrarySyncRecordUseCase.DeleteSyncRecordByMediaFileID(ctx, objectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sync record deleted successfully"))
}

// DeleteSyncRecordByMediaLibraryAudioID 根据媒体库音频文件ID删除同步记录
func (ctrl *MediaLibrarySyncRecordController) DeleteSyncRecordByMediaLibraryAudioID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("mediaLibraryAudioId")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid media library audio ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = ctrl.mediaLibrarySyncRecordUseCase.DeleteSyncRecordByMediaLibraryAudioID(ctx, objectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sync record deleted successfully"))
}

// GetSyncStatistics 获取同步统计信息
func (ctrl *MediaLibrarySyncRecordController) GetSyncStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	stats, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncStatistics(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ExistsByMediaFileID 检查是否已存在关联指定媒体文件ID的同步记录
func (ctrl *MediaLibrarySyncRecordController) ExistsByMediaFileID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("mediaFileId")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid media file ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	exists, err := ctrl.mediaLibrarySyncRecordUseCase.ExistsByMediaFileID(ctx, objectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"exists": exists})
}

// ExistsByMediaLibraryAudioID 检查是否已存在关联指定媒体库音频文件ID的同步记录
func (ctrl *MediaLibrarySyncRecordController) ExistsByMediaLibraryAudioID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("mediaLibraryAudioId")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid media library audio ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	exists, err := ctrl.mediaLibrarySyncRecordUseCase.ExistsByMediaLibraryAudioID(ctx, objectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"exists": exists})
}

// 为每个控制器方法添加Gin适配器
func (ctrl *MediaLibrarySyncRecordController) CreateSyncRecordAdapter(c *gin.Context) {
	var syncRecord scene_audio_db_models.MediaLibrarySyncRecord
	if err := c.ShouldBindJSON(&syncRecord); err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	result, err := ctrl.mediaLibrarySyncRecordUseCase.CreateSyncRecord(ctx, &syncRecord)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "CREATE_SYNC_RECORD_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "result", result, 1)
}

func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordByIDAdapter(c *gin.Context) {
	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID_FORMAT", "Invalid ID format")
		return
	}

	ctx := c.Request.Context()
	result, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordByID(ctx, objectId)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "GET_SYNC_RECORD_FAILED", err.Error())
		return
	}

	if result == nil {
		controller.ErrorResponse(c, http.StatusNotFound, "SYNC_RECORD_NOT_FOUND", "Sync record not found")
		return
	}

	controller.SuccessResponse(c, "result", result, 1)
}

func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordByMediaFileIDAdapter(c *gin.Context) {
	id := c.Param("media_file_id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_MEDIA_FILE_ID_FORMAT", "Invalid media file ID format")
		return
	}

	ctx := c.Request.Context()
	result, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordByMediaFileID(ctx, objectId)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "GET_SYNC_RECORD_FAILED", err.Error())
		return
	}

	if result == nil {
		controller.ErrorResponse(c, http.StatusNotFound, "SYNC_RECORD_NOT_FOUND", "Sync record not found")
		return
	}

	controller.SuccessResponse(c, "result", result, 1)
}

func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordByMediaLibraryAudioIDAdapter(c *gin.Context) {
	id := c.Param("media_library_audio_id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_MEDIA_LIBRARY_AUDIO_ID_FORMAT", "Invalid media library audio ID format")
		return
	}

	ctx := c.Request.Context()
	result, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordByMediaLibraryAudioID(ctx, objectId)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "GET_SYNC_RECORD_FAILED", err.Error())
		return
	}

	if result == nil {
		controller.ErrorResponse(c, http.StatusNotFound, "SYNC_RECORD_NOT_FOUND", "Sync record not found")
		return
	}

	controller.SuccessResponse(c, "result", result, 1)
}

func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordsByStatusAdapter(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		controller.ErrorResponse(c, http.StatusBadRequest, "MISSING_STATUS", "Status parameter is required")
		return
	}

	ctx := c.Request.Context()
	results, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordsByStatus(ctx, status)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "GET_SYNC_RECORDS_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "results", results, len(results))
}

func (ctrl *MediaLibrarySyncRecordController) GetSyncRecordsByTypeAdapter(c *gin.Context) {
	syncType := c.Param("type")
	if syncType == "" {
		controller.ErrorResponse(c, http.StatusBadRequest, "MISSING_TYPE", "SyncType parameter is required")
		return
	}

	ctx := c.Request.Context()
	results, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncRecordsByType(ctx, syncType)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "GET_SYNC_RECORDS_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "results", results, len(results))
}

func (ctrl *MediaLibrarySyncRecordController) UpdateSyncStatusAdapter(c *gin.Context) {
	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID_FORMAT", "Invalid ID format")
		return
	}

	var req struct {
		Status       string  `json:"status"`
		Progress     float64 `json:"progress"`
		ErrorMessage string  `json:"error_message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	err = ctrl.mediaLibrarySyncRecordUseCase.UpdateSyncStatus(ctx, objectId, req.Status, req.Progress, req.ErrorMessage)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_SYNC_STATUS_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "message", "Sync status updated successfully", 1)
}

func (ctrl *MediaLibrarySyncRecordController) UpdateSyncProgressAdapter(c *gin.Context) {
	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID_FORMAT", "Invalid ID format")
		return
	}

	var req struct {
		Progress       float64 `json:"progress"`
		LastUpdateTime int64   `json:"last_update_time"` // Unix timestamp
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format: "+err.Error())
		return
	}

	lastUpdateTime := primitive.NewDateTimeFromTime(time.Unix(req.LastUpdateTime, 0))

	ctx := c.Request.Context()
	err = ctrl.mediaLibrarySyncRecordUseCase.UpdateSyncProgress(ctx, objectId, req.Progress, lastUpdateTime)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_SYNC_PROGRESS_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "message", "Sync progress updated successfully", 1)
}

func (ctrl *MediaLibrarySyncRecordController) UpdateSyncResultAdapter(c *gin.Context) {
	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID_FORMAT", "Invalid ID format")
		return
	}

	var req struct {
		Status       string  `json:"status"`
		Progress     float64 `json:"progress"`
		ErrorMessage string  `json:"error_message"`
		FilePath     string  `json:"file_path"`
		FileSize     int64   `json:"file_size"`
		Checksum     string  `json:"checksum"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	err = ctrl.mediaLibrarySyncRecordUseCase.UpdateSyncResult(ctx, objectId, req.Status, req.Progress, req.ErrorMessage, req.FilePath, req.FileSize, req.Checksum)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_SYNC_RESULT_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "message", "Sync result updated successfully", 1)
}

func (ctrl *MediaLibrarySyncRecordController) DeleteSyncRecordByIDAdapter(c *gin.Context) {
	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID_FORMAT", "Invalid ID format")
		return
	}

	ctx := c.Request.Context()
	err = ctrl.mediaLibrarySyncRecordUseCase.DeleteSyncRecordByID(ctx, objectId)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "DELETE_SYNC_RECORD_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "message", "Sync record deleted successfully", 1)
}

func (ctrl *MediaLibrarySyncRecordController) DeleteSyncRecordByMediaFileIDAdapter(c *gin.Context) {
	id := c.Param("media_file_id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_MEDIA_FILE_ID_FORMAT", "Invalid media file ID format")
		return
	}

	ctx := c.Request.Context()
	err = ctrl.mediaLibrarySyncRecordUseCase.DeleteSyncRecordByMediaFileID(ctx, objectId)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "DELETE_SYNC_RECORD_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "message", "Sync record deleted successfully", 1)
}

func (ctrl *MediaLibrarySyncRecordController) DeleteSyncRecordByMediaLibraryAudioIDAdapter(c *gin.Context) {
	id := c.Param("media_library_audio_id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_MEDIA_LIBRARY_AUDIO_ID_FORMAT", "Invalid media library audio ID format")
		return
	}

	ctx := c.Request.Context()
	err = ctrl.mediaLibrarySyncRecordUseCase.DeleteSyncRecordByMediaLibraryAudioID(ctx, objectId)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "DELETE_SYNC_RECORD_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "message", "Sync record deleted successfully", 1)
}

func (ctrl *MediaLibrarySyncRecordController) GetSyncStatisticsAdapter(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := ctrl.mediaLibrarySyncRecordUseCase.GetSyncStatistics(ctx)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "GET_SYNC_STATISTICS_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "statistics", stats, 1)
}

func (ctrl *MediaLibrarySyncRecordController) ExistsByMediaFileIDAdapter(c *gin.Context) {
	id := c.Param("media_file_id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_MEDIA_FILE_ID_FORMAT", "Invalid media file ID format")
		return
	}

	ctx := c.Request.Context()
	exists, err := ctrl.mediaLibrarySyncRecordUseCase.ExistsByMediaFileID(ctx, objectId)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "CHECK_SYNC_RECORD_EXISTS_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "exists", exists, 1)
}

func (ctrl *MediaLibrarySyncRecordController) ExistsByMediaLibraryAudioIDAdapter(c *gin.Context) {
	id := c.Param("media_library_audio_id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		controller.ErrorResponse(c, http.StatusBadRequest, "INVALID_MEDIA_LIBRARY_AUDIO_ID_FORMAT", "Invalid media library audio ID format")
		return
	}

	ctx := c.Request.Context()
	exists, err := ctrl.mediaLibrarySyncRecordUseCase.ExistsByMediaLibraryAudioID(ctx, objectId)
	if err != nil {
		controller.ErrorResponse(c, http.StatusInternalServerError, "CHECK_SYNC_RECORD_EXISTS_FAILED", err.Error())
		return
	}

	controller.SuccessResponse(c, "exists", exists, 1)
}
