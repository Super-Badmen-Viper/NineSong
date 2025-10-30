package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WordCloudMetadata struct {
	ID    primitive.ObjectID `bson:"_id"`
	Name  string             `bson:"name"`
	Count int                `bson:"count"`
	Type  string             `bson:"type"` // "artist", "album", "media", "media_cue"
	Rank  int                `bson:"rank"`
}

type WordCloudRecommendation struct {
	ID    primitive.ObjectID `bson:"_id"`
	Type  string             `bson:"type"`
	Name  string             `bson:"name"`
	Score float64            `bson:"score"` // 相关性分数
}
