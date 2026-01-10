package scene_audio

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
)

type AudioCommonUtil struct {
	detector domain_file_entity.FileDetector
}

func NewAudioCommonUtil(detector domain_file_entity.FileDetector) *AudioCommonUtil {
	return &AudioCommonUtil{detector: detector}
}

func (util *AudioCommonUtil) ShouldProcess(path string, folderType int) bool {
	ext := strings.ToLower(filepath.Ext(path))

	audioExts := map[string]bool{".mp3": true, ".flac": true, ".wav": true, ".ape": true, ".m4a": true, ".ogg": true, ".aac": true, ".wma": true, ".opus": true, ".alac": true}
	videoExts := map[string]bool{".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true}
	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true, ".tiff": true}
	documentExts := map[string]bool{".txt": true, ".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".md": true}

	switch folderType {
	case 1:
		if audioExts[ext] {
			return true
		}
	case 2:
		if videoExts[ext] {
			return true
		}
	case 3:
		if imageExts[ext] {
			return true
		}
	case 4:
		if documentExts[ext] {
			return true
		}
	}

	if util.detector == nil {
		return false
	}

	fileType, err := util.detector.DetectMediaType(path)
	if err != nil {
		log.Printf("文件类型检测失败: %v", err)
		return false
	}
	if folderType == 1 && fileType == domain_file_entity.Audio {
		return true
	}
	if folderType == 2 && fileType == domain_file_entity.Video {
		return true
	}
	if folderType == 3 && fileType == domain_file_entity.Image {
		return true
	}
	if folderType == 4 && fileType == domain_file_entity.Document {
		return true
	}
	return false
}

func (util *AudioCommonUtil) FastCountFilesInFolder(rootPath string, folderType int) (int, error) {
	count := 0
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))

			switch folderType {
			case 1:
				audioExts := map[string]bool{".mp3": true, ".flac": true, ".wav": true, ".ape": true, ".m4a": true, ".ogg": true, ".aac": true, ".wma": true, ".opus": true, ".alac": true}
				if audioExts[ext] {
					count++
				}
			case 2:
				videoExts := map[string]bool{".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true}
				if videoExts[ext] {
					count++
				}
			case 4:
				documentExts := map[string]bool{".txt": true, ".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".md": true}
				if documentExts[ext] {
					count++
				}
			}
		}
		return nil
	})
	return count, err
}

func (util *AudioCommonUtil) IsPathInLibraries(path string, libraryFolders []*domain_file_entity.LibraryFolderMetadata) bool {
	for _, folder := range libraryFolders {
		libraryFolderPath := strings.Replace(folder.FolderPath, "/", "\\", -1)
		if !strings.HasSuffix(libraryFolderPath, "\\") {
			libraryFolderPath += "\\"
		}
		if strings.HasPrefix(strings.Replace(path, "/", "\\", -1), libraryFolderPath) {
			return true
		}
	}
	return false
}

func (util *AudioCommonUtil) CollectLibraryPaths(libraryFolders []*domain_file_entity.LibraryFolderMetadata) []string {
	paths := make([]string, 0, len(libraryFolders))
	for _, folder := range libraryFolders {
		paths = append(paths, folder.FolderPath)
	}
	return paths
}

func (util *AudioCommonUtil) WalkLibraryFolders(
	libraryFolders []*domain_file_entity.LibraryFolderMetadata,
	processFile func(path string, info os.FileInfo) error,
) error {
	for _, folder := range libraryFolders {
		err := filepath.Walk(folder.FolderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("访问路径 %s 出错: %v", path, err)
				return nil
			}
			if info.IsDir() {
				return nil
			}
			return processFile(path, info)
		})
		if err != nil {
			log.Printf("文件夹%v，文件遍历错误: %v，请检查该文件夹是否存在", folder.FolderPath, err)
		}
	}
	return nil
}

func (util *AudioCommonUtil) CleanupOldFolders(lyricsTempPath string, keepCount int) error {
	entries, err := os.ReadDir(lyricsTempPath)
	if err != nil {
		return err
	}

	var folders []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && IsTimestampFolder(entry.Name()) {
			folders = append(folders, entry)
		}
	}

	if len(folders) <= keepCount {
		return nil
	}

	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name() < folders[j].Name()
	})

	deleteCount := len(folders) - keepCount
	for i := 0; i < deleteCount; i++ {
		folderPath := filepath.Join(lyricsTempPath, folders[i].Name())
		if err := os.RemoveAll(folderPath); err != nil {
			log.Printf("删除失败 %s: %v", folderPath, err)
		}
	}

	return nil
}

func IsTimestampFolder(name string) bool {
	if len(name) != 14 {
		return false
	}
	for _, c := range name {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
