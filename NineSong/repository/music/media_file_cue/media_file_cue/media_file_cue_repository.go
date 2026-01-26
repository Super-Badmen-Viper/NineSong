package media_file_cue

// 导出core中的所有接口和类型
import (
	mediaFileCueDomainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file_cue/media_file_cue/core"
	mediaFileCueRepoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/media_file_cue/media_file_cue/core"
)

// 重新导出核心类型
type MediaFileCueRepository = mediaFileCueDomainCore.MediaFileCueRepository

// 重新导出构造函数
var NewMediaFileCueRepository = mediaFileCueRepoCore.NewMediaFileCueRepository
