package media_file_cue

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file_cue/media_file_cue/core"
)

// 重新导出核心类型
type MediaFileCueMetadata = core.MediaFileCueMetadata
type MediaFileCueRepository = core.MediaFileCueRepository
type CueConfig = core.CueConfig
type CueREM = core.CueREM
type CueFile = core.CueFile
type CueIndex = core.CueIndex
type CueTrack = core.CueTrack
type MediaFileCueFilterCounts = core.MediaFileCueFilterCounts
