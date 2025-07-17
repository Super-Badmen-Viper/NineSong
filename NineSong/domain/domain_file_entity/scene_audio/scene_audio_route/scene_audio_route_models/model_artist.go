package scene_audio_route_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ArtistMetadata struct {
	ID              primitive.ObjectID `bson:"_id"`
	Name            string             `bson:"name"`
	AlbumCount      int                `bson:"album_count"`
	GuestAlbumCount int                `bson:"guest_album_count"`
	SongCount       int                `bson:"song_count"`
	GuestSongCount  int                `bson:"guest_song_count"`
	CueCount        int                `bson:"cue_count"`
	GuestCueCount   int                `bson:"guest_cue_count"`
	Size            int                `bson:"size"`
	HasCoverArt     bool               `bson:"has_cover_art"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`

	Compilation       bool           `bson:"compilation"`          // 是否为合辑（多艺术家作品合集）
	AllArtistIDs      []ArtistIDPair `bson:"all_artist_ids"`       // 所有参与艺术家的唯一标识符列表
	AllAlbumArtistIDs []ArtistIDPair `bson:"all_album_artist_ids"` // 所有参与专辑艺术家的唯一标识符列表

	ImageFiles string `bson:"image_files"` // 为空则不存在cover封面，从媒体文件中提取

	PlayCount         int       `bson:"play_count"`
	PlayCompleteCount int       `bson:"play_complete_count"`
	PlayDate          time.Time `bson:"play_date"`
	Rating            int       `bson:"rating"`
	Starred           bool      `bson:"starred"`
	StarredAt         time.Time `bson:"starred_at"`
}

type ArtistFilterCounts struct {
	Total      int `json:"total"`
	Starred    int `json:"starred"`
	RecentPlay int `json:"recent_play"`
}
