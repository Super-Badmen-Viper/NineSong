package word_cloud

// 导出core中的所有接口和类型
import (
	wordCloudDomainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/word_cloud/word_cloud/core"
	wordCloudRepoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/word_cloud/word_cloud/core"
)

// 重新导出核心类型
type WordCloudRepository = wordCloudDomainCore.WordCloudRepository

// 重新导出构造函数
var NewWordCloudRepository = wordCloudRepoCore.NewWordCloudRepository
