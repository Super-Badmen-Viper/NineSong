package core

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LyricsFileMetadata 歌词文件元数据
type LyricsFileMetadata struct {
	ID       primitive.ObjectID `bson:"_id"`
	Path     string             `bson:"path"`
	Title    string             `bson:"title"`
	Artist   string             `bson:"artist"`
	FileType string             `bson:"fileType"`
}

// Validate 验证歌词文件元数据
func (l *LyricsFileMetadata) Validate() error {
	// 可以添加验证逻辑
	return nil
}
