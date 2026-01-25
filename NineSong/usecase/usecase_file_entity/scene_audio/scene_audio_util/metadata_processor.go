package usercase_audio_util

import (
	"fmt"
	"path/filepath"
	"strings"
)

// 支持的歌词文件格式
var SupportedLyricsFormats = []string{".lrc", ".txt", ".krc", ".qrc"}

// 支持的封面图片格式
var SupportedCoverFormats = []string{".jpg", ".jpeg", ".png"}

// IsValidLyricsFormat 检查是否为支持的歌词文件格式
func IsValidLyricsFormat(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, format := range SupportedLyricsFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// GetLyricsFormat 获取歌词文件格式
func GetLyricsFormat(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, format := range SupportedLyricsFormats {
		if ext == format {
			return format
		}
	}
	return ""
}

// IsValidCoverFormat 检查是否为支持的封面图片格式
func IsValidCoverFormat(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, format := range SupportedCoverFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// GetCoverFormat 获取封面图片格式
func GetCoverFormat(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, format := range SupportedCoverFormats {
		if ext == format {
			return format
		}
	}
	return ""
}

// GetLyricsMimeType 根据文件扩展名获取歌词文件的MIME类型
func GetLyricsMimeType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".lrc":
		return "text/plain; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".krc":
		return "application/x-krc" // 酷狗歌词格式
	case ".qrc":
		return "application/x-qrc" // QQ音乐歌词格式
	default:
		return "text/plain; charset=utf-8"
	}
}

// GetCoverMimeType 根据文件扩展名获取封面图片的MIME类型
func GetCoverMimeType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	default:
		return "image/jpeg"
	}
}

// GenerateLyricsFileName 生成歌词文件名
func GenerateLyricsFileName(baseName, format string) string {
	if format == "" {
		format = ".lrc" // 默认格式
	}
	// 确保格式以点开头
	if !strings.HasPrefix(format, ".") {
		format = "." + format
	}
	return fmt.Sprintf("%s%s", baseName, format)
}

// GenerateCoverFileName 生成封面文件名
func GenerateCoverFileName(baseName, size, format string) string {
	if format == "" {
		format = ".jpg" // 默认格式
	}
	// 确保格式以点开头
	if !strings.HasPrefix(format, ".") {
		format = "." + format
	}
	if size != "" {
		return fmt.Sprintf("%s_%s%s", baseName, size, format)
	}
	return fmt.Sprintf("%s%s", baseName, format)
}

// ValidateLyricsFile 验证歌词文件
func ValidateLyricsFile(fileName string) error {
	if !IsValidLyricsFormat(fileName) {
		return fmt.Errorf("不支持的歌词文件格式，支持的格式: %v", SupportedLyricsFormats)
	}
	return nil
}

// ValidateCoverFile 验证封面文件
func ValidateCoverFile(fileName string) error {
	if !IsValidCoverFormat(fileName) {
		return fmt.Errorf("不支持的封面文件格式，支持的格式: %v", SupportedCoverFormats)
	}
	return nil
}
