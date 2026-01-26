package media_file

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/media_file/media_file/core"
)

// 重新导出核心类型
type MediaFileRepository = core.MediaFileRepository

// 重新导出构造函数
var NewMediaFileRepository = core.NewMediaFileRepository
