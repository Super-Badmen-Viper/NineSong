package core

import (
	"context"
)

// LyricsFileRepository 歌词文件仓库接口
type LyricsFileRepository interface {
	GetLyricsFilePath(ctx context.Context, artist, title, fileType string) (string, error)
	UpdateLyricsFilePath(ctx context.Context, artist, title, fileType, path string) (bool, error)
	CleanAll(ctx context.Context) (bool, error)
}
