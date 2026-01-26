package core

import (
	"context"

	albumCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/album/album/core"
	artistCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/artist/artist/core"
	mediaFileCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file/media_file/core"
	mediaFileCueCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_file_cue/media_file_cue/core"
)

// HomeRepository 首页仓库接口
type HomeRepository interface {
	GetRandomArtistList(
		ctx context.Context,
		start string,
		end string,
	) ([]*artistCore.ArtistMetadata, error)
	GetRandomAlbumList(
		ctx context.Context,
		start string,
		end string,
	) ([]*albumCore.AlbumMetadata, error)
	GetRandomMediaFileList(
		ctx context.Context,
		start string,
		end string,
	) ([]*mediaFileCore.MediaFileMetadata, error)
	GetRandomMediaCueList(
		ctx context.Context,
		start string,
		end string,
	) ([]*mediaFileCueCore.MediaFileCueMetadata, error)
}
