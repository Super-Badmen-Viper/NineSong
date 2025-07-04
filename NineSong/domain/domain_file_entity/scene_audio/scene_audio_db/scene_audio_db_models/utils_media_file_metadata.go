package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (m *MediaFileMetadata) ToUpdateDoc() bson.M {
	data, _ := bson.Marshal(m)
	var raw bson.M
	_ = bson.Unmarshal(data, &raw)

	delete(raw, "_id")
	delete(raw, "created_at")

	raw["thumbnail_url"] = m.ThumbnailURL
	raw["medium_image_url"] = m.MediumImageURL
	raw["high_image_url"] = m.HighImageURL
	raw["updated_at"] = time.Now().UTC()

	return bson.M{"$set": raw}
}

func (m *MediaFileCueMetadata) ToUpdateDoc() bson.M {
	data, _ := bson.Marshal(m)
	var raw bson.M
	_ = bson.Unmarshal(data, &raw)

	delete(raw, "_id")
	delete(raw, "created_at")
	raw["back_image_url"] = m.CueResources.BackImage
	raw["cover_image_url"] = m.CueResources.CoverImage
	raw["disc_image_url"] = m.CueResources.DiscImage
	raw["has_cover_art"] = m.CueResources.CoverImage != ""
	raw["updated_at"] = time.Now().UTC()

	return bson.M{"$set": raw}
}
