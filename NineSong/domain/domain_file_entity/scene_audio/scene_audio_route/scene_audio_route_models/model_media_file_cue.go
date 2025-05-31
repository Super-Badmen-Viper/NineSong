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
	REM        CueREM     `json:"rem"`
	PERFORMER  string     `json:"performer"`
	TITLE      string     `json:"title"`
	FILE       CueFile    `json:"file"`
	CATALOG    string     `json:"catalog"`    // 新增：唱片唯一EAN编号[8](@ref)
	SONGWRITER string     `json:"songwriter"` // 新增：乐曲编曲者[8](@ref)
	CueTracks  []CueTrack `bson:"cue_tracks"` // CUE 文件中的曲目信息列表

	CueTrackCount int `bson:"cue_track_count"` // CUE 文件中的曲目数量
}

type MediaFileCueFilterCounts struct {
	Total      int `json:"total"`
	Starred    int `json:"starred"`
	RecentPlay int `json:"recent_play"`
}

type MediaFileCueListResponse struct {
	MediaFiles []MediaFileCueMetadata `json:"media_files_cue"`
	Count      int                    `json:"count"`
}

type CueREM struct {
	GENRE   string `json:"genre"`
	DATE    string `json:"date"`
	DISCID  string `json:"discid"`
	COMMENT string `json:"comment"`
}
type CueFile struct {
	FilePath string `json:"file_path"`
	FileType string `json:"file_type"`
}
type CueIndex struct {
	INDEX int    `json:"index"`
	TIME  string `json:"time"`
}
type CueTrack struct {
	TRACK     int        `json:"track"`
	TYPE      string     `json:"track_type"`
	TITLE     string     `json:"track_title"`
	PERFORMER string     `json:"track_performer"`
	FLAGS     string     `json:"track_flags"`
	INDEXES   []CueIndex `json:"track_indexes"`
	ISRC      string     `json:"track_isrc"`
	GAIN      float64    `json:"track_gain"`
	PEAK      float64    `json:"track_peak"`
}
