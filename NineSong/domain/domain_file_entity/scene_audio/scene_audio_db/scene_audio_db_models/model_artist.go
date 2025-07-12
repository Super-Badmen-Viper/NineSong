package scene_audio_db_models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ArtistMetadata struct {
	// 系统保留字段 (综合)
	ID        primitive.ObjectID `bson:"_id"`        // 文档唯一标识符
	CreatedAt time.Time          `bson:"created_at"` // 文档创建时间
	UpdatedAt time.Time          `bson:"updated_at"` // 文档最后更新时间

	// 基础元数据 (综合)
	Name            string   `bson:"name"`
	NamePinyin      []string `bson:"name_pinyin"`      // 艺术家名称拼音
	NamePinyinFull  string   `bson:"name_pinyin_full"` // 艺术家名称完整拼音
	AlbumCount      int      `bson:"album_count"`
	GuestAlbumCount int      `bson:"guest_album_count"`
	SongCount       int      `bson:"song_count"`
	GuestSongCount  int      `bson:"guest_song_count"`
	CueCount        int      `bson:"cue_count"`
	GuestCueCount   int      `bson:"guest_cue_count"`
	Size            int      `bson:"size"`
	Compilation     bool     `bson:"compilation"` // 是否为合辑（多艺术家作品合集）

	// 关系ID索引
	AllArtistIDs []ArtistIDPair `bson:"all_artist_ids"` // 所有参与艺术家的唯一标识符列表

	// 索引排序信息
	OrderArtistName string `bson:"order_artist_name"`
	SortArtistName  string `bson:"sort_artist_name"`

	// 视觉元素
	HasCoverArt    bool   `bson:"has_cover_art"`
	EmbedArtPath   string `bson:"embed_art_path"` // 嵌入式艺术家封面图路径
	ImageFiles     string `bson:"image_files"`    // 专辑相关图片文件路径
	SmallImageURL  string `bson:"small_image_url"`
	MediumImageURL string `bson:"medium_image_url"`
	LargeImageURL  string `bson:"large_image_url"`

	// MusicBrainz元数据 (github.com/michiwend/gomusicbrainz)
	MBZArtistID       string   `bson:"mbz_artist_id"`
	Description       string   `bson:"description"` // 艺术家描述信息
	Biography         string   `bson:"biography"`
	SimilarArtistsIDs []string `bson:"similar_artists_ids"`

	// 外部信息
	ExternalURL           string    `bson:"external_url"`             // 外部链接 URL
	ExternalInfoUpdatedAt time.Time `bson:"external_info_updated_at"` // 外部信息最后更新时间
}
