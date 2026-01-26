package sync_record

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/music/media_library/sync_record/core"
)

// 重新导出核心类型
type MediaLibrarySyncRecordUseCase = core.MediaLibrarySyncRecordUseCase

// 重新导出构造函数
var NewMediaLibrarySyncRecordUseCase = core.NewMediaLibrarySyncRecordUseCase
