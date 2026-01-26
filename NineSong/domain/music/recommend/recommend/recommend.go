package recommend

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/recommend/recommend/core"
)

// 重新导出核心类型
type RecommendationResult = core.RecommendationResult
type RecommendRepository = core.RecommendRepository
