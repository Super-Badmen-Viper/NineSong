package artist

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/artist/artist/core"
)

// 重新导出核心类型
type ArtistIDPair = core.ArtistIDPair
type ArtistMetadata = core.ArtistMetadata
type ArtistAlbumAndSongCounts = core.ArtistAlbumAndSongCounts
type ArtistRepository = core.ArtistRepository
