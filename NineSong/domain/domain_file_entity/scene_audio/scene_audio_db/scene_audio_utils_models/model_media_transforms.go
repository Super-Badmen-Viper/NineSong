package scene_audio_utils_models

import "go.mongodb.org/mongo-driver/bson/primitive"

// github.com/go-audio/transforms
type MediaTransformsMetadata struct {
	ID primitive.ObjectID `bson:"_id"`
}
