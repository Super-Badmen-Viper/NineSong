package annotation

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/annotation/annotation/core"
)

// 重新导出核心类型
type AnnotationMetadata = core.AnnotationMetadata
type TagSource = core.TagSource
type WeightedTag = core.WeightedTag
type AnnotationRepository = core.AnnotationRepository
