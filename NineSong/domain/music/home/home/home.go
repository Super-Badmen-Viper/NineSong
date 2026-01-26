package home

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/home/home/core"
)

// 重新导出核心类型
type HomeRepository = core.HomeRepository
