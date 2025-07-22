package scene_audio_route_interface

import (
	"context"
)

type RetrievalRepository interface {
	GetStreamPath(ctx context.Context, mediaFileId string, cueModel bool) (string, error)

	GetStreamTempPath(ctx context.Context, metadataType string) (string, error)

	GetDownloadPath(ctx context.Context, mediaFileId string) (string, error)

	GetCoverArtID(ctx context.Context, fileType string, targetID string) (string, error)

	GetLyricsLrcMetaData(ctx context.Context, mediaFileId, artist, title, fileType string) (string, error)

	GetLyricsLrcFile(ctx context.Context, mediaFileId string) (string, error)
}
