package playlist

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/playlist/playlist/core"
)

// 重新导出核心类型
type PlaylistRepository = core.PlaylistRepository

// 重新导出构造函数
var NewPlaylistRepository = core.NewPlaylistRepository
