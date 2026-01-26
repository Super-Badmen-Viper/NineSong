package sync_record

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/media_library/sync_record/core"
)

// 重新导出核心类型
type MediaLibrarySyncRecordRepository = core.MediaLibrarySyncRecordRepository

// 重新导出构造函数
var NewMediaLibrarySyncRecordRepository = core.NewMediaLibrarySyncRecordRepository
