package scene_audio_route_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaFileMetadata struct {
	ID             primitive.ObjectID `bson:"_id"`
	Path           string             `bson:"path"`
	Title          string             `bson:"title"`
	Album          string             `bson:"album"`
	Artist         string             `bson:"artist"`
	ArtistID       string             `bson:"artist_id"`
	AlbumArtist    string             `bson:"album_artist"`
	AlbumID        string             `bson:"album_id"`
	HasCoverArt    bool               `bson:"has_cover_art"`
	Year           int                `bson:"year"`
	Size           int                `bson:"size"`
	Suffix         string             `bson:"suffix"`       // 文件后缀
	FileName       string             `bson:"file_name"`    // 文件名（不包含路径）
	LibraryPath    string             `bson:"library_path"` // 音频文件所在的音乐库路径
	Duration       float64            `bson:"duration"`
	BitRate        int                `bson:"bit_rate"`
	EncodingFormat string             `bson:"encoding_format"` // 编码格式（如 PCM、MP3、AAC 等）
	Genre          string             `bson:"genre"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	AlbumArtistID  string             `bson:"album_artist_id"`
	Channels       int                `bson:"channels"`

	Compilation       bool           `bson:"compilation"`          // 是否为合辑（多艺术家作品合集）
	AllArtistIDs      []ArtistIDPair `bson:"all_artist_ids"`       // 所有参与艺术家的唯一标识符列表
	AllAlbumArtistIDs []ArtistIDPair `bson:"all_album_artist_ids"` // 所有参与专辑艺术家的唯一标识符列表

	PlayCount         int       `bson:"play_count"`
	PlayCompleteCount int       `bson:"play_complete_count"`
	PlayDate          time.Time `bson:"play_date"`
	Rating            int       `bson:"rating"`
	Starred           bool      `bson:"starred"`
	StarredAt         time.Time `bson:"starred_at"`

	Index int `bson:"index" json:"Index"`
}

type MediaFileFilterCounts struct {
	Total      int `json:"total"`
	Starred    int `json:"starred"`
	RecentPlay int `json:"recent_play"`
}
