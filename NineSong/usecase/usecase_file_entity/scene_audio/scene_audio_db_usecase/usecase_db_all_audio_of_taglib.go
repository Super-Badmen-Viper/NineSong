package scene_audio_db_usecase

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.senan.xyz/taglib"
)

func (e *AudioMetadataExtractorTaglib) Extract(
	path string,
	fileMetadata *domain_file_entity.FileMetadata,
) (
	*scene_audio_db_models.MediaFileMetadata,
	*scene_audio_db_models.AlbumMetadata,
	*scene_audio_db_models.ArtistMetadata,
	error,
) {
	if err := e.enrichFileMetadata(path, fileMetadata); err != nil {
		return nil, nil, nil, err
	}

	tags, err := taglib.ReadTags(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("标签解析失败[%s]: %w", path, err)
	}
	properties, err := taglib.ReadProperties(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("属性解析失败[%s]: %w", path, err)
	}

	now := time.Now().UTC()

	artistID := generateDeterministicID(e.getTagString(tags, taglib.Artist))
	albumID := generateDeterministicID(e.getTagString(tags, taglib.Album))
	albumArtistID := generateDeterministicID(e.getTagString(tags, taglib.AlbumArtist))

	mediaFile := e.buildMediaFile(path, tags, properties, fileMetadata, artistID, albumID, albumArtistID)
	album := e.buildAlbum(tags, now, artistID, albumID, albumArtistID)
	artist := e.buildArtist(tags, now, artistID)

	return mediaFile, album, artist, nil
}

func (e *AudioMetadataExtractorTaglib) enrichFileMetadata(path string, fm *domain_file_entity.FileMetadata) error {
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

func (e *AudioMetadataExtractorTaglib) buildMediaFile(
	path string,
	tags map[string][]string,
	properties taglib.Properties,
	fm *domain_file_entity.FileMetadata,
	artistID, albumID, albumArtistID primitive.ObjectID,
) *scene_audio_db_models.MediaFileMetadata {
	titleTag := e.getTagString(tags, taglib.Title)
	artistTag := e.getTagString(tags, taglib.Artist)
	albumTag := e.getTagString(tags, taglib.Album)
	albumArtist := e.getTagString(tags, taglib.AlbumArtist)

	currentTrack, totalTracks := e.getTagIntPair(tags, taglib.TrackNumber)
	currentDisc, totalDiscs := e.getTagIntPair(tags, taglib.DiscNumber)

	compilationArtist := e.hasMultipleArtists(artistTag)
	formattedArtist := artistTag
	var allArtistIDs []string
	if compilationArtist {
		formattedArtist, allArtistIDs = formatMultipleArtists(artistTag)
	}
	compilationAlbumArtist := e.hasMultipleArtists(albumArtist)
	formattedAlbumArtist := albumArtist
	var allAlbumArtistIDs []string
	if compilationAlbumArtist {
		formattedAlbumArtist, allAlbumArtistIDs = formatMultipleArtists(albumArtist)
	}

	title := e.cleanText(titleTag)
	artist := e.cleanText(formattedArtist)
	album := e.cleanText(albumTag)
	var parts []string
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
		Title:       titleTag,
		Artist:      formattedArtist,
		Album:       e.getTagString(tags, taglib.Album),
		AlbumArtist: formattedAlbumArtist,
		Genre:       e.getTagString(tags, taglib.Genre),
		Year:        e.getTagInt(tags, taglib.Date),
		TrackNumber: currentTrack,
		DiscNumber:  currentDisc,
		TotalTracks: totalTracks,
		TotalDiscs:  totalDiscs,
		Composer:    e.getTagString(tags, taglib.Composer),
		Comment:     e.getTagString(tags, taglib.Comment),
		Lyrics:      e.getTagString(tags, taglib.Lyrics),
		Compilation: compilationArtist,

		// 基础元数据: 关系ID索引
		ArtistID:          artistID.Hex(),
		AlbumID:           albumID.Hex(),
		AlbumArtistID:     albumArtistID.Hex(),
		AllArtistIDs:      allArtistIDs,
		AllAlbumArtistIDs: allAlbumArtistIDs,
		MvID:              "",
		KaraokeID:         "",
		LyricsID:          "",

		// 基础元数据: 索引排序信息
		Index:                0,
		SortTitle:            e.getSortTitle(titleTag),
		SortAlbumName:        e.getSortAlbumName(albumTag),
		SortArtistName:       e.getSortArtistName(formattedArtist),
		SortAlbumArtistName:  e.getSortAlbumArtistName(formattedAlbumArtist),
		OrderTitle:           e.getOrderTitle(titleTag),
		OrderArtistName:      e.getOrderArtistName(formattedArtist),
		OrderAlbumName:       e.getOrderAlbumName(albumTag),
		OrderAlbumArtistName: e.getOrderAlbumArtistName(formattedAlbumArtist),

		// 音频分析 (综合)
		SampleRate: int(properties.SampleRate),
		Duration:   float64(properties.Length),
		BitRate:    int(properties.Bitrate),
		Channels:   int(properties.Channels),
	}
}

func (e *AudioMetadataExtractorTaglib) buildAlbum(
	tags map[string][]string,
	now time.Time,
	artistID, albumID, albumArtistID primitive.ObjectID,
) *scene_audio_db_models.AlbumMetadata {
	artistTag := e.getTagString(tags, taglib.Artist)
	albumTag := e.getTagString(tags, taglib.Album)
	albumArtist := e.getTagString(tags, taglib.AlbumArtist)

	compilationArtist := e.hasMultipleArtists(artistTag)
	formattedArtist := artistTag
	var allArtistIDs []string
	if compilationArtist {
		formattedArtist, allArtistIDs = formatMultipleArtists(artistTag)
	}
	compilationAlbumArtist := e.hasMultipleArtists(albumArtist)
	formattedAlbumArtist := albumArtist
	var allAlbumArtistIDs []string
	if compilationAlbumArtist {
		formattedAlbumArtist, allAlbumArtistIDs = formatMultipleArtists(albumArtist)
	}

	return &scene_audio_db_models.AlbumMetadata{
		// 系统保留字段 (综合)
		ID:        albumID,
		CreatedAt: now,
		UpdatedAt: now,

		// 基础元数据 (综合)
		Name:        e.getTagString(tags, taglib.Album),
		Artist:      formattedArtist,
		AlbumArtist: formattedAlbumArtist,
		Genre:       e.getTagString(tags, taglib.Genre),
		Comment:     e.getTagString(tags, taglib.Comment),
		Compilation: compilationArtist,
		SongCount:   0,
		Duration:    0,
		Size:        0,
		MinYear:     e.getTagInt(tags, taglib.Date),
		MaxYear:     e.getTagInt(tags, taglib.Date),

		// 关系ID索引
		ArtistID:          artistID.Hex(),
		AlbumArtistID:     albumArtistID.Hex(),
		AllArtistIDs:      allArtistIDs,
		AllAlbumArtistIDs: allAlbumArtistIDs,

		// 索引排序信息
		SortAlbumName:        e.getSortAlbumName(albumTag),
		SortArtistName:       e.getSortArtistName(formattedArtist),
		SortAlbumArtistName:  e.getSortAlbumArtistName(formattedAlbumArtist),
		OrderAlbumName:       e.getOrderAlbumName(albumTag),
		OrderAlbumArtistName: e.getOrderAlbumArtistName(formattedAlbumArtist),
	}
}

func (e *AudioMetadataExtractorTaglib) buildArtist(
	tags map[string][]string,
	now time.Time,
	artistID primitive.ObjectID,
) *scene_audio_db_models.ArtistMetadata {
	artistTag := e.getTagString(tags, taglib.Artist)

	return &scene_audio_db_models.ArtistMetadata{
		// 系统保留字段 (综合)
		ID:        artistID,
		CreatedAt: now,
		UpdatedAt: now,

		// 基础元数据 (综合)
		Name:       artistTag,
		AlbumCount: 0,
		SongCount:  0,
		Size:       0,

		// 索引排序信息
		SortArtistName:  e.getSortArtistName(artistTag),
		OrderArtistName: e.getOrderArtistName(artistTag),
	}
}

type AudioMetadataExtractorTaglib struct {
	mediaID primitive.ObjectID
}

func (e *AudioMetadataExtractorTaglib) hasMultipleArtists(artist string) bool {
	separators := []string{"/", ",", "&", ";", "//"}
	artist = strings.TrimSpace(artist)
	for _, sep := range separators {
		if strings.Contains(artist, sep) {
			return true
		}
	}
	return false
}

func formatMultipleArtists(artistTag string) (string, []string) {
	separators := []string{"//", "/", ",", "&", ";"}
	currentList := []string{artistTag}

	for _, sep := range separators {
		var newList []string
		for _, item := range currentList {
			parts := strings.Split(item, sep)
			for _, p := range parts {
				if p != "" {
					newList = append(newList, p)
				}
			}
		}
		currentList = newList
	}

	uniqueArtists := make(map[string]struct{})
	var dedupedList []string
	for _, artist := range currentList {
		if _, exists := uniqueArtists[artist]; !exists {
			uniqueArtists[artist] = struct{}{}
			dedupedList = append(dedupedList, artist)
		}
	}

	joinedArtists := strings.Join(dedupedList, "、")

	allArtistIDs := make([]string, len(dedupedList))
	for i, artist := range dedupedList {
		allArtistIDs[i] = generateDeterministicID(artist).Hex()
	}

	return joinedArtists, allArtistIDs
}

func (e *AudioMetadataExtractorTaglib) cleanText(text string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	cleaned := reg.ReplaceAllString(text, "")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

func (e *AudioMetadataExtractorTaglib) getSortTitle(title string) string {
	return e.removeArticles(title)
}

func (e *AudioMetadataExtractorTaglib) getSortAlbumName(album string) string {
	return e.removeNonAlphabeticChars(album)
}

func (e *AudioMetadataExtractorTaglib) getSortArtistName(artist string) string {
	return e.removeNonAlphabeticChars(artist)
}

func (e *AudioMetadataExtractorTaglib) getSortAlbumArtistName(albumArtist string) string {
	return e.removeNonAlphabeticChars(albumArtist)
}

func (e *AudioMetadataExtractorTaglib) getOrderTitle(title string) string {
	return e.removePrefixes(title)
}

func (e *AudioMetadataExtractorTaglib) getOrderArtistName(artist string) string {
	return e.removePrefixes(artist)
}

func (e *AudioMetadataExtractorTaglib) getOrderAlbumName(album string) string {
	return e.removeArticles(album)
}

func (e *AudioMetadataExtractorTaglib) getOrderAlbumArtistName(albumArtist string) string {
	return e.removeArticles(albumArtist)
}

func (e *AudioMetadataExtractorTaglib) removeArticles(s string) string {
	articlesPattern := regexp.MustCompile(`(?i)^(the|a|an|a\s|an\s|the\s)`)
	return articlesPattern.ReplaceAllString(strings.ToLower(s), "")
}

func (e *AudioMetadataExtractorTaglib) removeNonAlphabeticChars(s string) string {
	nonAlphaPattern := regexp.MustCompile(`[^a-zA-Z]`)
	return nonAlphaPattern.ReplaceAllString(strings.ToLower(s), "")
}

func (e *AudioMetadataExtractorTaglib) removePrefixes(s string) string {
	prefixesPattern := regexp.MustCompile(`(?i)^(of|in|to|for|on|at|from|by|with|and\s)`)
	return prefixesPattern.ReplaceAllString(strings.ToLower(s), "")
}

func (e *AudioMetadataExtractorTaglib) getTagString(tags map[string][]string, key string) string {
	values := tags[key]
	if len(values) > 0 {
		return strings.TrimSpace(values[0])
	}
	return ""
}

func (e *AudioMetadataExtractorTaglib) getTagInt(tags map[string][]string, key string) int {
	value := e.getTagString(tags, key)
	if value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return 0
}

func (e *AudioMetadataExtractorTaglib) getTagIntPair(tags map[string][]string, key string) (int, int) {
	value := e.getTagString(tags, key)
	if value != "" {
		var current, total int
		if _, err := fmt.Sscanf(value, "%d/%d", &current, &total); err == nil {
			return current, total
		}
	}
	return 0, 0
}
