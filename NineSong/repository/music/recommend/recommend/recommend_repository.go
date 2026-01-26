package recommend

// 导出core中的所有接口和类型
import (
	recommendDomainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/recommend/recommend/core"
	recommendRepoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/recommend/recommend/core"
)

// 重新导出核心类型
type RecommendRepository = recommendDomainCore.RecommendRepository

// 重新导出构造函数
var NewRecommendRepository = recommendRepoCore.NewRecommendRepository
