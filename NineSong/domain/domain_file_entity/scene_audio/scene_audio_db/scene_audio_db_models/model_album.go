package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AlbumMetadata struct {
	// 系统保留字段 (综合)
	ID        primitive.ObjectID `bson:"_id"`        // 文档唯一标识符
	CreatedAt time.Time          `bson:"created_at"` // 文档创建时间
	UpdatedAt time.Time          `bson:"updated_at"` // 文档最后更新时间

	// 基础元数据 (综合)
	Name                  string   `bson:"name"`                     // 专辑名称
	Artist                string   `bson:"artist"`                   // 表演者名称
	AlbumArtist           string   `bson:"album_artist"`             // 专辑级艺术家名称（可能不同于曲目艺术家）
	NamePinyin            []string `bson:"name_pinyin"`              // 专辑名称的拼音（用于搜索和排序）
	NamePinyinFull        string   `bson:"name_pinyin_full"`         // 专辑名称的完整拼音（用于搜索和排序）
	ArtistPinyin          []string `bson:"artist_pinyin"`            // 表演者名称的拼音表示（用于搜索和排序）
	ArtistPinyinFull      string   `bson:"artist_pinyin_full"`       // 表演者名称的完整拼音表示（用于搜索和排序）
	AlbumArtistPinyin     []string `bson:"album_artist_pinyin"`      // 专辑艺术家名称的拼音表示（用于搜索和排序）
	AlbumArtistPinyinFull string   `bson:"album_artist_pinyin_full"` // 专辑艺术家名称的完整拼音表示（用于搜索和排序）
	Genre                 string   `bson:"genre"`                    // 音乐流派（如流行、摇滚等）
	Comment               string   `bson:"comment"`                  // 注释信息
	SongCount             int      `bson:"song_count"`               // 专辑中的歌曲总数
	Duration              float64  `bson:"duration"`                 // 专辑总时长（秒）
	Size                  int      `bson:"size"`                     // 专辑文件总大小（字节）
	MinYear               int      `bson:"min_year"`                 // 专辑中歌曲的最早发行年份
	MaxYear               int      `bson:"max_year"`                 // 专辑中歌曲的最晚发行年份
	Compilation           bool     `bson:"compilation"`              // 是否为合辑（多艺术家作品合集）

	// 关系ID索引
	ArtistID          string         `bson:"artist_id"`            // 艺术家在系统中的唯一标识符
	AlbumArtistID     string         `bson:"album_artist_id"`      // 专辑艺术家在系统中的唯一标识符
	AllArtistIDs      []ArtistIDPair `bson:"all_artist_ids"`       // 所有参与艺术家的唯一标识符列表
	AllAlbumArtistIDs []ArtistIDPair `bson:"all_album_artist_ids"` // 所有参与专辑艺术家的唯一标识符列表

	// 索引排序信息
	OrderAlbumName       string `bson:"order_album_name"`        // 排序用专辑名称（忽略冠词，便于排序）
	OrderAlbumArtistName string `bson:"order_album_artist_name"` // 排序用专辑艺术家名称（便于排序）
	SortAlbumName        string `bson:"sort_album_name"`         // 标准化专辑名称（去除非字母字符，便于排序）
	SortArtistName       string `bson:"sort_artist_name"`        // 标准化艺术家名称（便于排序）
	SortAlbumArtistName  string `bson:"sort_album_artist_name"`  // 标准化专辑艺术家名称（便于排序）

	// 视觉元素
	HasCoverArt    bool   `bson:"has_cover_art"`    // 是否包含专辑封面图
	EmbedArtPath   string `bson:"embed_art_path"`   // 嵌入式专辑封面图路径
	ImageFiles     string `bson:"image_files"`      // 专辑相关图片文件路径
	SmallImageURL  string `bson:"small_image_url"`  // 小尺寸封面图的 URL 地址
	MediumImageURL string `bson:"medium_image_url"` // 中等尺寸封面图的 URL 地址
	LargeImageURL  string `bson:"large_image_url"`  // 大尺寸封面图的 URL 地址

	// MusicBrainz元数据 (github.com/michiwend/gomusicbrainz)
	MBZAlbumID       string `bson:"mbz_album_id"`        // MusicBrainz 专辑唯一标识符
	MBZAlbumArtistID string `bson:"mbz_album_artist_id"` // MusicBrainz 专辑艺术家唯一标识符
	MBZAlbumType     string `bson:"mbz_album_type"`      // 专辑类型（如专辑、单曲等）
	MBZAlbumComment  string `bson:"mbz_album_comment"`   // MusicBrainz 专辑评论信息
	Paths            string `bson:"paths"`               // 抽象专辑所处文件系统目录路径
	Description      string `bson:"description"`         // 专辑描述信息
	CatalogNum       string `bson:"catalog_num"`         // 唱片目录编号（发行方的内部编号）

	// 外部信息
	ExternalURL           string    `bson:"external_url"`             // 外部链接 URL
	ExternalInfoUpdatedAt time.Time `bson:"external_info_updated_at"` // 外部信息最后更新时间
}

type ArtistIDPair struct {
	ArtistName string `bson:"artist_name"` // 艺术家名称
	ArtistID   string `bson:"artist_id"`   // 艺术家唯一 ID
}

type AlbumSongCounts struct {
	ID        primitive.ObjectID `bson:"_id"`        // 文档唯一标识符
	SongCount int                `bson:"song_count"` // 专辑中的歌曲总数
}
