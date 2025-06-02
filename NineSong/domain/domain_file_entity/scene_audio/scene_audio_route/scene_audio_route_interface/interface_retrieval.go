package scene_audio_route_interface

import (
	"context"
)

type RetrievalRepository interface {
	GetStreamPath(ctx context.Context, mediaFileId string) (string, error)

	GetDownloadPath(ctx context.Context, mediaFileId string) (string, error)

	GetCoverArt(ctx context.Context, fileType string, targetID string) (string, error)

	GetLyricsLrcMetaData(ctx context.Context, mediaFileId string) (string, error)

	GetLyricsLrcFile(ctx context.Context, mediaFileId string) (string, error)
}
