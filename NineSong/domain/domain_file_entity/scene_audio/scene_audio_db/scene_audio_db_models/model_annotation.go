package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AnnotationMetadata struct {
	ID        primitive.ObjectID `bson:"_id"`        // 文档唯一标识符
	UserID    string             `bson:"user_id"`    // 用户唯一标识符，标识创建此注释的用户
	ItemID    string             `bson:"item_id"`    // 媒体项目唯一标识符，标识被注释的媒体项目
	ItemType  string             `bson:"item_type"`  // 媒体项目类型（如音乐、视频、图片等）
	PlayCount int                `bson:"play_count"` // 播放次数，记录该媒体项目被播放的次数
	PlayDate  time.Time          `bson:"play_date"`  // 播放日期，最近一次播放此媒体项目的日期和时间
	Rating    int                `bson:"rating"`     // 评分，用户对此媒体项目的评分（如1-5分）
	Starred   bool               `bson:"starred"`    // 是否收藏，标识该媒体项目是否被用户收藏
	StarredAt time.Time          `bson:"starred_at"` // 收藏时间，媒体项目被收藏的日期和时间
}
