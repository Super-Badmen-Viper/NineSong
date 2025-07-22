package scene_audio_db_interface

import (
	"context"
)

type LyricsFileRepository interface {
	GetLyricsFilePath(ctx context.Context, artist, title, fileType string) (string, error)
	UpdateLyricsFilePath(ctx context.Context, artist, title, fileType, path string) (bool, error)
	CleanAll(ctx context.Context) (bool, error)
}
