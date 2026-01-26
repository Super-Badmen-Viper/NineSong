package word_cloud

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/word_cloud/word_cloud/core"
)

// 重新导出核心类型
type WordCloudMetadata = core.WordCloudMetadata
type WordCloudRecommendation = core.WordCloudRecommendation
type WordCloudRepository = core.WordCloudRepository
