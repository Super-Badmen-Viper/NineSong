package scene_audio_db_models

import "go.mongodb.org/mongo-driver/bson/primitive"

type MediaSoundEffectsMetadata struct {
	ID primitive.ObjectID `bson:"_id"`
}
