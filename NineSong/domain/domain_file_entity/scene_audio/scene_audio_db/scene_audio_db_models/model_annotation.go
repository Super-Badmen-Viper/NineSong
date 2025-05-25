package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AnnotationMetadata struct {
	ID        primitive.ObjectID `bson:"_id"`
	UserID    string             `bson:"user_id"`
	ItemID    string             `bson:"item_id"`
	ItemType  string             `bson:"item_type"`
	PlayCount int                `bson:"play_count"`
	PlayDate  time.Time          `bson:"play_date"`
	Rating    int                `bson:"rating"`
	Starred   bool               `bson:"starred"`
	StarredAt time.Time          `bson:"starred_at"`
}
