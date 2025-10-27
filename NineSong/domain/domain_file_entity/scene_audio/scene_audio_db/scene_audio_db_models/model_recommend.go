package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RecommendationResult 推荐结果
// 结合用户行为数据（播放次数、评分、收藏）和内容数据（词云标签）生成的推荐结果
type RecommendationResult struct {
	ID        primitive.ObjectID `bson:"_id"`
	ItemID    string             `bson:"item_id"`    // 推荐项目的ID
	ItemType  string             `bson:"item_type"`  // 推荐项目类型: media_file, media_file_cue, album, artist
	Name      string             `bson:"name"`       // 项目名称
	Score     float64            `bson:"score"`      // 综合推荐分数
	Reason    string             `bson:"reason"`     // 推荐理由
	PlayCount int                `bson:"play_count"` // 播放次数
	Rating    int                `bson:"rating"`     // 评分
	Starred   bool               `bson:"starred"`    // 是否收藏
}
