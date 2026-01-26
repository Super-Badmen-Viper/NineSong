package playlist_track

// 导出core中的所有接口和类型
import (
	playlistTrackDomainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/playlist/playlist_track/core"
	playlistTrackRepoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/playlist/playlist_track/core"
)

// 重新导出核心类型
type PlaylistTrackRepository = playlistTrackDomainCore.PlaylistTrackRepository

// 重新导出构造函数
var NewPlaylistTrackRepository = playlistTrackRepoCore.NewPlaylistTrackRepository
