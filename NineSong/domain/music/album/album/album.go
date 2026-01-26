package album

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/album/album/core"
)

// 重新导出核心类型
type AlbumMetadata = core.AlbumMetadata
type AlbumSongCounts = core.AlbumSongCounts
type AlbumFilterParams = core.AlbumFilterParams
type AlbumSortParams = core.AlbumSortParams
type AlbumRepository = core.AlbumRepository
