package scene_audio_route_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaFileMetadata struct {
	ID            primitive.ObjectID `bson:"_id"`
	Path          string             `bson:"path"`
	Title         string             `bson:"title"`
	Album         string             `bson:"album"`
	Artist        string             `bson:"artist"`
	ArtistID      string             `bson:"artist_id"`
	AlbumArtist   string             `bson:"album_artist"`
	AlbumID       string             `bson:"album_id"`
	HasCoverArt   bool               `bson:"has_cover_art"`
	Year          int                `bson:"year"`
	Size          int                `bson:"size"`
	Suffix        string             `bson:"suffix"` // 文件后缀
	Duration      float64            `bson:"duration"`
	BitRate       int                `bson:"bit_rate"`
	Genre         string             `bson:"genre"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
	AlbumArtistID string             `bson:"album_artist_id"`
	Channels      int                `bson:"channels"`

	Compilation       bool           `bson:"compilation"`          // 是否为合辑（多艺术家作品合集）
	AllArtistIDs      []ArtistIDPair `bson:"all_artist_ids"`       // 所有参与艺术家的唯一标识符列表
	AllAlbumArtistIDs []ArtistIDPair `bson:"all_album_artist_ids"` // 所有参与专辑艺术家的唯一标识符列表

	CueComplete   bool          `bson:"cue_complete"`    // 是否为完整的 CUE 文件（包含所有曲目）
	CueResources  CueConfigs    `bson:"cue_resources"`   // CUE 文件相关资源信息
	CueGlobalMeta CueGlobalMeta `bson:"cue_global_meta"` // CUE 文件全局元数据
	CueTracks     []CueTrack    `bson:"cue_tracks"`      // CUE 文件中的曲目信息列表

	PlayCount int       `bson:"play_count"`
	PlayDate  time.Time `bson:"play_date"`
	Rating    int       `bson:"rating"`
	Starred   bool      `bson:"starred"`
	StarredAt time.Time `bson:"starred_at"`

	Index int `bson:"index" json:"Index"`
}

type MediaFileFilterCounts struct {
	Total      int `json:"total"`
	Starred    int `json:"starred"`
	RecentPlay int `json:"recent_play"`
}

type MediaFileListResponse struct {
	MediaFiles []MediaFileMetadata `json:"media_files"`
	Count      int                 `json:"count"`
}

type CueConfigs struct {
	CuePath    string // .cue文件路径
	AudioPath  string // 关联的.wav文件路径
	BackImage  string // back.jpg路径
	CoverImage string // cover.jpg路径
	DiscImage  string // disc.jpg路径
	ListFile   string // list.txt路径
	LogFile    string // log.txt路径
}
type CueREM struct {
	GENRE   string
	DATE    string
	DISCID  string
	COMMENT string
}
type CueFile struct {
	FilePath string
	FileType string
}
type CueIndex struct {
	INDEX int
	TIME  string
}
type CueGlobalMeta struct {
	REM        CueREM  `json:"cue_rem"`
	PERFORMER  string  `json:"cue_performer"`
	TITLE      string  `json:"cue_title"`
	FILE       CueFile `json:"cue_file"`
	CATALOG    string  `json:"cue_catalog"`    // 新增：唱片唯一EAN编号[8](@ref)
	SONGWRITER string  `json:"cue_songwriter"` // 新增：乐曲编曲者[8](@ref)
}
type CueTrack struct {
	TRACK     int        `json:"cue_track_track"`
	TYPE      string     `json:"cue_track_type"`
	TITLE     string     `json:"cue_track_title"`
	PERFORMER string     `json:"cue_track_performer"`
	FLAGS     string     `json:"cue_track_flags"`
	INDEXES   []CueIndex `json:"cue_track_indexes"`
	ISRC      string     `json:"cue_track_isrc"` // 新增：国际标准录音代码[9](@ref)
	GAIN      float64    `json:"cue_track_gain"` // 新增：ReplayGain增益值
	PEAK      float64    `json:"cue_track_peak"` // 新增：ReplayGain峰值
}
