package file

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/file_entity/file/core"
)

// 重新导出核心类型
type FileRepository = core.FileRepository

// 重新导出构造函数
var NewFileRepo = core.NewFileRepo
