package core

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WordCloudMetadata 词云元数据
type WordCloudMetadata struct {
	ID    primitive.ObjectID `bson:"_id"`
	Name  string             `bson:"name"`
	Count int                `bson:"count"`
	Type  string             `bson:"type"` // "artist", "album", "media", "media_cue"
	Rank  int                `bson:"rank"`
}

// WordCloudRecommendation 词云推荐
type WordCloudRecommendation struct {
	ID    primitive.ObjectID `bson:"_id"`
	Type  string             `bson:"type"`
	Name  string             `bson:"name"`
	Score float64            `bson:"score"` // 相关性分数
}

// Validate 验证词云元数据
func (w *WordCloudMetadata) Validate() error {
	// 可以添加验证逻辑
	return nil
}

// Validate 验证词云推荐
func (w *WordCloudRecommendation) Validate() error {
	// 可以添加验证逻辑
	return nil
}
