package folder

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/music/file_entity/folder/core"
)

// 重新导出核心类型
type LibraryUsecase = core.LibraryUsecase
type LibraryUsecaseWithScan = core.LibraryUsecaseWithScan

// 重新导出构造函数
var NewLibraryUsecase = core.NewLibraryUsecase
