package media_library_audio

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/media_library/media_library_audio/core"
)

// 重新导出核心类型
type MediaLibraryAudioRepository = core.MediaLibraryAudioRepository

// 重新导出构造函数
var NewMediaLibraryAudioRepository = core.NewMediaLibraryAudioRepository
