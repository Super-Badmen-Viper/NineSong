package sync_record

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/sync_record/core"
)

// 重新导出核心类型
type MediaLibrarySyncRecord = core.MediaLibrarySyncRecord
type MediaLibrarySyncRecordRepository = core.MediaLibrarySyncRecordRepository
type MediaLibrarySyncRecordResponse = core.MediaLibrarySyncRecordResponse

// 重新导出核心函数
var ConvertDBToRouteModel = core.ConvertDBToRouteModel
