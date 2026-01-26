package lyrics_file

// 导出core中的所有接口和类型
import (
	lyricsFileDomainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/lyrics_file/lyrics_file/core"
	lyricsFileRepoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/lyrics_file/lyrics_file/core"
)

// 重新导出核心类型
type LyricsFileRepository = lyricsFileDomainCore.LyricsFileRepository

// 重新导出构造函数
var NewLyricsFileRepository = lyricsFileRepoCore.NewLyricsFileRepository
