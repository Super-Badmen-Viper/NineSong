package scene_audio_route_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// MediaFileCueMetadata 核心元数据结构
type MediaFileCueMetadata struct {
	// 系统保留字段 (综合)
	ID        primitive.ObjectID `bson:"_id"`        // 文档唯一标识符
	CreatedAt time.Time          `bson:"created_at"` // 文档创建时间
	UpdatedAt time.Time          `bson:"updated_at"` // 文档最后更新时间
	FullText  string             `bson:"full_text"`  // 音频文件全文文本内容，用于搜索
	Path      string             `bson:"path"`       // 音频文件的存储路径
	Suffix    string             `bson:"suffix"`     // 文件格式后缀（如 mp3、flac 等）
	Size      int                `bson:"size"`       // 文件大小（字节）

	// CUE记录信息
	Rem         CueREM     `bson:"rem"`
	Performer   string     `bson:"performer"`
	PerformerID string     `bson:"performer_id"`
	Title       string     `bson:"title"`
	File        CueFile    `bson:"file"`
	Catalog     string     `bson:"catalog"`    // 新增：唱片唯一EAN编号[8](@ref)
	SongWriter  string     `bson:"songwriter"` // 新增：乐曲编曲者[8](@ref)
	CueTracks   []CueTrack `bson:"cue_tracks"` // CUE 文件中的曲目信息列表

	CueTrackCount int `bson:"cue_track_count"` // CUE 文件中的曲目数量
}

type MediaFileCueFilterCounts struct {
	Total      int `bson:"total"`
	Starred    int `bson:"starred"`
	RecentPlay int `bson:"recent_play"`
}

type MediaFileCueListResponse struct {
	MediaFiles []MediaFileCueMetadata `bson:"media_files_cue"`
	Count      int                    `bson:"count"`
}

type CueREM struct {
	GENRE   string `bson:"genre"`
	DATE    string `bson:"date"`
	DISCID  string `bson:"discid"`
	COMMENT string `bson:"comment"`
}
type CueFile struct {
	FilePath string `bson:"file_path"`
	FileType string `bson:"file_type"`
}
type CueIndex struct {
	INDEX int    `bson:"index"`
	TIME  string `bson:"time"`
}
type CueTrack struct {
	TRACK       int        `bson:"track"`
	TYPE        string     `bson:"track_type"`
	Title       string     `bson:"track_title"`
	Performer   string     `bson:"track_performer"`
	PerformerID string     `bson:"track_performer_id"`
	FLAGS       string     `bson:"track_flags"`
	INDEXES     []CueIndex `bson:"track_indexes"`
	ISRC        string     `bson:"track_isrc"`
	GAIN        float64    `bson:"track_gain"`
	PEAK        float64    `bson:"track_peak"`
}
