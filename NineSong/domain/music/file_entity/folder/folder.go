package folder

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/file_entity/folder/core"
)

// 重新导出核心类型
type LibraryStatus = core.LibraryStatus
type FolderType = core.FolderType
type FolderEntry = core.FolderEntry
type LibraryFolderMetadata = core.LibraryFolderMetadata
type FolderRepository = core.FolderRepository

// 重新导出常量
const (
	StatusActive   = core.StatusActive
	StatusDisabled = core.StatusDisabled
	StatusScanning = core.StatusScanning
)

const (
	MusicLibrary    = core.MusicLibrary
	VideoLibrary    = core.VideoLibrary
	DocumentLibrary = core.DocumentLibrary
)

// 重新导出错误
var (
	ErrLibraryNotFound  = core.ErrLibraryNotFound
	ErrLibraryDuplicate = core.ErrLibraryDuplicate
	ErrLibraryInUse     = core.ErrLibraryInUse
)
