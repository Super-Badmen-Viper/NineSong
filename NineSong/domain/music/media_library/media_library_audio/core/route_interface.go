package core

import (
	"context"
	"io"

	syncRecordCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/sync_record/core"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProgressCallback 进度回调函数类型
type ProgressCallback func(progress float64, status string, message string)

// MediaLibraryAudioUsecase 定义媒体库音频文件用例接口
type MediaLibraryAudioUsecase interface {
	// UploadAudioFile 上传音频文件
	UploadAudioFile(ctx context.Context, fileData io.Reader, fileName, libraryIDStr, uploaderIDStr string) (*MediaLibraryAudioResponse, error)

	// UploadAudioFileWithProgress 带进度回调的音频文件上传
	UploadAudioFileWithProgress(ctx context.Context, fileData io.Reader, fileName, libraryIDStr, uploaderIDStr string, progressCallback ProgressCallback) (*MediaLibraryAudioResponse, error)

	// UploadAudioFileChunked 分块上传音频文件
	UploadAudioFileChunked(ctx context.Context, fileData io.Reader, fileName, libraryIDStr, uploaderIDStr, uploadID string, chunkIndex, totalChunks int, isLastChunk bool) (*MediaLibraryAudioResponse, error)

	// DownloadAudioFile 下载音频文件
	DownloadAudioFile(ctx context.Context, fileID string) (io.Reader, string, error)

	// DownloadAudioFileWithProgress 带进度回调的音频文件下载
	DownloadAudioFileWithProgress(ctx context.Context, fileID string, progressCallback ProgressCallback) (io.Reader, string, error)

	// ExtractFileInfo 提取音频文件的详细信息
	ExtractFileInfo(ctx context.Context, fileID string) (*FileInfo, error)

	// GetAudioFileByID 获取音频文件信息
	GetAudioFileByID(ctx context.Context, fileID string) (*MediaLibraryAudioResponse, error)

	// GetAudioFilesByLibrary 获取指定媒体库的所有音频文件
	GetAudioFilesByLibrary(ctx context.Context, libraryID string) ([]*MediaLibraryAudioResponse, error)

	// GetAudioFilesByUploader 获取指定用户上传的所有音频文件
	GetAudioFilesByUploader(ctx context.Context, uploaderID string) ([]*MediaLibraryAudioResponse, error)

	// DeleteAudioFile 删除音频文件
	DeleteAudioFile(ctx context.Context, fileID string) error

	// GetUploadProgress 获取上传进度
	GetUploadProgress(ctx context.Context, uploadID string) (*MediaLibraryAudioResponse, error)

	// GetDownloadProgress 获取下载进度
	GetDownloadProgress(ctx context.Context, fileID string) (*MediaLibraryAudioResponse, error)
}

// MediaLibrarySyncRecordUsecase 定义媒体库同步记录用例接口
// 注意：这个接口实际上应该在sync_record模块中，但为了兼容性暂时放在这里
type MediaLibrarySyncRecordUsecase interface {
	// CreateSyncRecord 创建媒体库同步记录
	CreateSyncRecord(ctx context.Context, syncRecord *syncRecordCore.MediaLibrarySyncRecord) (*syncRecordCore.MediaLibrarySyncRecordResponse, error)

	// UpdateSyncRecord 更新媒体库同步记录
	UpdateSyncRecord(ctx context.Context, id string, syncRecord *syncRecordCore.MediaLibrarySyncRecord) (*syncRecordCore.MediaLibrarySyncRecordResponse, error)

	// GetSyncRecordByID 根据ID获取同步记录
	GetSyncRecordByID(ctx context.Context, id primitive.ObjectID) (*syncRecordCore.MediaLibrarySyncRecordResponse, error)

	// GetSyncRecordByMediaFileID 根据媒体文件ID获取同步记录
	GetSyncRecordByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (*syncRecordCore.MediaLibrarySyncRecordResponse, error)

	// GetSyncRecordByMediaLibraryAudioID 根据媒体库音频文件ID获取同步记录
	GetSyncRecordByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (*syncRecordCore.MediaLibrarySyncRecordResponse, error)

	// GetSyncRecordsByStatus 根据状态获取同步记录列表
	GetSyncRecordsByStatus(ctx context.Context, status string) ([]*syncRecordCore.MediaLibrarySyncRecordResponse, error)

	// GetSyncRecordsByType 根据同步类型获取同步记录列表
	GetSyncRecordsByType(ctx context.Context, syncType string) ([]*syncRecordCore.MediaLibrarySyncRecordResponse, error)

	// DeleteSyncRecordByID 根据ID删除同步记录
	DeleteSyncRecordByID(ctx context.Context, id primitive.ObjectID) error

	// UpdateSyncStatus 更新同步状态
	UpdateSyncStatus(ctx context.Context, id primitive.ObjectID, status string, progress float64, errorMessage string) error

	// UpdateSyncProgress 更新同步进度
	UpdateSyncProgress(ctx context.Context, id primitive.ObjectID, progress float64, lastUpdateTime primitive.DateTime) error

	// UpdateSyncResult 更新同步结果
	UpdateSyncResult(ctx context.Context, id primitive.ObjectID, status string, progress float64, errorMessage string, filePath string, fileSize int64, checksum string) error

	// DeleteSyncRecordByMediaFileID 根据媒体文件ID删除同步记录
	DeleteSyncRecordByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) error

	// DeleteSyncRecordByMediaLibraryAudioID 根据媒体库音频文件ID删除同步记录
	DeleteSyncRecordByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) error

	// GetSyncStatistics 获取同步统计信息
	GetSyncStatistics(ctx context.Context) (map[string]interface{}, error)

	// ExistsByMediaFileID 检查是否已存在关联指定媒体文件ID的同步记录
	ExistsByMediaFileID(ctx context.Context, mediaFileID primitive.ObjectID) (bool, error)

	// ExistsByMediaLibraryAudioID 检查是否已存在关联指定媒体库音频文件ID的同步记录
	ExistsByMediaLibraryAudioID(ctx context.Context, mediaLibraryAudioID primitive.ObjectID) (bool, error)
}
