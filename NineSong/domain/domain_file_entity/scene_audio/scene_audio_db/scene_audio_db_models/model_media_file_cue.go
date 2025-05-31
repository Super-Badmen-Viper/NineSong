package scene_audio_db_models

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

	CueTrackCount int       `bson:"cue_track_count"` // CUE 文件中的曲目数量
	CueResources  CueConfig `bson:"cue_resources"`   // CUE 文件相关资源信息

	// 基础元数据: 视觉元素
	HasCoverArt   bool   `bson:"has_cover_art"` // 是否包含专辑封面图
	BackImageURL  string `bson:"back_image_url"`
	CoverImageURL string `bson:"cover_image_url"`
	DiscImageURL  string `bson:"disc_image_url"`

	// 音频分析 (综合)
	CueSampleRate int     `bson:"cue_sample_rate"` // 音频采样率（Hz）
	CueDuration   float64 `bson:"cue_duration"`    // 音频时长（秒）
	CueBitRate    int     `bson:"cue_bit_rate"`    // 比特率（bps）
	CueChannels   int     `bson:"cue_channels"`    // 音频通道数（如 2 表示立体声）

	// 高级音频参数 (github.com/go-audio/audio)
	BitDepth      int    `bson:"bit_depth"`      // 音频位深（位）
	ChannelLayout string `bson:"channel_layout"` // 声道布局（如立体声、环绕声等）
}

type CueConfig struct {
	CuePath    string `json:"cue_path"`
	AudioPath  string `json:"audio_path"`
	BackImage  string `json:"back_image"`  // 背景图片路径
	CoverImage string `json:"cover_image"` // 封面图片路径
	DiscImage  string `json:"disc_image"`  // 光盘图片路径
	ListFile   string `json:"list_file"`   // 列表文件路径
	LogFile    string `json:"log_file"`    // 日志文件路径
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
	TRACK     int               `json:"track"`
	TYPE      string            `json:"track_type"`
	TITLE     string            `json:"track_title"`
	PERFORMER string            `json:"track_performer"`
	FLAGS     string            `json:"track_flags"`
	INDEXES   []CueIndex        `json:"track_indexes"`
	ISRC      string            `json:"track_isrc"`
	GAIN      float64           `json:"track_gain"`
	PEAK      float64           `json:"track_peak"`
	Extended  MediaFileMetadata `bson:"cue_track_extended"` // 嵌入 MediaFileMetadata 以复用通用字段
}
