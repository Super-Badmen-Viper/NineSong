package annotation

// 导出core中的所有接口和类型
import (
	annotationDomainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/annotation/annotation/core"
	annotationRepoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/annotation/annotation/core"
)

// 重新导出核心类型
type AnnotationRepository = annotationDomainCore.AnnotationRepository

// 重新导出构造函数
var NewAnnotationRepository = annotationRepoCore.NewAnnotationRepository
