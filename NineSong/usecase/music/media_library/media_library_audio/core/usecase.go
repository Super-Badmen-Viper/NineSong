package core

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	domainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/media_library_audio/core"
	fileUsecaseCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/music/file_entity/file/core"
	"github.com/dhowden/tag"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mediaLibraryAudioUsecase struct {
	mediaLibraryAudioRepo    domainCore.MediaLibraryAudioRepository
	chunkedUploadSessionRepo domainCore.ChunkedUploadSessionRepository
	basicFileUsecase         *fileUsecaseCore.FileUsecase // 使用基础文件用例来触发扫描
	timeout                  time.Duration
}

func NewMediaLibraryAudioUsecase(
	repo domainCore.MediaLibraryAudioRepository,
	chunkedUploadSessionRepo domainCore.ChunkedUploadSessionRepository,
	basicFileUsecase *fileUsecaseCore.FileUsecase,
	timeout time.Duration,
) domainCore.MediaLibraryAudioUsecase {
	return &mediaLibraryAudioUsecase{
		mediaLibraryAudioRepo:    repo,
		chunkedUploadSessionRepo: chunkedUploadSessionRepo,
		basicFileUsecase:         basicFileUsecase,
		timeout:                  timeout,
	}
}

// UploadAudioFile 处理音频文件上传
func (uc *mediaLibraryAudioUsecase) UploadAudioFile(ctx context.Context, fileData io.Reader, fileName, libraryIDStr, uploaderIDStr string) (*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 验证输入参数
	if fileName == "" {
		return nil, fmt.Errorf("file name cannot be empty")
	}

	// 将字符串ID转换为ObjectID
	libraryID, err := primitive.ObjectIDFromHex(libraryIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid library ID: %v", err)
	}

	uploaderID, err := primitive.ObjectIDFromHex(uploaderIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid uploader ID: %v", err)
	}

	// 创建临时文件来保存上传的数据
	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, fileName)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFilePath) // 清理临时文件

	// 将上传的数据写入临时文件
	fileSize, err := io.Copy(tempFile, fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// 计算文件校验和
	checksum, err := uc.calculateFileChecksum(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file checksum: %v", err)
	}

	// 检查是否已存在相同校验和的文件
	existingFile, err := uc.mediaLibraryAudioRepo.GetByChecksum(ctx, checksum)
	if err == nil && existingFile != nil {
		// 如果已存在相同的文件，则直接返回该文件信息
		response := &domainCore.MediaLibraryAudioResponse{
			ID:          existingFile.ID,
			FileName:    existingFile.FileName,
			FilePath:    existingFile.FilePath,
			StoragePath: existingFile.StoragePath,
			FileSize:    existingFile.FileSize,
			FileType:    existingFile.FileType,
			Checksum:    existingFile.Checksum,
			LibraryID:   existingFile.LibraryID,
			UploaderID:  existingFile.UploaderID,
			UploadTime:  existingFile.UploadTime,
			CreatedAt:   existingFile.CreatedAt,
			UpdatedAt:   existingFile.UpdatedAt,
		}
		// 即使文件已存在，也触发扫描以确保元数据是最新的
		go uc.triggerScanAfterUpload(existingFile.LibraryID.Hex())
		return response, nil
	}

	// 准备音频文件实体
	audioFile := &domainCore.MediaLibraryAudio{
		FileName:   fileName,
		FileSize:   fileSize,
		Checksum:   checksum,
		LibraryID:  libraryID,
		UploaderID: uploaderID,
		Status:     "uploading", // 初始状态
		IsUploaded: false,
	}

	// 验证实体
	if err := audioFile.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	// 保存到数据库
	if err := uc.mediaLibraryAudioRepo.Create(ctx, audioFile); err != nil {
		return nil, fmt.Errorf("failed to save audio file metadata: %v", err)
	}

	// 更新状态为processing
	if err := uc.mediaLibraryAudioRepo.UpdateStatus(ctx, audioFile.ID, "processing"); err != nil {
		log.Printf("Warning: failed to update status to processing: %v", err)
	}

	// 获取目标存储路径
	storagePath := uc.getStoragePath(fileName, libraryID.Hex())

	// 移动文件到最终位置
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	if err := os.Rename(tempFilePath, storagePath); err != nil {
		return nil, fmt.Errorf("failed to move file to storage: %v", err)
	}

	// 更新数据库记录
	audioFile.FilePath = storagePath
	audioFile.StoragePath = storagePath
	audioFile.Status = "ready"
	audioFile.IsUploaded = true
	audioFile.SetTimestamps()

	// 创建更新文档
	updateDoc := bson.M{
		"$set": bson.M{
			"file_path":    audioFile.FilePath,
			"storage_path": audioFile.StoragePath,
			"status":       audioFile.Status,
			"is_uploaded":  audioFile.IsUploaded,
			"updated_at":   audioFile.UpdatedAt,
		},
	}

	if _, err := uc.mediaLibraryAudioRepo.UpdateByID(ctx, audioFile.ID, updateDoc); err != nil {
		return nil, fmt.Errorf("failed to update audio file record: %v", err)
	}

	// 触发媒体库扫描
	go uc.triggerScanAfterUpload(libraryID.Hex())

	// 返回响应
	response := &domainCore.MediaLibraryAudioResponse{
		ID:          audioFile.ID,
		FileName:    audioFile.FileName,
		FilePath:    audioFile.FilePath,
		StoragePath: audioFile.StoragePath,
		FileSize:    audioFile.FileSize,
		FileType:    audioFile.FileType,
		Checksum:    audioFile.Checksum,
		LibraryID:   audioFile.LibraryID,
		UploaderID:  audioFile.UploaderID,
		UploadTime:  audioFile.UploadTime,
		CreatedAt:   audioFile.CreatedAt,
		UpdatedAt:   audioFile.UpdatedAt,
	}

	return response, nil
}

// UploadAudioFileWithProgress 带进度回调的音频文件上传
func (uc *mediaLibraryAudioUsecase) UploadAudioFileWithProgress(ctx context.Context, fileData io.Reader, fileName, libraryIDStr, uploaderIDStr string, progressCallback domainCore.ProgressCallback) (*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 验证输入参数
	if fileName == "" {
		return nil, fmt.Errorf("file name cannot be empty")
	}

	// 将字符串ID转换为ObjectID
	libraryID, err := primitive.ObjectIDFromHex(libraryIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid library ID: %v", err)
	}

	uploaderID, err := primitive.ObjectIDFromHex(uploaderIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid uploader ID: %v", err)
	}

	// 创建临时文件来保存上传的数据
	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, fileName)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFilePath) // 清理临时文件

	// 如果提供了进度回调，则包装源读取器以跟踪进度
	var sourceReader io.Reader
	if progressCallback != nil {
		// 首先报告初始状态
		progressCallback(0.0, "uploading", "Starting upload...")

		// 创建一个临时的io.ReadSeeker来获取文件大小
		// 因为我们不知道源文件的大小，我们需要先读取整个内容
		var buf bytes.Buffer
		_, err := io.Copy(&buf, fileData)
		if err != nil {
			return nil, fmt.Errorf("failed to read file data: %v", err)
		}

		// 创建带进度回调的读取器
		sourceReader = &ProgressReader{
			reader:           bytes.NewReader(buf.Bytes()),
			totalSize:        int64(buf.Len()),
			currentSize:      0,
			progressCallback: progressCallback,
		}
	} else {
		sourceReader = fileData
	}

	// 将上传的数据写入临时文件
	fileSize, err := io.Copy(tempFile, sourceReader)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// 计算文件校验和
	checksum, err := uc.calculateFileChecksum(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file checksum: %v", err)
	}

	// 检查是否已存在相同校验和的文件
	existingFile, err := uc.mediaLibraryAudioRepo.GetByChecksum(ctx, checksum)
	if err == nil && existingFile != nil {
		// 如果已存在相同的文件，则直接返回该文件信息
		response := &domainCore.MediaLibraryAudioResponse{
			ID:          existingFile.ID,
			FileName:    existingFile.FileName,
			FilePath:    existingFile.FilePath,
			StoragePath: existingFile.StoragePath,
			FileSize:    existingFile.FileSize,
			FileType:    existingFile.FileType,
			Checksum:    existingFile.Checksum,
			LibraryID:   existingFile.LibraryID,
			UploaderID:  existingFile.UploaderID,
			UploadTime:  existingFile.UploadTime,
			CreatedAt:   existingFile.CreatedAt,
			UpdatedAt:   existingFile.UpdatedAt,
		}
		// 即使文件已存在，也触发扫描以确保元数据是最新的
		go uc.triggerScanAfterUpload(existingFile.LibraryID.Hex())

		// 报告完成状态
		if progressCallback != nil {
			progressCallback(100.0, "completed", "File already exists, scan completed")
		}

		return response, nil
	}

	// 准备音频文件实体
	audioFile := &domainCore.MediaLibraryAudio{
		FileName:   fileName,
		FileSize:   fileSize,
		Checksum:   checksum,
		LibraryID:  libraryID,
		UploaderID: uploaderID,
		Status:     "uploading", // 初始状态
		IsUploaded: false,
	}

	// 验证实体
	if err := audioFile.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	// 保存到数据库
	if err := uc.mediaLibraryAudioRepo.Create(ctx, audioFile); err != nil {
		return nil, fmt.Errorf("failed to save audio file metadata: %v", err)
	}

	// 更新状态为processing
	if err := uc.mediaLibraryAudioRepo.UpdateStatus(ctx, audioFile.ID, "processing"); err != nil {
		log.Printf("Warning: failed to update status to processing: %v", err)
	}

	// 获取目标存储路径
	storagePath := uc.getStoragePath(fileName, libraryID.Hex())

	// 移动文件到最终位置
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	if err := os.Rename(tempFilePath, storagePath); err != nil {
		return nil, fmt.Errorf("failed to move file to storage: %v", err)
	}

	// 更新数据库记录
	audioFile.FilePath = storagePath
	audioFile.StoragePath = storagePath
	audioFile.Status = "ready"
	audioFile.IsUploaded = true
	audioFile.SetTimestamps()

	// 创建更新文档
	updateDoc := bson.M{
		"$set": bson.M{
			"file_path":    audioFile.FilePath,
			"storage_path": audioFile.StoragePath,
			"status":       audioFile.Status,
			"is_uploaded":  audioFile.IsUploaded,
			"updated_at":   audioFile.UpdatedAt,
		},
	}

	if _, err := uc.mediaLibraryAudioRepo.UpdateByID(ctx, audioFile.ID, updateDoc); err != nil {
		return nil, fmt.Errorf("failed to update audio file record: %v", err)
	}

	// 触发媒体库扫描
	go uc.triggerScanAfterUpload(libraryID.Hex())

	// 报告完成状态
	if progressCallback != nil {
		progressCallback(100.0, "completed", "Upload completed successfully")
	}

	// 返回响应
	response := &domainCore.MediaLibraryAudioResponse{
		ID:          audioFile.ID,
		FileName:    audioFile.FileName,
		FilePath:    audioFile.FilePath,
		StoragePath: audioFile.StoragePath,
		FileSize:    audioFile.FileSize,
		FileType:    audioFile.FileType,
		Checksum:    audioFile.Checksum,
		LibraryID:   audioFile.LibraryID,
		UploaderID:  audioFile.UploaderID,
		UploadTime:  audioFile.UploadTime,
		CreatedAt:   audioFile.CreatedAt,
		UpdatedAt:   audioFile.UpdatedAt,
	}

	return response, nil
}

// DownloadAudioFile 处理音频文件下载
func (uc *mediaLibraryAudioUsecase) DownloadAudioFile(ctx context.Context, fileIDStr string) (io.Reader, string, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	fileID, err := primitive.ObjectIDFromHex(fileIDStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid file ID: %v", err)
	}

	// 从数据库获取文件信息
	audioFile, err := uc.mediaLibraryAudioRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find audio file: %v", err)
	}

	// 增加下载计数
	go func() {
		// 在 goroutine 中异步增加下载计数
		downloadCtx, cancel := context.WithTimeout(context.Background(), uc.timeout)
		defer cancel()

		if err := uc.mediaLibraryAudioRepo.IncrementDownloadCount(downloadCtx, fileID); err != nil {
			log.Printf("Warning: failed to increment download count: %v", err)
		}
	}()

	// 打开文件并返回读取器
	fileReader, err := os.Open(audioFile.StoragePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open audio file: %v", err)
	}

	return fileReader, audioFile.FileName, nil
}

// DownloadAudioFileWithProgress 带进度回调的音频文件下载
func (uc *mediaLibraryAudioUsecase) DownloadAudioFileWithProgress(ctx context.Context, fileIDStr string, progressCallback domainCore.ProgressCallback) (io.Reader, string, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	fileID, err := primitive.ObjectIDFromHex(fileIDStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid file ID: %v", err)
	}

	// 从数据库获取文件信息
	audioFile, err := uc.mediaLibraryAudioRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find audio file: %v", err)
	}

	// 增加下载计数
	go func() {
		// 在 goroutine 中异步增加下载计数
		downloadCtx, cancel := context.WithTimeout(context.Background(), uc.timeout)
		defer cancel()

		if err := uc.mediaLibraryAudioRepo.IncrementDownloadCount(downloadCtx, fileID); err != nil {
			log.Printf("Warning: failed to increment download count: %v", err)
		}
	}()

	// 打开文件
	file, err := os.Open(audioFile.StoragePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open audio file: %v", err)
	}

	// 如果提供了进度回调，则包装文件读取器以跟踪进度
	if progressCallback != nil {
		// 首先报告初始状态
		progressCallback(0.0, "downloading", "Starting download...")

		// 包装原始文件读取器以跟踪进度
		progressReader := &ProgressReader{
			reader:           file,
			totalSize:        audioFile.FileSize,
			currentSize:      0,
			progressCallback: progressCallback,
		}

		return progressReader, audioFile.FileName, nil
	}

	return file, audioFile.FileName, nil
}

// GetAudioFileByID 根据ID获取音频文件信息
func (uc *mediaLibraryAudioUsecase) GetAudioFileByID(ctx context.Context, fileIDStr string) (*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	fileID, err := primitive.ObjectIDFromHex(fileIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %v", err)
	}

	// 从数据库获取文件信息
	audioFile, err := uc.mediaLibraryAudioRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find audio file: %v", err)
	}

	// 返回响应模型
	response := &domainCore.MediaLibraryAudioResponse{
		ID:          audioFile.ID,
		FileName:    audioFile.FileName,
		FilePath:    audioFile.FilePath,
		StoragePath: audioFile.StoragePath,
		FileSize:    audioFile.FileSize,
		FileType:    audioFile.FileType,
		Checksum:    audioFile.Checksum,
		LibraryID:   audioFile.LibraryID,
		UploaderID:  audioFile.UploaderID,
		UploadTime:  audioFile.UploadTime,
		CreatedAt:   audioFile.CreatedAt,
		UpdatedAt:   audioFile.UpdatedAt,
	}

	return response, nil
}

// GetAudioFilesByLibrary 获取指定媒体库的所有音频文件
func (uc *mediaLibraryAudioUsecase) GetAudioFilesByLibrary(ctx context.Context, libraryIDStr string) ([]*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	libraryID, err := primitive.ObjectIDFromHex(libraryIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid library ID: %v", err)
	}

	// 从数据库获取文件列表
	audioFiles, err := uc.mediaLibraryAudioRepo.GetByLibraryID(ctx, libraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio files by library: %v", err)
	}

	// 转换为响应模型
	var responses []*domainCore.MediaLibraryAudioResponse
	for _, file := range audioFiles {
		response := &domainCore.MediaLibraryAudioResponse{
			ID:          file.ID,
			FileName:    file.FileName,
			FilePath:    file.FilePath,
			StoragePath: file.StoragePath,
			FileSize:    file.FileSize,
			FileType:    file.FileType,
			Checksum:    file.Checksum,
			LibraryID:   file.LibraryID,
			UploaderID:  file.UploaderID,
			UploadTime:  file.UploadTime,
			CreatedAt:   file.CreatedAt,
			UpdatedAt:   file.UpdatedAt,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetAudioFilesByUploader 获取指定用户上传的所有音频文件
func (uc *mediaLibraryAudioUsecase) GetAudioFilesByUploader(ctx context.Context, uploaderIDStr string) ([]*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	uploaderID, err := primitive.ObjectIDFromHex(uploaderIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid uploader ID: %v", err)
	}

	// 从数据库获取文件列表
	audioFiles, err := uc.mediaLibraryAudioRepo.GetByUploaderID(ctx, uploaderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio files by uploader: %v", err)
	}

	// 转换为响应模型
	var responses []*domainCore.MediaLibraryAudioResponse
	for _, file := range audioFiles {
		response := &domainCore.MediaLibraryAudioResponse{
			ID:          file.ID,
			FileName:    file.FileName,
			FilePath:    file.FilePath,
			StoragePath: file.StoragePath,
			FileSize:    file.FileSize,
			FileType:    file.FileType,
			Checksum:    file.Checksum,
			LibraryID:   file.LibraryID,
			UploaderID:  file.UploaderID,
			UploadTime:  file.UploadTime,
			CreatedAt:   file.CreatedAt,
			UpdatedAt:   file.UpdatedAt,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// ExtractFileInfo 提取音频文件的详细信息
func (uc *mediaLibraryAudioUsecase) ExtractFileInfo(ctx context.Context, fileIDStr string) (*domainCore.FileInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	fileID, err := primitive.ObjectIDFromHex(fileIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %v", err)
	}

	// 从数据库获取文件信息
	audioFile, err := uc.mediaLibraryAudioRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find audio file: %v", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(audioFile.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	// 创建基本的文件信息结构
	result := &domainCore.FileInfo{
		ID:        fileID.Hex(),
		FileName:  audioFile.FileName,
		FileSize:  fileInfo.Size(),
		FileType:  audioFile.FileType,
		MimeType:  uc.getMimeType(audioFile.FileName),
		CreatedAt: audioFile.CreatedAt.Time().Format(time.RFC3339),
		UpdatedAt: audioFile.UpdatedAt.Time().Format(time.RFC3339),
	}

	// 使用音频分析库提取详细信息
	detailedInfo, err := uc.extractDetailedAudioInfo(audioFile.StoragePath)
	if err != nil {
		log.Printf("Warning: Failed to extract detailed audio info for %s: %v", audioFile.StoragePath, err)
		// 如果提取失败，仍然返回基本信息
		result.Format = uc.getFileExtension(audioFile.FileName)
		result.Duration = 0
		result.Bitrate = 0
		result.SampleRate = 0
		result.Channels = 0
		result.Encoding = "unknown"
	} else {
		// 使用提取到的详细信息
		result.Format = detailedInfo.Format
		result.Duration = detailedInfo.Duration
		result.Bitrate = detailedInfo.Bitrate
		result.SampleRate = detailedInfo.SampleRate
		result.Channels = detailedInfo.Channels
		result.Encoding = detailedInfo.Encoding
		result.Artist = detailedInfo.Artist
		result.Title = detailedInfo.Title
		result.Album = detailedInfo.Album
		result.Year = detailedInfo.Year
		result.Genre = detailedInfo.Genre
	}

	return result, nil
}

// extractDetailedAudioInfo 使用音频分析库提取详细信息
func (uc *mediaLibraryAudioUsecase) extractDetailedAudioInfo(filePath string) (*domainCore.FileInfo, error) {
	// 首先尝试使用 github.com/dhowden/tag 库提取基础音频信息
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 使用 dhowden/tag 提取基础元数据
	metadata, err := tag.ReadFrom(file)
	if err != nil {
		// 如果 dhowden/tag 无法解析，尝试使用更强大的 taglib 库
		return uc.extractDetailedAudioInfoWithTaglib(filePath)
	}

	// 获取文件信息用于 calculating duration and bitrate
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	// Create result with basic information
	result := &domainCore.FileInfo{
		Format:     uc.getFileExtension(filePath),
		FileSize:   fileInfo.Size(), // Use the file size from fileInfo
		Duration:   0,               // Need to calculate from more advanced library
		Bitrate:    0,               // Need to calculate
		SampleRate: 0,               // Some formats may not provide directly
		Channels:   0,               // Some formats may not provide directly
		Encoding:   "unknown",
		Artist:     metadata.Artist(),
		Title:      metadata.Title(),
		Album:      metadata.Album(),
		Year:       metadata.Year(),
		Genre:      metadata.Genre(),
	}

	// 对于 MP3 文件，尝试更精确地估算比特率和持续时间
	fileExt := strings.ToLower(filepath.Ext(filePath))
	switch fileExt {
	case ".mp3":
		// MP3 文件通常有固定或可变比特率
		// 这里做一个简单的估算，基于文件大小和假设的比特率
		// 更精确的方法需要使用专门的库分析MP3帧头
		result.Encoding = "MPEG Audio Layer 3"
	case ".wav":
		result.Encoding = "Waveform Audio File Format"
		// WAV 文件通常有固定的采样率和声道数
		// 如果我们能解析WAV头，可以获取更多细节
	case ".flac":
		result.Encoding = "Free Lossless Audio Codec"
	case ".aac":
		result.Encoding = "Advanced Audio Coding"
	case ".m4a":
		result.Encoding = "MPEG-4 Audio"
	case ".ogg":
		result.Encoding = "Ogg Vorbis"
	case ".wma":
		result.Encoding = "Windows Media Audio"
	}

	return result, nil
}

// extractDetailedAudioInfoWithTaglib 使用 taglib 库提取更详细的音频信息
func (uc *mediaLibraryAudioUsecase) extractDetailedAudioInfoWithTaglib(filePath string) (*domainCore.FileInfo, error) {
	// 由于 taglib 库的 extractor 需要复杂的参数，我们使用 dhowden/tag 作为后备方案
	// 首先尝试使用 dhowden/tag 读取元数据
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio metadata: %v", err)
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	// 基础信息填充
	result := &domainCore.FileInfo{
		Format:     uc.getFileExtension(filePath),
		FileSize:   fileInfo.Size(),
		Duration:   0,                                      // 这个需要专门的音频库来精确计算
		Bitrate:    int(float64(fileInfo.Size()*8) / 1000), // 粗略估算，单位为 kbps
		SampleRate: 0,
		Channels:   0,
		Encoding:   "unknown",
		Artist:     metadata.Artist(),
		Title:      metadata.Title(),
		Album:      metadata.Album(),
		Year:       metadata.Year(),
		Genre:      metadata.Genre(),
	}

	// 尝试更精确地识别编码格式
	fileExt := strings.ToLower(filepath.Ext(filePath))
	switch fileExt {
	case ".mp3":
		result.Encoding = "MPEG Audio Layer 3"
		// 对于 MP3，如果我们能解析头部信息，可以得到更准确的比特率和采样率
		// 这里暂时使用估算值
	case ".wav":
		result.Encoding = "Waveform Audio File Format"
	case ".flac":
		result.Encoding = "Free Lossless Audio Codec"
	case ".aac":
		result.Encoding = "Advanced Audio Coding"
	case ".m4a":
		result.Encoding = "MPEG-4 Audio"
	case ".ogg":
		result.Encoding = "Ogg Vorbis"
	case ".wma":
		result.Encoding = "Windows Media Audio"
	}

	return result, nil
}

// getMimeType 根据文件扩展名获取MIME类型
func (uc *mediaLibraryAudioUsecase) getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".m4a":
		return "audio/mp4"
	case ".ogg":
		return "audio/ogg"
	case ".wma":
		return "audio/x-ms-wma"
	default:
		return "application/octet-stream"
	}
}

// getFileExtension 获取文件扩展名
func (uc *mediaLibraryAudioUsecase) getFileExtension(filename string) string {
	return strings.TrimPrefix(filepath.Ext(filename), ".")
}

// DeleteAudioFile 删除音频文件
func (uc *mediaLibraryAudioUsecase) DeleteAudioFile(ctx context.Context, fileIDStr string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	fileID, err := primitive.ObjectIDFromHex(fileIDStr)
	if err != nil {
		return fmt.Errorf("invalid file ID: %v", err)
	}

	// 先获取文件信息以便删除物理文件
	audioFile, err := uc.mediaLibraryAudioRepo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to find audio file: %v", err)
	}

	// 从数据库删除记录
	if err := uc.mediaLibraryAudioRepo.DeleteByID(ctx, fileID); err != nil {
		return fmt.Errorf("failed to delete audio file from database: %v", err)
	}

	// 删除物理文件
	if audioFile.StoragePath != "" {
		if err := os.Remove(audioFile.StoragePath); err != nil {
			log.Printf("Warning: failed to remove physical file: %v", err)
		}
	}

	return nil
}

// UploadAudioFileChunked 实现分块上传功能
func (uc *mediaLibraryAudioUsecase) UploadAudioFileChunked(ctx context.Context, fileData io.Reader, fileName, libraryIDStr, uploaderIDStr, uploadID string, chunkIndex, totalChunks int, isLastChunk bool) (*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 验证输入参数
	if fileName == "" {
		return nil, fmt.Errorf("file name cannot be empty")
	}
	if uploadID == "" {
		return nil, fmt.Errorf("upload ID cannot be empty")
	}

	// 将字符串ID转换为ObjectID
	libraryID, err := primitive.ObjectIDFromHex(libraryIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid library ID: %v", err)
	}

	uploaderID, err := primitive.ObjectIDFromHex(uploaderIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid uploader ID: %v", err)
	}

	// 检查是否存在分块上传会话
	session, err := uc.getOrCreateChunkedUploadSession(ctx, uploadID, fileName, libraryID, uploaderID, totalChunks)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create chunked upload session: %v", err)
	}

	// 确保上传会话处于正确的状态
	if session.Status != "pending" && session.Status != "uploading" {
		return nil, fmt.Errorf("upload session is not in a valid state for uploading: %s", session.Status)
	}

	// 更新会话状态为上传中（如果是第一个分块）
	if chunkIndex == 0 && session.Status == "pending" {
		session.Status = "uploading"
		if updateErr := uc.updateChunkedUploadSessionStatus(ctx, session.ID, "uploading"); updateErr != nil {
			return nil, fmt.Errorf("failed to update session status: %v", updateErr)
		}
	}

	// 创建临时文件来保存当前分块数据
	tempDir := os.TempDir()
	chunkFileName := fmt.Sprintf("%s_%d.chunk", uploadID, chunkIndex)
	chunkFilePath := filepath.Join(tempDir, chunkFileName)

	// 创建或追加到临时文件
	chunkFile, err := os.OpenFile(chunkFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk file: %v", err)
	}
	defer chunkFile.Close()

	// 写入当前分块数据
	chunkSize, err := io.Copy(chunkFile, fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to write chunk data: %v", err)
	}

	// 更新已上传分块数
	session.UploadedChunks++
	if updateErr := uc.updateChunkedUploadSessionUploadedChunks(ctx, session.ID, session.UploadedChunks); updateErr != nil {
		return nil, fmt.Errorf("failed to update uploaded chunks count: %v", updateErr)
	}

	// 更新已上传字节数
	session.UploadedBytes += chunkSize
	if updateErr := uc.updateChunkedUploadSessionUploadedBytes(ctx, session.ID, session.UploadedBytes); updateErr != nil {
		return nil, fmt.Errorf("failed to update uploaded bytes count: %v", updateErr)
	}

	// 如果不是最后一个分块，更新进度并返回当前状态
	if !isLastChunk {
		// 更新会话进度 - 基于已上传字节数与总字节数的比例
		progress := float64(session.UploadedBytes) / float64(session.FileSize) * 100
		if updateErr := uc.updateChunkedUploadSessionProgress(ctx, session.ID, progress); updateErr != nil {
			return nil, fmt.Errorf("failed to update session progress: %v", updateErr)
		}

		// 返回当前进度信息
		return &domainCore.MediaLibraryAudioResponse{
			ID:       session.ID,
			FileName: session.FileName,
			Status:   session.Status,
		}, nil
	}

	// 如果是最后一个分块，合并所有分块文件
	if isLastChunk {
		// 合并所有分块文件
		mergedFilePath, err := uc.mergeChunkedFiles(uploadID, totalChunks, fileName)
		if err != nil {
			// 更新会话状态为失败
			uc.updateChunkedUploadSessionStatus(ctx, session.ID, "failed")
			return nil, fmt.Errorf("failed to merge chunked files: %v", err)
		}

		// 计算合并后文件的校验和
		checksum, err := uc.calculateFileChecksum(mergedFilePath)
		if err != nil {
			uc.updateChunkedUploadSessionStatus(ctx, session.ID, "failed")
			return nil, fmt.Errorf("failed to calculate file checksum: %v", err)
		}

		// 检查是否已存在相同校验和的文件
		existingFile, err := uc.mediaLibraryAudioRepo.GetByChecksum(ctx, checksum)
		if err == nil && existingFile != nil {
			// 如果已存在相同的文件，则直接返回该文件信息
			response := &domainCore.MediaLibraryAudioResponse{
				ID:          existingFile.ID,
				FileName:    existingFile.FileName,
				FilePath:    existingFile.FilePath,
				StoragePath: existingFile.StoragePath,
				FileSize:    existingFile.FileSize,
				FileType:    existingFile.FileType,
				Checksum:    existingFile.Checksum,
				LibraryID:   existingFile.LibraryID,
				UploaderID:  existingFile.UploaderID,
				UploadTime:  existingFile.UploadTime,
				CreatedAt:   existingFile.CreatedAt,
				UpdatedAt:   existingFile.UpdatedAt,
			}

			// 更新会话状态为已完成
			uc.updateChunkedUploadSessionStatus(ctx, session.ID, "completed")

			// 即使文件已存在，也触发扫描以确保元数据是最新的
			go uc.triggerScanAfterUpload(existingFile.LibraryID.Hex())

			return response, nil
		}

		// 获取目标存储路径
		storagePath := uc.getStoragePath(fileName, libraryID.Hex())

		// 移动合并后的文件到最终位置
		if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
			uc.updateChunkedUploadSessionStatus(ctx, session.ID, "failed")
			return nil, fmt.Errorf("failed to create storage directory: %v", err)
		}

		if err := os.Rename(mergedFilePath, storagePath); err != nil {
			uc.updateChunkedUploadSessionStatus(ctx, session.ID, "failed")
			return nil, fmt.Errorf("failed to move merged file to storage: %v", err)
		}

		// 准备音频文件实体
		audioFile := &domainCore.MediaLibraryAudio{
			FileName:   fileName,
			FileSize:   session.FileSize,
			Checksum:   checksum,
			LibraryID:  libraryID,
			UploaderID: uploaderID,
			Status:     "ready",
			IsUploaded: true,
		}

		// 验证实体
		if err := audioFile.Validate(); err != nil {
			uc.updateChunkedUploadSessionStatus(ctx, session.ID, "failed")
			return nil, fmt.Errorf("validation failed: %v", err)
		}

		// 保存到数据库
		if err := uc.mediaLibraryAudioRepo.Create(ctx, audioFile); err != nil {
			uc.updateChunkedUploadSessionStatus(ctx, session.ID, "failed")
			return nil, fmt.Errorf("failed to save audio file metadata: %v", err)
		}

		// 更新会话状态为已完成
		uc.updateChunkedUploadSessionStatus(ctx, session.ID, "completed")

		// 触发媒体库扫描
		go uc.triggerScanAfterUpload(libraryID.Hex())

		// 返回响应
		response := &domainCore.MediaLibraryAudioResponse{
			ID:          audioFile.ID,
			FileName:    audioFile.FileName,
			FilePath:    audioFile.FilePath,
			StoragePath: audioFile.StoragePath,
			FileSize:    audioFile.FileSize,
			FileType:    audioFile.FileType,
			Checksum:    audioFile.Checksum,
			LibraryID:   audioFile.LibraryID,
			UploaderID:  audioFile.UploaderID,
			UploadTime:  audioFile.UploadTime,
			CreatedAt:   audioFile.CreatedAt,
			UpdatedAt:   audioFile.UpdatedAt,
		}

		return response, nil
	}

	// 对于中间分块，返回当前进度信息
	return &domainCore.MediaLibraryAudioResponse{
		ID:       session.ID,
		FileName: session.FileName,
		Status:   session.Status,
	}, nil
}

// GetUploadProgress 获取上传进度
func (uc *mediaLibraryAudioUsecase) GetUploadProgress(ctx context.Context, uploadID string) (*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	session, err := uc.chunkedUploadSessionRepo.GetByUploadID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload session: %v", err)
	}

	if session == nil {
		return nil, fmt.Errorf("upload session not found")
	}

	// 计算进度百分比 - 优先使用字节级别的进度（更精确）
	var progress float64
	if session.FileSize > 0 && session.UploadedBytes > 0 {
		// 基于已上传字节数的进度
		progress = float64(session.UploadedBytes) / float64(session.FileSize) * 100
	} else if session.TotalChunks > 0 {
		// 回退到基于分块数的进度
		progress = float64(session.UploadedChunks) / float64(session.TotalChunks) * 100
	}

	// 返回进度信息
	response := &domainCore.MediaLibraryAudioResponse{
		ID:       session.ID,
		FileName: session.FileName,
		Status:   session.Status,
		Progress: progress,
	}

	return response, nil
}

// GetDownloadProgress 获取下载进度
func (uc *mediaLibraryAudioUsecase) GetDownloadProgress(ctx context.Context, fileID string) (*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %v", err)
	}

	// 获取音频文件信息
	audioFile, err := uc.mediaLibraryAudioRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find audio file: %v", err)
	}

	if audioFile == nil {
		return nil, fmt.Errorf("audio file not found")
	}

	// 对于下载，我们假设如果文件存在则下载已完成（100%）
	// 在实际应用中，可能需要更复杂的下载进度跟踪机制
	response := &domainCore.MediaLibraryAudioResponse{
		ID:       audioFile.ID,
		FileName: audioFile.FileName,
		Status:   "ready", // 文件已准备就绪可下载
		Progress: 100.0,   // 假设文件已完全准备好
	}

	return response, nil
}

// GetFileMetadata 获取文件元数据信息
func (uc *mediaLibraryAudioUsecase) GetFileMetadata(ctx context.Context, fileIDStr string) (*domainCore.MediaLibraryAudioResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 将字符串ID转换为ObjectID
	fileID, err := primitive.ObjectIDFromHex(fileIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %v", err)
	}

	// 从数据库获取文件信息
	audioFile, err := uc.mediaLibraryAudioRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find audio file: %v", err)
	}

	// 返回包含详细文件信息的响应
	response := &domainCore.MediaLibraryAudioResponse{
		ID:          audioFile.ID,
		FileName:    audioFile.FileName,
		FilePath:    audioFile.FilePath,
		StoragePath: audioFile.StoragePath,
		FileSize:    audioFile.FileSize,
		FileType:    audioFile.FileType,
		Checksum:    audioFile.Checksum,
		LibraryID:   audioFile.LibraryID,
		UploaderID:  audioFile.UploaderID,
		UploadTime:  audioFile.UploadTime,
		CreatedAt:   audioFile.CreatedAt,
		UpdatedAt:   audioFile.UpdatedAt,
	}

	return response, nil
}

// calculateFileChecksum 计算文件的MD5校验和
func (uc *mediaLibraryAudioUsecase) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getStoragePath 生成文件存储路径
func (uc *mediaLibraryAudioUsecase) getStoragePath(fileName, libraryID string) string {
	// 根据库ID和文件名生成存储路径
	// 可以根据需要调整路径结构
	storageBasePath := "./uploads/media_library"
	return filepath.Join(storageBasePath, libraryID, fileName)
}

// triggerScanAfterUpload 上传完成后触发媒体库扫描
func (uc *mediaLibraryAudioUsecase) triggerScanAfterUpload(libraryID string) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), uc.timeout)
	defer cancel()

	// 调用基础文件用例来触发扫描模式0（扫描新的和有修改的文件）
	// 由于ProcessDirectory方法需要目录路径，我们先尝试获取所有媒体库目录
	// 或者可以传递空数组让其扫描所有目录
	dirPaths := []string{} // 空数组将扫描所有媒体库目录
	folderType := 1        // 假设类型1代表音频文件
	scanModel := 0         // 扫描模式0：扫描新的和有修改的文件

	err := uc.basicFileUsecase.ProcessDirectory(ctx, dirPaths, folderType, scanModel)
	if err != nil {
		log.Printf("Warning: failed to trigger media library scan after upload: %v", err)
	} else {
		log.Printf("Successfully triggered media library scan (mode 0) after file upload for library: %s", libraryID)
	}
}

// getOrCreateChunkedUploadSession 获取或创建分块上传会话
func (uc *mediaLibraryAudioUsecase) getOrCreateChunkedUploadSession(ctx context.Context, uploadID, fileName string, libraryID, uploaderID primitive.ObjectID, totalChunks int) (*domainCore.ChunkedUploadSession, error) {
	session, err := uc.chunkedUploadSessionRepo.GetByUploadID(ctx, uploadID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		// 创建新的上传会话
		session = &domainCore.ChunkedUploadSession{
			UploadID:       uploadID,
			FileName:       fileName,
			TotalChunks:    totalChunks,
			UploadedChunks: 0,
			Status:         "pending",
			LibraryID:      libraryID,
			UploaderID:     uploaderID,
		}
		session.SetTimestamps()

		if err := uc.chunkedUploadSessionRepo.Create(ctx, session); err != nil {
			return nil, err
		}
	}

	return session, nil
}

// updateChunkedUploadSessionStatus 更新分块上传会话状态
func (uc *mediaLibraryAudioUsecase) updateChunkedUploadSessionStatus(ctx context.Context, sessionID primitive.ObjectID, status string) error {
	return uc.chunkedUploadSessionRepo.UpdateStatus(ctx, sessionID, status)
}

// updateChunkedUploadSessionUploadedChunks 更新已上传分块数
func (uc *mediaLibraryAudioUsecase) updateChunkedUploadSessionUploadedChunks(ctx context.Context, sessionID primitive.ObjectID, uploadedChunks int) error {
	return uc.chunkedUploadSessionRepo.UpdateUploadedChunks(ctx, sessionID, uploadedChunks)
}

func (uc *mediaLibraryAudioUsecase) updateChunkedUploadSessionProgress(ctx context.Context, sessionID primitive.ObjectID, progress float64) error {
	return uc.chunkedUploadSessionRepo.UpdateProgress(ctx, sessionID, progress)
}

// updateChunkedUploadSessionUploadedBytes 更新已上传字节数
func (uc *mediaLibraryAudioUsecase) updateChunkedUploadSessionUploadedBytes(ctx context.Context, sessionID primitive.ObjectID, uploadedBytes int64) error {
	return uc.chunkedUploadSessionRepo.UpdateUploadedBytes(ctx, sessionID, uploadedBytes)
}

// mergeChunkedFiles 合并分块文件
func (uc *mediaLibraryAudioUsecase) mergeChunkedFiles(uploadID string, totalChunks int, fileName string) (string, error) {
	tempDir := os.TempDir()
	mergedFilePath := filepath.Join(tempDir, fmt.Sprintf("%s_merged_%s", uploadID, fileName))

	// 创建合并文件
	mergedFile, err := os.Create(mergedFilePath)
	if err != nil {
		return "", err
	}
	defer mergedFile.Close()

	// 按顺序合并所有分块
	var totalSize int64
	for i := 0; i < totalChunks; i++ {
		chunkFileName := fmt.Sprintf("%s_%d.chunk", uploadID, i)
		chunkFilePath := filepath.Join(tempDir, chunkFileName)

		chunkFile, err := os.Open(chunkFilePath)
		if err != nil {
			return "", err
		}

		chunkSize, err := io.Copy(mergedFile, chunkFile)
		chunkFile.Close()

		if err != nil {
			return "", err
		}

		totalSize += chunkSize

		// 删除已处理的分块文件
		os.Remove(chunkFilePath)
	}

	// 更新会话中的文件大小
	// 注意：这里不能直接更新会话，因为这发生在不同的上下文中
	// 我们会在完成合并后更新会话信息

	return mergedFilePath, nil
}

// ProgressReader 包装io.Reader以跟踪读取进度
type ProgressReader struct {
	reader           io.Reader
	totalSize        int64
	currentSize      int64
	progressCallback domainCore.ProgressCallback
}

// Read 实现io.Reader接口
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.currentSize += int64(n)

	// 计算进度百分比
	if pr.totalSize > 0 {
		progress := float64(pr.currentSize) / float64(pr.totalSize) * 100
		if pr.progressCallback != nil {
			pr.progressCallback(progress, "downloading", fmt.Sprintf("Downloading... (%d/%d bytes)", pr.currentSize, pr.totalSize))
		}
	}

	return n, err
}
