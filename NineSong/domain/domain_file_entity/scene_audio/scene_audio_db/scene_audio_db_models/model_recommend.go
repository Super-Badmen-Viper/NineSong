package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RecommendationResult 推荐结果
// 结合用户行为数据（播放次数、评分、收藏）和内容数据（词云标签）生成的推荐结果
type RecommendationResult struct {
	ID              primitive.ObjectID        `bson:"_id"`
	ItemID          string                    `bson:"item_id"`          // 推荐项目的ID
	ItemType        string                    `bson:"item_type"`        // 推荐项目类型: media_file, media_file_cue, album, artist
	Name            string                    `bson:"name"`             // 项目名称
	Score           float64                   `bson:"score"`            // 综合推荐分数
	Reason          string                    `bson:"reason"`           // 推荐理由
	PlayCount       int                       `bson:"play_count"`       // 播放次数
	Rating          int                       `bson:"rating"`           // 评分
	Starred         bool                      `bson:"starred"`          // 是否收藏
	Algorithm       string                    `bson:"algorithm"`        // 使用的推荐算法
	Parameters      map[string]string         `bson:"parameters"`       // 使用的参数
	Basis           []string                  `bson:"basis"`            // 推荐依据（基于哪些数据）
	AnnotationBasis []AnnotationMetadata      `bson:"annotation_basis"` // 基于的注释信息
	TagBasis        []WordCloudMetadata       `bson:"tag_basis"`        // 基于的标签信息
	RelatedItems    []WordCloudRecommendation `bson:"related_items"`    // 相关项目信息
}
