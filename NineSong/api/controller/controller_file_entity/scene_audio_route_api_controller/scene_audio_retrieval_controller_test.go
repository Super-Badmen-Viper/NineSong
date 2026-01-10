package scene_audio_route_api_controller

//package scene_audio_route_api_controller
//
//import (
//	"encoding/json"
//	"errors"
//	"fmt"
//	"io"
//	"log"
//	"math"
//	"net/http"
//	"os"
//	"path/filepath"
//	"strconv"
//	"strings"
//	"sync"
//	"syscall"
//	"time"
//
//	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
//	"github.com/gin-gonic/gin"
//	ffmpeggo "github.com/u2takey/ffmpeg-go"
//)
//
//// 解码信息结构体，用于存储FFmpeg解码输出信息
//type DecodeInfo struct {
//	MediaFileID  string            `json:"media_file_id"`
//	Title        string            `json:"title"`
//	Artist       string            `json:"artist"`
//	Album        string            `json:"album"`
//	Duration     float64           `json:"duration"`
//	Codec        string            `json:"codec"`
//	SampleRate   int               `json:"sample_rate"`
//	Bitrate      int               `json:"bitrate"`
//	Lyrics       string            `json:"lyrics"`
//	Transcoding  bool              `json:"transcoding"`
//	LastUpdated  time.Time         `json:"last_updated"`
//	FFmpegOutput string            `json:"ffmpeg_output"`
//	Metadata     map[string]string `json:"metadata"`
//}
//
//// 全局变量存储当前解码信息
//var (
//	currentDecodeInfo DecodeInfo
//	decodeInfoMutex   sync.RWMutex
//)
//
//type RetrievalController struct {
//	RetrievalUsecase scene_audio_route_interface.RetrievalRepository
//}
//
//func NewRetrievalController(uc scene_audio_route_interface.RetrievalRepository) *RetrievalController {
//	return &RetrievalController{RetrievalUsecase: uc}
//}
//
//func (c *RetrievalController) FixedStreamHandler(ctx *gin.Context) {
//	var req struct {
//		MediaFileID       string `form:"media_file_id" binding:"required"`
//		PlayComponentType string `form:"play_component_type"`
//		CueModel          bool   `form:"cue_model"`
//	}
//
//	if err := ctx.ShouldBind(&req); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"code":    "INVALID_PARAMETERS",
//			"message": "缺少必要参数: media_file_id",
//		})
//		return
//	}
//
//	if len(req.PlayComponentType) == 0 {
//		req.PlayComponentType = "web"
//	}
//	if req.PlayComponentType != "web" {
//		req.PlayComponentType = "mpv"
//	}
//
//	filePath, err := c.RetrievalUsecase.GetStreamPath(ctx.Request.Context(), req.MediaFileID, req.CueModel)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{
//			"code":    "RESOURCE_NOT_FOUND",
//			"message": "音频文件不存在",
//		})
//		return
//	}
//	tempSteamFolderPath, _ := c.RetrievalUsecase.GetStreamTempPath(ctx.Request.Context(), "stream")
//	streamMediaFile(ctx, filePath, req.MediaFileID, tempSteamFolderPath, req.PlayComponentType)
//}
//
//func (c *RetrievalController) RealStreamHandler(ctx *gin.Context) {
//	var req struct {
//		MediaFileID       string `form:"media_file_id" binding:"required"`
//		PlayComponentType string `form:"play_component_type"`
//		CueModel          bool   `form:"cue_model"`
//	}
//
//	if err := ctx.ShouldBind(&req); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"code":    "INVALID_PARAMETERS",
//			"message": "缺少必要参数: media_file_id",
//		})
//		return
//	}
//
//	if len(req.PlayComponentType) == 0 {
//		req.PlayComponentType = "web"
//	}
//	if req.PlayComponentType != "web" {
//		req.PlayComponentType = "mpv"
//	}
//
//	filePath, err := c.RetrievalUsecase.GetStreamPath(ctx.Request.Context(), req.MediaFileID, req.CueModel)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{
//			"code":    "RESOURCE_NOT_FOUND",
//			"message": "音频文件不存在",
//		})
//		return
//	}
//	tempSteamFolderPath, _ := c.RetrievalUsecase.GetStreamTempPath(ctx.Request.Context(), "stream")
//	streamMediaFile(ctx, filePath, req.MediaFileID, tempSteamFolderPath, req.PlayComponentType)
//}
//
//func (c *RetrievalController) DownloadHandler(ctx *gin.Context) {
//	var req struct {
//		MediaFileID string `form:"media_file_id" binding:"required"`
//		CueModel    bool   `form:"cue_model"`
//	}
//
//	if err := ctx.ShouldBind(&req); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"code":    "INVALID_PARAMETERS",
//			"message": "缺少必要参数: media_file_id",
//		})
//		return
//	}
//
//	filePath, err := c.RetrievalUsecase.GetStreamPath(ctx.Request.Context(), req.MediaFileID, req.CueModel)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{
//			"code":    "RESOURCE_NOT_FOUND",
//			"message": "音频文件不存在",
//		})
//		return
//	}
//	tempSteamFolderPath, _ := c.RetrievalUsecase.GetStreamTempPath(ctx.Request.Context(), "stream")
//	streamMediaFile(ctx, filePath, req.MediaFileID, tempSteamFolderPath, "")
//}
//
//func (c *RetrievalController) CoverArtIDHandler(ctx *gin.Context) {
//	var req struct {
//		Type     string `form:"type" binding:"required,oneof=media album artist"`
//		TargetID string `form:"target_id" binding:"required,hexadecimal,len=24"`
//	}
//
//	if err := ctx.ShouldBind(&req); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"code":    "INVALID_PARAMETERS",
//			"message": "参数格式错误: type必须为media或album，target_id必须为24位十六进制",
//		})
//		return
//	}
//
//	filePath, err := c.RetrievalUsecase.GetCoverArtID(ctx.Request.Context(), req.Type, req.TargetID)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{
//			"code":    "COVER_NOT_FOUND",
//			"message": "封面文件不存在",
//		})
//		return
//	}
//
//	ctx.Header("Content-Type", "image/jpeg")
//	ctx.File(filePath)
//}
//
//func (c *RetrievalController) CoverArtPathHandler(ctx *gin.Context) {
//	var req struct {
//		Type     string `form:"type" binding:"required,oneof=back cover disc"`
//		TargetID string `form:"target_id" binding:"required,hexadecimal,len=24"`
//	}
//
//	if err := ctx.ShouldBind(&req); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"code":    "INVALID_PARAMETERS",
//			"message": "参数格式错误: type必须为media或album，target_id必须为24位十六进制",
//		})
//		return
//	}
//
//	filePath, err := c.RetrievalUsecase.GetCoverArtID(ctx.Request.Context(), req.Type, req.TargetID)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{
//			"code":    "COVER_NOT_FOUND",
//			"message": "封面文件不存在",
//		})
//		return
//	}
//
//	ctx.Header("Content-Type", "image/jpeg")
//	ctx.File(filePath)
//}
//
//func (c *RetrievalController) LyricsHandlerMetadata(ctx *gin.Context) {
//	var req struct {
//		MediaFileID string `form:"media_file_id"`
//		Artist      string `form:"artist"`
//		Title       string `form:"title"`
//		FileType    string `form:"file_type"`
//	}
//
//	if err := ctx.ShouldBind(&req); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"code":    "INVALID_PARAMETERS",
//			"message": "参数格式错误",
//		})
//		return
//	}
//
//	lyricsContent, err := c.RetrievalUsecase.GetLyricsLrcMetaData(
//		ctx.Request.Context(), req.MediaFileID, req.Artist, req.Title, req.FileType,
//	)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{
//			"code":    "LYRICS_NOT_FOUND",
//			"message": "未找到关联的歌词内容",
//		})
//		return
//	}
//
//	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(lyricsContent))
//}
//func (c *RetrievalController) LyricsHandlerFile(ctx *gin.Context) {
//	var req struct {
//		MediaFileID string `form:"media_file_id" binding:"required"`
//	}
//
//	if err := ctx.ShouldBind(&req); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"code":    "INVALID_PARAMETERS",
//			"message": "缺少必要参数: media_file_id",
//		})
//		return
//	}
//
//	filePath, err := c.RetrievalUsecase.GetLyricsLrcMetaData(ctx.Request.Context(), req.MediaFileID, "", "", "")
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{
//			"code":    "RESOURCE_NOT_FOUND",
//			"message": "歌词文件不存在",
//		})
//		return
//	}
//	serveTextFile(ctx, filePath)
//}
//
//// GET接口：获取当前解码信息
//func (c *RetrievalController) GetDecodeInfoHandler(ctx *gin.Context) {
//	decodeInfoMutex.RLock()
//	info := currentDecodeInfo
//	decodeInfoMutex.RUnlock()
//
//	ctx.JSON(http.StatusOK, info)
//}
//
//// 更新解码信息
//func updateDecodeInfo(mediaFileID string, path string, lyrics string, ffmpegOutput string) {
//	// 使用ffmpeg探测文件元数据
//	data, err := ffmpeggo.Probe(path)
//	if err != nil {
//		log.Printf("文件探测失败: %v", err)
//		return
//	}
//
//	// 解析元数据
//	var probeData map[string]interface{}
//	if err := json.Unmarshal([]byte(data), &probeData); err != nil {
//		log.Printf("JSON解析失败: %v", err)
//		return
//	}
//
//	// 提取基本信息
//	format := probeData["format"].(map[string]interface{})
//	duration, _ := strconv.ParseFloat(fmt.Sprintf("%v", format["duration"]), 64)
//	bitrate, _ := strconv.Atoi(fmt.Sprintf("%v", format["bit_rate"]))
//	bitrate = bitrate / 1000 // 转换为kbps
//
//	// 提取流信息
//	streams := probeData["streams"].([]interface{})
//	codec := ""
//	sampleRate := 0
//	for _, stream := range streams {
//		streamMap := stream.(map[string]interface{})
//		if streamMap["codec_type"] == "audio" {
//			codec = fmt.Sprintf("%v", streamMap["codec_name"])
//			sampleRate, _ = strconv.Atoi(fmt.Sprintf("%v", streamMap["sample_rate"]))
//			break
//		}
//	}
//
//	// 提取元数据
//	metadata := make(map[string]string)
//	if tags, ok := format["tags"].(map[string]interface{}); ok {
//		for k, v := range tags {
//			metadata[k] = fmt.Sprintf("%v", v)
//		}
//	}
//
//	// 构建解码信息
//	decodeInfo := DecodeInfo{
//		MediaFileID:  mediaFileID,
//		Title:        metadata["title"],
//		Artist:       metadata["artist"],
//		Album:        metadata["album"],
//		Duration:     duration,
//		Codec:        codec,
//		SampleRate:   sampleRate,
//		Bitrate:      bitrate,
//		Lyrics:       lyrics,
//		Transcoding:  true,
//		LastUpdated:  time.Now(),
//		FFmpegOutput: ffmpegOutput,
//		Metadata:     metadata,
//	}
//
//	// 更新全局解码信息
//	decodeInfoMutex.Lock()
//	currentDecodeInfo = decodeInfo
//	decodeInfoMutex.Unlock()
//}
//
//// 直接播放文件，支持范围请求
//func directPlay(ctx *gin.Context, path string) {
//	// 增加范围请求支持
//	file, err := os.Open(path)
//	if err != nil {
//		handleFileError(ctx, path, err)
//		return
//	}
//	defer func(file *os.File) {
//		err := file.Close()
//		if err != nil {
//			ctx.JSON(http.StatusInternalServerError, gin.H{
//				"code":    "FILE_CLOSE_ERROR",
//				"message": "关闭文件时发生错误",
//			})
//			return
//		}
//	}(file)
//
//	fileInfo, _ := file.Stat()
//
//	// 设置正确的内容长度
//	ctx.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
//
//	// 支持范围请求
//	ctx.Header("Accept-Ranges", "bytes")
//
//	// 设置缓存控制
//	ctx.Header("Cache-Control", "public, max-age=86400") // 24小时缓存
//
//	// 对于AAC音频文件，确保Content-Type为audio/aac
//	contentType := detectContentType(path)
//	if strings.Contains(strings.ToLower(path), ".aac") || needTranscode(path) {
//		contentType = "audio/aac"
//	}
//	ctx.Header("Content-Type", contentType)
//
//	// 支持直接文件服务
//	ctx.File(path)
//}
//func detectContentType(path string) string {
//	ext := filepath.Ext(path)
//	switch ext {
//	case ".jpg", ".jpeg":
//		return "image/jpeg"
//	case ".png":
//		return "image/png"
//	case ".mp3":
//		return "audio/mpeg"
//	case ".flac":
//		return "audio/flac"
//	case ".wav":
//		return "audio/wav"
//	case ".aac":
//		return "audio/aac"
//	case ".m4a":
//		return "audio/mp4" // M4A文件的标准MIME类型
//	case ".lrc":
//		return "text/plain; charset=utf-8"
//	default:
//		return "application/octet-stream"
//	}
//}
//func handleFileError(ctx *gin.Context, path string, err error) {
//	var pathErr *os.PathError
//	var sysErr syscall.Errno
//
//	switch {
//	case errors.As(err, &pathErr) && os.IsNotExist(pathErr.Err):
//		ctx.JSON(404, gin.H{
//			"code":    "FILE_NOT_FOUND",
//			"message": "文件不存在: " + path,
//			"detail":  pathErr.Error(),
//		})
//	case errors.As(err, &sysErr) && errors.Is(sysErr, syscall.EACCES):
//		ctx.JSON(403, gin.H{
//			"code":    "PERMISSION_DENIED",
//			"message": "无访问权限: " + path,
//		})
//	default:
//		ctx.JSON(500, gin.H{
//			"code":    "SERVER_ERROR",
//			"message": "系统错误: " + err.Error(),
//		})
//	}
//}
//
//// 文本文件服务
//func serveTextFile(ctx *gin.Context, path string) {
//	ctx.Header("Content-Type", "text/plain; charset=utf-8")
//	ctx.File(path)
//}
//
//// 检测m4a文件是否为AAC编码
//func isAACEncodedM4A(path string) bool {
//	if strings.ToLower(filepath.Ext(path)) != ".m4a" {
//		return false
//	}
//
//	// 使用FFmpeg探测文件编码
//	data, err := ffmpeggo.Probe(path)
//	if err != nil {
//		log.Printf("文件探测失败: %v", err)
//		// 如果探测失败，我们基于文件扩展名做保守判断
//		// 大多数M4A文件是AAC编码的，所以默认返回true
//		return true
//	}
//
//	// 解析JSON数据检测编码格式
//	type Stream struct {
//		CodecName string `json:"codec_name"`
//		CodecType string `json:"codec_type"`
//	}
//	type ProbeData struct {
//		Streams []Stream `json:"streams"`
//	}
//
//	var probe ProbeData
//	if err := json.Unmarshal([]byte(data), &probe); err != nil {
//		log.Printf("JSON解析失败: %v", err)
//		// 如果解析失败，返回true作为默认值
//		return true
//	}
//
//	for _, stream := range probe.Streams {
//		if stream.CodecType == "audio" && stream.CodecName == "aac" {
//			return true
//		}
//	}
//	return false
//}
//
//// 检测是否需要转码（非howler.js支持的高效音频格式需要转码）
//func needTranscode(path string) bool {
//	// 获取文件扩展名并转换为小写
//	ext := strings.ToLower(filepath.Ext(path))
//
//	// 特殊处理m4a格式：检查是否为AAC编码
//	if ext == ".m4a" {
//		return !isAACEncodedM4A(path)
//	}
//
//	// howler.js支持的高效音频格式列表，不需要转码
//	// 这些格式在大多数浏览器中都有良好支持
//	allowedFormats := map[string]bool{
//		".mp3":  true, // 广泛支持
//		".flac": true, // 无损格式，支持良好
//		".wav":  true, // 无损格式，支持良好
//		".aac":  true, // 高效格式，支持良好
//	}
//
//	// 检查扩展名是否在允许列表中
//	return !allowedFormats[ext]
//}
//
//// 生成缓存文件路径
//func getTranscodeCachePath(tempSteamFolderPath string, mediaFileID string, path string) string {
//	// 使用媒体文件ID和源文件的修改时间生成缓存文件名
//	fileInfo, err := os.Stat(path)
//	if err != nil {
//		// 如果无法获取文件信息，只使用媒体文件ID
//		return filepath.Join(tempSteamFolderPath, mediaFileID+"_transcoded.aac")
//	}
//
//	// 使用源文件路径和修改时间生成缓存文件名
//	mtime := fileInfo.ModTime().Unix()
//	baseName := filepath.Base(path)
//	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
//	cacheFileName := fmt.Sprintf("%s_%s_%d_transcoded.aac", mediaFileID, nameWithoutExt, mtime)
//
//	return filepath.Join(tempSteamFolderPath, cacheFileName)
//}
//
//// 检查缓存文件是否有效
//func isCacheValid(cachePath string, sourcePath string) bool {
//	// 检查缓存文件是否存在
//	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
//		return false
//	}
//
//	// 检查源文件和缓存文件的修改时间
//	sourceInfo, err := os.Stat(sourcePath)
//	if err != nil {
//		return false
//	}
//
//	cacheInfo, err := os.Stat(cachePath)
//	if err != nil {
//		return false
//	}
//
//	// 如果源文件比缓存文件更新，则缓存无效
//	return sourceInfo.ModTime().Before(cacheInfo.ModTime())
//}
//
//// 统一流处理函数
//func streamMediaFile(ctx *gin.Context, path string, mediaFileID string, tempSteamFolderPath string, playComponentType string) {
//	// 从控制器获取RetrievalUsecase
//	controller, exists := ctx.Get("controller")
//	lyrics := ""
//	if exists {
//		if retrievalController, ok := controller.(*RetrievalController); ok && mediaFileID != "" {
//			lyricsContent, _ := retrievalController.RetrievalUsecase.GetLyricsLrcMetaData(ctx, mediaFileID, "", "", "")
//			lyrics = lyricsContent
//		}
//	}
//
//	// 更新解码信息
//	updateDecodeInfo(mediaFileID, path, lyrics, "")
//
//	// 检测并转码非常规音频格式
//	if playComponentType == "web" {
//		if needTranscode(path) {
//			// 使用FFmpeg探测音频文件的真实时长和比特率
//			data, err := ffmpeggo.Probe(path)
//			if err != nil {
//				log.Printf("文件探测失败: %v", err)
//				directPlay(ctx, path)
//				return
//			}
//
//			// 解析JSON数据获取音频信息
//			var probeData map[string]interface{}
//			if err := json.Unmarshal([]byte(data), &probeData); err != nil {
//				log.Printf("JSON解析失败: %v", err)
//				directPlay(ctx, path)
//				return
//			}
//
//			// 提取基本信息
//			format := probeData["format"].(map[string]interface{})
//			duration, _ := strconv.ParseFloat(fmt.Sprintf("%v", format["duration"]), 64)
//			bitrate, _ := strconv.Atoi(fmt.Sprintf("%v", format["bit_rate"]))
//			if bitrate == 0 {
//				// 如果无法获取比特率，使用默认值
//				bitrate = 2400000 // 2400 kbps 默认值
//			}
//
//			// 检查是否有Range请求
//			rangeHeader := ctx.GetHeader("Range")
//
//			// 如果没有Range请求，尝试使用缓存
//			if rangeHeader == "" {
//				cachePath := getTranscodeCachePath(tempSteamFolderPath, mediaFileID, path)
//
//				// 检查缓存是否有效
//				if isCacheValid(cachePath, path) {
//					// 使用缓存文件
//					ctx.Header("Content-Type", "audio/aac")
//					ctx.Header("Accept-Ranges", "bytes")
//					ctx.Header("Cache-Control", "public, max-age=3600") // 1小时缓存
//					directPlay(ctx, cachePath)
//					return
//				}
//			}
//
//			// 修复：重新计算虚拟大小，确保基于实际比特率
//			targetBitrate := 128000
//			virtualSize := int64(duration * float64(targetBitrate) / 8)
//
//			// 修复：确保virtualSize至少为1，避免除零错误
//			if virtualSize <= 0 {
//				fileInfo, err := os.Stat(path)
//				if err != nil {
//					log.Printf("获取文件信息失败: %v", err)
//					directPlay(ctx, path)
//					return
//				}
//				virtualSize = fileInfo.Size()
//			}
//
//			// 修复范围请求处理
//			startPos := int64(0)
//			endPos := virtualSize - 1
//			statusCode := http.StatusOK
//			var requestedDuration float64
//
//			if rangeHeader != "" {
//				if strings.HasPrefix(rangeHeader, "bytes=") {
//					rangeStr := rangeHeader[6:]
//					parts := strings.Split(rangeStr, "-")
//					if len(parts) == 2 {
//						if parts[0] != "" {
//							if start, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
//								startPos = start
//							}
//						}
//						if parts[1] != "" {
//							if end, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
//								endPos = end
//							}
//						}
//					}
//					statusCode = http.StatusPartialContent
//					ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startPos, endPos, virtualSize))
//
//					// 修复：计算实际请求的时长
//					contentLength := endPos - startPos + 1
//					requestedDuration = float64(contentLength) * 8 / float64(targetBitrate)
//				}
//			} else {
//				// 没有Range请求时，请求整个文件
//				requestedDuration = duration
//			}
//
//			// 修复：设置更合理的响应头
//			ctx.Header("Content-Type", "audio/aac")
//			ctx.Header("Accept-Ranges", "bytes")
//			ctx.Header("Cache-Control", "no-cache")
//
//			// 修复：对于流式传输，不要设置固定的Content-Length
//			// 让FFmpeg控制输出时长，使用chunked传输
//			ctx.Header("Transfer-Encoding", "chunked")
//
//			// 计算seek时间
//			seekTime := math.Max(0, float64(startPos)/float64(virtualSize)*duration)
//			log.Printf("计算seek时间: startPos=%d, virtualSize=%d, duration=%.2f, seekTime=%.3f, requestedDuration=%.3f",
//				startPos, virtualSize, duration, seekTime, requestedDuration)
//
//			// 设置状态码
//			ctx.Status(statusCode)
//
//			// 修复：创建管道用于流式传输
//			pipeReader, pipeWriter := io.Pipe()
//			defer pipeReader.Close()
//			defer pipeWriter.Close()
//
//			// 修复：改进的FFmpeg参数配置
//			ffmpegArgs := ffmpeggo.KwArgs{
//				"ss": seekTime, // 从指定时间开始
//			}
//
//			// 修复：添加输出时长限制参数
//			if requestedDuration > 0 {
//				ffmpegArgs["t"] = requestedDuration
//			}
//
//			outputArgs := ffmpeggo.KwArgs{
//				"c:a":     "aac",
//				"b:a":     "128k",
//				"ar":      44100,
//				"ac":      2,
//				"f":       "adts",
//				"threads": "0",
//			}
//
//			// 修复：启动FFmpeg转码进程
//			go func() {
//				defer pipeWriter.Close()
//
//				log.Printf("启动FFmpeg转码: seekTime=%.3f, duration=%.3f", seekTime, requestedDuration)
//
//				err := ffmpeggo.Input(path, ffmpegArgs).
//					Output("pipe:1", outputArgs).
//					WithOutput(pipeWriter).
//					WithErrorOutput(os.Stderr). // 输出错误信息到stderr便于调试
//					Run()
//
//				if err != nil {
//					log.Printf("FFmpeg转码失败: %v", err)
//					pipeWriter.CloseWithError(err)
//				} else {
//					log.Printf("FFmpeg转码完成")
//				}
//			}()
//
//			// 修复：改进的数据传输逻辑
//			buffer := make([]byte, 32*1024) // 使用32KB缓冲区
//			totalWritten := int64(0)
//
//			for {
//				select {
//				case <-ctx.Request.Context().Done():
//					// 客户端断开连接
//					log.Printf("客户端连接断开，停止传输")
//					return
//				default:
//					n, err := pipeReader.Read(buffer)
//					if n > 0 {
//						written, writeErr := ctx.Writer.Write(buffer[:n])
//						if flusher, ok := ctx.Writer.(http.Flusher); ok {
//							flusher.Flush() // 立即发送数据
//						}
//						totalWritten += int64(written)
//
//						if writeErr != nil {
//							log.Printf("写入响应失败: %v", writeErr)
//							return
//						}
//					}
//					if err != nil {
//						if err == io.EOF {
//							log.Printf("音频流传输完成: 总字节数=%d", totalWritten)
//						} else {
//							log.Printf("读取管道数据失败: %v", err)
//						}
//						return
//					}
//				}
//			}
//			return
//		}
//	}
//
//	// 不需要转码，直接播放
//	directPlay(ctx, path)
//}
