package scene_audio_db_models

import "go.mongodb.org/mongo-driver/bson/primitive"

type LyricsFileMetadata struct {
	ID       primitive.ObjectID `bson:"_id"`
	Path     string             `bson:"path"`
	Title    string             `bson:"title"`
	Artist   string             `bson:"artist"`
	FileType string             `bson:"fileType"`
}
