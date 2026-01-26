package core

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlaylistTrackMetadata 播放列表曲目元数据
type PlaylistTrackMetadata struct {
	ID          primitive.ObjectID `bson:"_id"`
	PlaylistID  primitive.ObjectID `bson:"playlist_id"`
	MediaFileID primitive.ObjectID `bson:"media_file_id"`
	Index       int                `bson:"index"`
}

// Validate 验证播放列表曲目元数据
func (p *PlaylistTrackMetadata) Validate() error {
	// 可以添加验证逻辑
	return nil
}
