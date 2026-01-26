package file

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/music/file_entity/file/core"
)

// 重新导出核心类型
type FileUsecase = core.FileUsecase
type ScanManager = core.ScanManager

// 重新导出构造函数
var NewFileUsecase = core.NewFileUsecase
var NewScanManager = core.NewScanManager
