package scene_audio_db_usecase

import (
	"crypto/sha256"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/dhowden/tag"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (e *AudioMetadataExtractorTag) Extract(
	path string,
	fileMetadata *domain_file_entity.FileMetadata,
) (
	*scene_audio_db_models.MediaFileMetadata,
	*scene_audio_db_models.AlbumMetadata,
	*scene_audio_db_models.ArtistMetadata,
	tag.Metadata,
	error,
) {
	e.mediaID = primitive.NewObjectID()
	if err := e.enrichFileMetadata(path, fileMetadata); err != nil {
		return nil, nil, nil, nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("文件访问失败[%s]: %w", path, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("文件关闭失败[%s]: %v", path, err)
		}
	}(file)

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("标签解析失败[%s]: %w", path, err)
	}
	now := time.Now().UTC()
	rawTags := metadata.Raw()

	artistID := generateDeterministicID(metadata.Artist())
	albumID := generateDeterministicID(metadata.Album())
	albumArtistID := generateDeterministicID(metadata.AlbumArtist())

	mediaFile := e.buildMediaFile(
		path,
		metadata,
		rawTags,
		fileMetadata,
		artistID,
		albumID,
		albumArtistID,
	)
	album := e.buildAlbum(
		metadata,
		rawTags,
		now,
		artistID,
		albumID,
		albumArtistID,
	)
	artist := e.buildArtist(
		metadata,
		rawTags,
		artistID,
	)
	return mediaFile, album, artist, metadata, nil
}

func (e *AudioMetadataExtractorTag) enrichFileMetadata(path string, fm *domain_file_entity.FileMetadata) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("文件关闭失败[%s]: %v", path, err)
		}
	}(file)

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("校验和计算失败: %w", err)
	}
	fm.Checksum = fmt.Sprintf("%x", hash.Sum(nil))

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("文件状态获取失败: %w", err)
	}

	fm.FilePath = path
	fm.Size = info.Size()
	fm.ModTime = info.ModTime().UTC()
	fm.FileType = domain_file_entity.Audio

	if fm.CreatedAt.IsZero() {
		fm.CreatedAt = time.Now().UTC()
	}
	fm.UpdatedAt = time.Now().UTC()

	return nil
}

func (e *AudioMetadataExtractorTag) buildMediaFile(
	path string,
	m tag.Metadata,
	rawTags map[string]interface{},
	fm *domain_file_entity.FileMetadata,
	artistID, albumID, albumArtistID primitive.ObjectID,
) *scene_audio_db_models.MediaFileMetadata {
	currentTrack, totalTracks := m.Track()
	currentDisc, totalDiscs := m.Disc()

	compilation := e.hasMultipleArtists(m.Artist())
	formattedArtist := m.Artist()
	if compilation {
		formattedArtist = regexp.MustCompile(`[/,&;\\s]+`).ReplaceAllString(strings.TrimSpace(m.Artist()), " ")
		formattedArtist = regexp.MustCompile(`\s+`).ReplaceAllString(formattedArtist, "、")
	}

	title := e.cleanText(m.Title())
	artist := e.cleanText(formattedArtist)
	album := e.cleanText(m.Album())
	parts := []string{}
	if title != "" {
		parts = append(parts, title)
	}
	if artist != "" {
		parts = append(parts, artist)
	}
	if album != "" {
		parts = append(parts, album)
	}
	fullText := strings.Join(parts, " ")

	return &scene_audio_db_models.MediaFileMetadata{
		// 系统保留字段 (综合)
		ID:        e.mediaID,
		CreatedAt: fm.CreatedAt,
		UpdatedAt: fm.UpdatedAt,
		FullText:  fullText,
		Path:      fm.FilePath,
		Suffix:    strings.ToLower(strings.TrimPrefix(filepath.Ext(path), ".")),
		Size:      int(fm.Size),

		// 基础元数据 (github.com/dhowden/tag、go.senan.xyz/taglib)
		Title:       m.Title(),
		SubTitle:    e.getTagString(rawTags, "sub_title"),
		Artist:      formattedArtist,
		Album:       m.Album(),
		AlbumArtist: m.AlbumArtist(),
		Genre:       m.Genre(),
		Year:        m.Year(),
		TrackNumber: currentTrack,
		DiscNumber:  currentDisc,
		TotalTracks: totalTracks,
		TotalDiscs:  totalDiscs,
		Composer:    m.Composer(),
		Comment:     m.Comment(),
		Lyrics:      m.Lyrics(),
		Compilation: compilation,

		// 基础元数据: 关系ID索引
		ArtistID:      artistID.Hex(),
		AlbumID:       albumID.Hex(),
		AlbumArtistID: albumArtistID.Hex(),
		MvID:          "",
		KaraokeID:     "",
		LyricsID:      "",

		// 基础元数据: 索引排序信息
		Index:                0,
		SortTitle:            e.getSortTitle(m.Title()),
		SortAlbumName:        e.getSortAlbumName(m.Album()),
		SortArtistName:       e.getSortArtistName(m.Artist()),
		SortAlbumArtistName:  e.getSortAlbumArtistName(m.AlbumArtist()),
		OrderTitle:           e.getOrderTitle(m.Title()),
		OrderArtistName:      e.getOrderArtistName(m.Artist()),
		OrderAlbumName:       e.getOrderAlbumName(m.Album()),
		OrderAlbumArtistName: e.getOrderAlbumArtistName(m.AlbumArtist()),
	}
}

func (e *AudioMetadataExtractorTag) buildAlbum(
	m tag.Metadata,
	rawTags map[string]interface{},
	now time.Time,
	artistID, albumID, albumArtistID primitive.ObjectID,
) *scene_audio_db_models.AlbumMetadata {
	return &scene_audio_db_models.AlbumMetadata{
		ID:            albumID,
		ArtistID:      artistID.Hex(),
		AlbumArtistID: albumArtistID.Hex(),

		Name:        m.Album(),
		Artist:      m.Artist(),
		AlbumArtist: m.AlbumArtist(),
		Genre:       m.Genre(),
		MinYear:     m.Year(),
		MaxYear:     m.Year(),
		CreatedAt:   now,
		UpdatedAt:   now,

		EmbedArtPath:          "",
		Compilation:           false,
		SongCount:             0,
		Duration:              0,
		FullText:              "",
		Size:                  0,
		CatalogNum:            e.getTagString(rawTags, "catalognum"),
		Comment:               m.Comment(),
		AllArtistIDs:          "",
		ImageFiles:            "",
		Paths:                 "",
		Description:           "",
		SmallImageURL:         "",
		MediumImageURL:        "",
		LargeImageURL:         "",
		ExternalURL:           "",
		ExternalInfoUpdatedAt: time.Time{},
	}
}

func (e *AudioMetadataExtractorTag) buildArtist(
	m tag.Metadata,
	rawTags map[string]interface{},
	artistID primitive.ObjectID,
) *scene_audio_db_models.ArtistMetadata {
	return &scene_audio_db_models.ArtistMetadata{
		ID:          artistID,
		Name:        m.Artist(),
		MBZArtistID: e.getTagString(rawTags, "musicbrainz_artistid"),
		// 默认空值字段
		AlbumCount:            0,
		FullText:              "",
		OrderArtistName:       e.getTagString(rawTags, "order_artist_name"),
		SortArtistName:        e.getTagString(rawTags, "sort_artist_name"),
		SongCount:             0,
		Size:                  0,
		Biography:             e.getTagString(rawTags, "biography"),
		SmallImageURL:         e.getTagString(rawTags, "artist_image_small"),
		MediumImageURL:        e.getTagString(rawTags, "artist_image_medium"),
		LargeImageURL:         e.getTagString(rawTags, "artist_image_large"),
		ExternalURL:           e.getTagString(rawTags, "artist_external_url"),
		ExternalInfoUpdatedAt: time.Time{},
	}
}

type AudioMetadataExtractorTag struct {
	mediaID primitive.ObjectID
}

func generateDeterministicID(seed string) primitive.ObjectID {
	hash := sha256.Sum256([]byte(seed))
	return primitive.ObjectID(hash[:12])
}

func (e *AudioMetadataExtractorTag) getTagString(tags map[string]interface{}, key string) string {
	if val, ok := tags[key]; ok {
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s)
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func (e *AudioMetadataExtractorTag) getTagInt(tags map[string]interface{}, key string) int {
	if s := e.getTagString(tags, key); s != "" {
		var result int
		if _, err := fmt.Sscanf(s, "%d", &result); err == nil {
			return result
		}
	}
	return 0
}

func (e *AudioMetadataExtractorTag) getTagFloat(tags map[string]interface{}, key string) float64 {
	if s := e.getTagString(tags, key); s != "" {
		var result float64
		if _, err := fmt.Sscanf(s, "%f", &result); err == nil {
			return result
		}
	}
	return 0.0
}

func (e *AudioMetadataExtractorTag) hasMultipleArtists(artist string) bool {
	separators := []string{"//", "/", ",", "&", ";", " "}
	artist = strings.TrimSpace(artist)
	for _, sep := range separators {
		if strings.Contains(artist, sep) {
			return true
		}
	}
	return false
}

func (e *AudioMetadataExtractorTag) cleanText(text string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	cleaned := reg.ReplaceAllString(text, "")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

func (e *AudioMetadataExtractorTag) getSortTitle(title string) string {
	return e.removeArticles(title)
}

func (e *AudioMetadataExtractorTag) getSortAlbumName(album string) string {
	return e.removeNonAlphabeticChars(album)
}

func (e *AudioMetadataExtractorTag) getSortArtistName(artist string) string {
	return e.removeNonAlphabeticChars(artist)
}

func (e *AudioMetadataExtractorTag) getSortAlbumArtistName(albumArtist string) string {
	return e.removeNonAlphabeticChars(albumArtist)
}

func (e *AudioMetadataExtractorTag) getOrderTitle(title string) string {
	return e.removePrefixes(title)
}

func (e *AudioMetadataExtractorTag) getOrderArtistName(artist string) string {
	return e.removePrefixes(artist)
}

func (e *AudioMetadataExtractorTag) getOrderAlbumName(album string) string {
	return e.removeArticles(album)
}

func (e *AudioMetadataExtractorTag) getOrderAlbumArtistName(albumArtist string) string {
	return e.removeArticles(albumArtist)
}

func (e *AudioMetadataExtractorTag) removeArticles(s string) string {
	articlesPattern := regexp.MustCompile(`(?i)^(the|a|an|a\s|an\s|the\s)`)
	return articlesPattern.ReplaceAllString(strings.ToLower(s), "")
}

func (e *AudioMetadataExtractorTag) removeNonAlphabeticChars(s string) string {
	nonAlphaPattern := regexp.MustCompile(`[^a-zA-Z]`)
	return nonAlphaPattern.ReplaceAllString(strings.ToLower(s), "")
}

func (e *AudioMetadataExtractorTag) removePrefixes(s string) string {
	prefixesPattern := regexp.MustCompile(`(?i)^(of|in|to|for|on|at|from|by|with|and\s)`)
	return prefixesPattern.ReplaceAllString(strings.ToLower(s), "")
}
