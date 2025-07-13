package scene_audio_route_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlbumMetadata struct {
	ID          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	ArtistID    string             `bson:"artist_id"`
	Artist      string             `bson:"artist"`
	AlbumArtist string             `bson:"album_artist"`
	HasCoverArt bool               `bson:"has_cover_art"`

	MinYear       int       `bson:"min_year"`
	MaxYear       int       `bson:"max_year"`
	SongCount     int       `bson:"song_count"`
	Duration      float64   `bson:"duration"`
	Size          int       `bson:"size"`
	Genre         string    `bson:"genre"`
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
	AlbumArtistID string    `bson:"album_artist_id"`
	Comment       string    `bson:"comment"`
	ImageFiles    string    `bson:"image_files"` // 为空则不存在cover封面，从媒体文件中提取

	Compilation       bool           `bson:"compilation"`          // 是否为合辑（多艺术家作品合集）
	AllArtistIDs      []ArtistIDPair `bson:"all_artist_ids"`       // 所有参与艺术家的唯一标识符列表
	AllAlbumArtistIDs []ArtistIDPair `bson:"all_album_artist_ids"` // 所有参与专辑艺术家的唯一标识符列表

	PlayCount         int       `bson:"play_count"`
	PlayCompleteCount int       `bson:"play_complete_count"`
	PlayDate          time.Time `bson:"play_date"`
	Rating            int       `bson:"rating"`
	Starred           bool      `bson:"starred"`
	StarredAt         time.Time `bson:"starred_at"`
}

type AlbumFilterCounts struct {
	Total      int `json:"total"`
	Starred    int `json:"starred"`
	RecentPlay int `json:"recent_play"`
}

type ArtistIDPair struct {
	ArtistName string `bson:"artist_name"` // 艺术家名称
	ArtistID   string `bson:"artist_id"`   // 艺术家唯一 ID
}
