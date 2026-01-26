package artist

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/artist/artist/core"
)

// 重新导出核心类型
type ArtistRepository = core.ArtistRepository

// 重新导出构造函数
var NewArtistRepository = core.NewArtistRepository
