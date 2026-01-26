package media_file

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file/media_file/core"
)

// 重新导出核心类型
type MediaFileMetadata = core.MediaFileMetadata
type MediaFileQualityGroup = core.MediaFileQualityGroup
type MediaFileCounts = core.MediaFileCounts
type MediaFileRepository = core.MediaFileRepository
