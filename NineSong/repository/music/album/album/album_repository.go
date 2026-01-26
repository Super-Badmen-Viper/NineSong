package album

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/album/album/core"
)

// 重新导出核心类型
type AlbumRepository = core.AlbumRepository

// 重新导出构造函数
var NewAlbumRepository = core.NewAlbumRepository
