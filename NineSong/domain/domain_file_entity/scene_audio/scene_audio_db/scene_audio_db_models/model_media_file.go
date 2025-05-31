package scene_audio_db_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaFileMetadata 核心元数据结构
type MediaFileMetadata struct {
	// 系统保留字段 (综合)
	ID        primitive.ObjectID `bson:"_id"`        // 文档唯一标识符
	CreatedAt time.Time          `bson:"created_at"` // 文档创建时间
	UpdatedAt time.Time          `bson:"updated_at"` // 文档最后更新时间
	FullText  string             `bson:"full_text"`  // 音频文件全文文本内容，用于搜索
	Path      string             `bson:"path"`       // 音频文件的存储路径
	Suffix    string             `bson:"suffix"`     // 文件格式后缀（如 mp3、flac 等）
	Size      int                `bson:"size"`       // 文件大小（字节）

	// 基础元数据 (github.com/dhowden/tag、go.senan.xyz/taglib)
	Title             string   `bson:"title"`               // 标准曲目标题
	Album             string   `bson:"album"`               // 所属专辑名称
	Artist            string   `bson:"artist"`              // 表演者名称
	AlbumArtist       string   `bson:"album_artist"`        // 专辑级艺术家名称（可能不同于曲目艺术家）
	TitlePinyin       []string `bson:"title_pinyin"`        // 曲目标题的拼音表示（用于搜索和排序）
	AlbumPinyin       []string `bson:"album_pinyin"`        // 专辑名称的拼音表示（用于搜索和排序）
	ArtistPinyin      []string `bson:"artist_pinyin"`       // 表演者名称的拼音表示（用于搜索和排序）
	AlbumArtistPinyin []string `bson:"album_artist_pinyin"` // 专辑艺术家名称的拼音表示（用于搜索和排序）
	Genre             string   `bson:"genre"`               // 音乐流派（如流行、摇滚等）
	Year              int      `bson:"year"`                // 发行年份
	TrackNumber       int      `bson:"track_number"`        // 轨道序号（曲目在专辑中的编号）
	DiscNumber        int      `bson:"disc_number"`         // 光盘编号（多光盘专辑中的编号）
	TotalTracks       int      `bson:"total_tracks"`        // 专辑总轨道数
	TotalDiscs        int      `bson:"total_discs"`         // 总光盘数
	Composer          string   `bson:"composer"`            // 作曲家名称
	Comment           string   `bson:"comment"`             // 注释信息
	Lyrics            string   `bson:"lyrics"`              // 歌词文本内容
	Compilation       bool     `bson:"compilation"`         // 是否为合辑（多艺术家作品合集）

	// 基础元数据: 关系ID索引
	ArtistID          string         `bson:"artist_id"`            // 艺术家在系统中的唯一标识符
	AlbumID           string         `bson:"album_id"`             // 专辑在系统中的唯一标识符
	AlbumArtistID     string         `bson:"album_artist_id"`      // 专辑艺术家在系统中的唯一标识符
	AllArtistIDs      []ArtistIDPair `bson:"all_artist_ids"`       // 所有参与艺术家的唯一标识符列表
	AllAlbumArtistIDs []ArtistIDPair `bson:"all_album_artist_ids"` // 所有参与专辑艺术家的唯一标识符列表
	MvID              string         `bson:"mv_id"`                // 音频对应的音乐视频唯一标识符（如有）
	KaraokeID         string         `bson:"karaoke_id"`           // 音频对应的卡拉 OK 版本唯一标识符（如有）
	LyricsID          string         `bson:"lyrics_id"`            // 歌词文件的唯一标识符（如有）
	MusicID           string         `bson:"music_id"`             // 音频对应的乐谱唯一标识符（如有）
	GeneratorID       string         `bson:"generator_id"`         // 该音频生成的参数类型（如有）
	TransformsID      string         `bson:"transforms_id"`        // 该音频转换的数据类型（如有）

	// 基础元数据: 索引排序信息
	Index                int    `bson:"index" json:"Index"`      // 索引值，可用于排序或其他用途
	SortTitle            string `bson:"sort_title"`              // 排序用标题（去除冠词，如 "The"）
	SortAlbumName        string `bson:"sort_album_name"`         // 标准化专辑名称（去除非字母字符，便于排序）
	SortArtistName       string `bson:"sort_artist_name"`        // 标准化艺术家名称（便于排序）
	SortAlbumArtistName  string `bson:"sort_album_artist_name"`  // 标准化专辑艺术家名称（便于排序）
	OrderTitle           string `bson:"order_title"`             // 排序用标题（去除前缀词，如 "A"）
	OrderAlbumName       string `bson:"order_album_name"`        // 排序用专辑名称（忽略冠词，便于排序）
	OrderArtistName      string `bson:"order_artist_name"`       // 排序用艺术家名称（便于排序）
	OrderAlbumArtistName string `bson:"order_album_artist_name"` // 排序用专辑艺术家名称（便于排序）

	// 基础元数据: 视觉元素
	HasCoverArt    bool   `bson:"has_cover_art"`    // 是否包含专辑封面图
	ThumbnailURL   string `bson:"thumbnail_url"`    // 缩略图的 URL 地址（尺寸 200x200px）
	MediumImageURL string `bson:"medium_image_url"` // 专辑封面中等分辨率图的 URL 地址
	HighImageURL   string `bson:"high_image_url"`   // 高清封面图的 URL 地址（尺寸不低于 1200x1200px）

	// 基础元数据: 音质
	QualityVersion          string    `bson:"quality_version"`            // 音质版本描述（如 320kbps、FLAC 等）
	QualityVersionID        string    `bson:"quality_version_id"`         // 音质版本的唯一标识符
	QualityVersionCreatedAt time.Time `bson:"quality_version_created_at"` // 音质版本的创建时间（UTC 时间）
	QualityVersionUpdateAt  time.Time `bson:"quality_version_update_at"`  // 音质版本的最后更新时间（UTC 时间）

	// 扩展存储 (综合)
	CustomTags map[string]string `bson:"custom_tags"` // 自定义标签键值对
	MoodTags   []string          `bson:"mood_tags"`   // 心情标签（如欢快、悲伤等）
	SceneTags  []string          `bson:"scene_tags"`  // 场景标签（如运动、学习等）
	Language   string            `bson:"language"`    // 主要语言（使用 ISO 639-2 标准代码）

	// 地理信息 (综合)
	LocationName string  `bson:"location_name"` // 录制地理位置的名称（如城市、场所等）
	GPSLatitude  float64 `bson:"gps_lat"`       // 纬度坐标（基于 WGS84 坐标系）
	GPSLongitude float64 `bson:"gps_lng"`       // 经度坐标（基于 WGS84 坐标系）

	// MusicBrainz元数据 (github.com/michiwend/gomusicbrainz)
	MBZTrackID        string `bson:"mbz_track_id"`         // MusicBrainz 曲目唯一标识符
	MBZAlbumID        string `bson:"mbz_album_id"`         // MusicBrainz 专辑唯一标识符
	MBZArtistID       string `bson:"mbz_artist_id"`        // MusicBrainz 艺术家唯一标识符
	MBZAlbumArtistID  string `bson:"mbz_album_artist_id"`  // MusicBrainz 专辑艺术家唯一标识符
	MBZReleaseTrackID string `bson:"mbz_release_track_id"` // MusicBrainz 发行版曲目唯一标识符
	MBZAlbumType      string `bson:"mbz_album_type"`       // 专辑类型（如专辑、单曲等）
	MBZAlbumComment   string `bson:"mbz_album_comment"`    // MusicBrainz 专辑评论信息
	DiscSubtitle      string `bson:"disc_subtitle"`        // 光盘副标题（如特别版、纪念版等说明）
	CatalogNum        string `bson:"catalog_num"`          // 唱片目录编号（发行方的内部编号）

	// 音频分析 (综合)
	SampleRate int     `bson:"sample_rate"` // 音频采样率（Hz）
	Duration   float64 `bson:"duration"`    // 音频时长（秒）
	BitRate    int     `bson:"bit_rate"`    // 比特率（bps）
	Channels   int     `bson:"channels"`    // 音频通道数（如 2 表示立体声）

	// 高级音频参数 (github.com/go-audio/audio)
	BitDepth      int    `bson:"bit_depth"`      // 音频位深（位）
	ChannelLayout string `bson:"channel_layout"` // 声道布局（如立体声、环绕声等）

	// 音频标准化与动态响度控制 (综合)
	NormalizationThreshold float64 `bson:"norm_threshold"` // 音频标准化阈值
	RGAlbumGain            float64 `bson:"rg_album_gain"`  // ReplayGain 专辑增益值
	RGAlbumPeak            float64 `bson:"rg_album_peak"`  // ReplayGain 专辑峰值
	RGTrackGain            float64 `bson:"rg_track_gain"`  // ReplayGain 曲目增益值
	RGTrackPeak            float64 `bson:"rg_track_peak"`  // ReplayGain 曲目峰值

	// 版权与发行 (综合)
	ISRC               string    `bson:"isrc"`                // 国际标准录音代码（遵循 ISO 3901 标准）
	MCDI               string    `bson:"mcdi"`                // 音乐 CD 标识符（基于 CD-TEXT 标准）
	AcoustID           string    `bson:"acoust_id"`           // 音频指纹标识（用于 AcoustID 数据库与 MusicBrainz 关联）
	AcoustFingerprint  string    `bson:"acoust_fingerprint"`  // 音频指纹数据（用于音频识别）
	EducationalContent bool      `bson:"educational_content"` // 是否为教育内容（如教学音频、有声教材等）
	Copyright          string    `bson:"copyright"`           // 版权信息（如版权持有者、版权声明等）
	Publisher          string    `bson:"publisher"`           // 出版者名称
	ReleaseDate        time.Time `bson:"release_date"`        // 发行日期
	RecordingDate      time.Time `bson:"recording_date"`      // 录音日期
	Producer           string    `bson:"producer"`            // 制作人名称
	Engineer           string    `bson:"engineer"`            // 工程师名称（如录音工程师、混音工程师等）
	Studio             string    `bson:"studio"`              // 录音室名称
	RecordingLocation  string    `bson:"recording_location"`  // 录音地点详细描述（如城市中的录音棚名称等）
}
