package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WordCloudMetadata struct {
	ID    primitive.ObjectID
	Name  string
	Count int
	Type  string // "media_file" or "media_file_cue"
	Rank  int
}

type Recommendation struct {
	ID    primitive.ObjectID
	Type  string
	Name  string
	Score float64 // 相关性分数
}
