package scene_audio_db_usecase

import (
	"crypto/sha256"
	"fmt"
	"github.com/mozillazg/go-pinyin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.senan.xyz/taglib"
)

func (e *AudioMetadataExtractorTaglib) Extract(
	path string, libraryPath string,
	fileMetadata *domain_file_entity.FileMetadata,
	res *scene_audio_db_models.CueConfig,
) (
	*scene_audio_db_models.MediaFileMetadata,
	*scene_audio_db_models.AlbumMetadata,
	[]*scene_audio_db_models.ArtistMetadata,
	*scene_audio_db_models.MediaFileCueMetadata,
	error,
) {
	if err := e.enrichFileMetadata(path, libraryPath, fileMetadata); err != nil {
		return nil, nil, nil, nil, err
	}

	var tags map[string][]string
	tags, err := taglib.ReadTags(path)
	if err != nil {
		if res == nil {
			return nil, nil, nil, nil, fmt.Errorf("标签解析失败[%s]: %w", path, err)
		}
		if tags == nil {
			tags = make(map[string][]string)
		}
	}
	properties, err := taglib.ReadProperties(path)
	if err != nil {
		if res == nil {
			return nil, nil, nil, nil, fmt.Errorf("属性解析失败[%s]: %w", path, err)
		}
	}

	now := time.Now().UTC()

	suffix := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

	var artistID, albumID, albumArtistID primitive.ObjectID
	var artistTag, albumArtistTag, albumTag string
	var artistSortTag, albumArtistSortTag, albumSortTag string

	var mediaFile *scene_audio_db_models.MediaFileMetadata
	var compilationArtist bool
	var formattedArtist string
	var allArtistIDs []scene_audio_db_models.ArtistIDPair
	var formattedAlbumArtist string
	var allAlbumArtistIDs []scene_audio_db_models.ArtistIDPair
	var albumPinyin []string
	var artistPinyin []string
	var albumArtistPinyin []string

	var album *scene_audio_db_models.AlbumMetadata

	var artist []*scene_audio_db_models.ArtistMetadata

	var mediaFileCue *scene_audio_db_models.MediaFileCueMetadata

	if res != nil && res.CuePath != "" && res.AudioPath != "" {
		globalMeta, tracks, err := parseCueFile(res.CuePath)
		if err != nil {
			log.Printf("CUE解析警告: %v, 使用标签元数据", err)
		} else {
			mediaFileCue, albumTag, artistTag, albumArtistTag, allArtistIDs = e.buildMediaFileCue(
				tags, properties, fileMetadata,
				globalMeta, mediaFileCue,
				tracks, suffix,
				albumTag, artistTag, albumArtistTag,
			)
			mediaFileCue.CueResources = scene_audio_db_models.CueConfig{
				CuePath:    res.CuePath,
				AudioPath:  res.AudioPath,
				BackImage:  res.BackImage,
				CoverImage: res.CoverImage,
				DiscImage:  res.DiscImage,
				ListFile:   res.ListFile,
				LogFile:    res.LogFile,
			}
		}
	} else {
		artistTag = e.getTagString(tags, taglib.Artist)
		albumArtistTag = e.getTagString(tags, taglib.AlbumArtist)
		albumTag = e.getTagString(tags, taglib.Album)

		artistSortTag = e.getTagString(tags, taglib.ArtistSort)
		albumArtistSortTag = e.getTagString(tags, taglib.AlbumArtistSort)
		albumSortTag = e.getTagString(tags, taglib.AlbumSort)

		if suffix == "m4a" {
			if len(artistSortTag) > len(artistTag) {
				artistTag = artistSortTag
			}
			if len(albumArtistSortTag) > len(albumArtistTag) {
				albumArtistTag = albumArtistSortTag
			}
			if len(albumSortTag) > len(albumTag) {
				albumTag = albumSortTag
			}
		}
	}

	albumID = generateDeterministicID(artistTag + albumTag)
	artistID = generateDeterministicID(artistTag)
	albumArtistID = generateDeterministicID(albumArtistTag)

	mediaFile,
		compilationArtist,
		formattedArtist, allArtistIDs,
		formattedAlbumArtist, allAlbumArtistIDs,
		albumPinyin, artistPinyin, albumArtistPinyin =
		e.buildMediaFile(
			tags, properties, fileMetadata,
			artistID, albumID, albumArtistID,
			suffix,
			albumTag, artistTag, albumArtistTag,
		)

	album = e.buildAlbum(
		tags, now, artistID, albumID, albumArtistID,
		compilationArtist,
		formattedArtist, allArtistIDs,
		formattedAlbumArtist, allAlbumArtistIDs,
		albumPinyin, artistPinyin, albumArtistPinyin,
	)

	// 这是NineSong面向音乐场景的业务特性，默认为单体艺术家，并探索其相关业务逻辑的用户友好性与数据管理增强
	if compilationArtist {
		if mediaFile != nil {
			for index, artistIDPair := range mediaFile.AllArtistIDs {
				if index == 0 {
					// 出现复合艺术家的情况，那么单曲-专辑的艺术家ID，应该为主艺术家的ID，而不是复合艺术家的复合ID
					mediaFile.ArtistID = artistIDPair.ArtistID
					album.ArtistID = artistIDPair.ArtistID
				}
				artistId, _ := primitive.ObjectIDFromHex(artistIDPair.ArtistID)
				artist = append(
					artist, e.buildArtist(
						now, artistId, artistIDPair.ArtistName,
						compilationArtist,
						formattedArtist, allArtistIDs,
						artistPinyin,
					),
				)
			}
		}
	} else {
		artist = append(
			artist, e.buildArtist(
				now, artistID, "",
				compilationArtist,
				formattedArtist, allArtistIDs,
				artistPinyin,
			),
		)
	}

	if mediaFileCue != nil {
		return nil, nil, artist, mediaFileCue, nil
	}
	return mediaFile, album, artist, nil, nil
}

func (e *AudioMetadataExtractorTaglib) enrichFileMetadata(
	path string, libraryPath string,
	fileMetadata *domain_file_entity.FileMetadata,
) error {
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
	fileMetadata.Checksum = fmt.Sprintf("%x", hash.Sum(nil))

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("文件状态获取失败: %w", err)
	}

	fileMetadata.FilePath = path
	fileMetadata.FileName = filepath.Base(path)
	fileMetadata.LibraryPath = libraryPath
	fileMetadata.Size = info.Size()
	fileMetadata.ModTime = info.ModTime().UTC()
	fileMetadata.FileType = domain_file_entity.Audio

	if fileMetadata.CreatedAt.IsZero() {
		fileMetadata.CreatedAt = time.Now().UTC()
	}
	fileMetadata.UpdatedAt = time.Now().UTC()

	return nil
}

func (e *AudioMetadataExtractorTaglib) buildMediaFileCue(
	tags map[string][]string,
	properties taglib.Properties,
	fileMetadata *domain_file_entity.FileMetadata,
	globalMeta map[string]string,
	mediaFileCue *scene_audio_db_models.MediaFileCueMetadata,
	tracks []scene_audio_db_models.CueTrack,
	suffix string,
	albumTag, artistTag, albumArtistTag string,
) (
	*scene_audio_db_models.MediaFileCueMetadata,
	string, string, string,
	[]scene_audio_db_models.ArtistIDPair,
) {
	if genre, ok := globalMeta["GENRE"]; ok {
		tags["GENRE"] = []string{genre}
	}
	if date, ok := globalMeta["DATE"]; ok {
		if year, err := strconv.Atoi(date); err == nil {
			tags["DATE"] = []string{strconv.Itoa(year)}
		}
	}
	if comment, ok := globalMeta["COMMENT"]; ok {
		tags["Comment"] = []string{comment}
	}
	if title, ok := globalMeta["TITLE"]; ok {
		tags["Title"] = []string{title}
	}
	if performer, ok := globalMeta["PERFORMER"]; ok {
		albumArtistTag = performer
		artistTag = performer
	}
	if title, ok := globalMeta["TITLE"]; ok {
		albumTag = title
	}

	artistID := generateDeterministicID(artistTag)

	compilationArtist := e.hasMultipleArtists(artistTag)
	formattedArtist := artistTag
	var allArtistIDs []scene_audio_db_models.ArtistIDPair
	if compilationArtist {
		formattedArtist, allArtistIDs = formatMultipleArtists(artistTag)
	} else {
		allArtistIDs = append(allArtistIDs, scene_audio_db_models.ArtistIDPair{
			ArtistName: artistTag,
			ArtistID:   artistID.Hex(),
		})
	}

	titleText := e.cleanText(globalMeta["TITLE"])
	artistText := e.cleanText(formattedArtist)
	albumText := e.cleanText(albumTag)
	var parts []string
	if titleText != "" {
		parts = append(parts, titleText)
	}
	if artistText != "" {
		parts = append(parts, artistText)
	}
	if albumText != "" {
		parts = append(parts, albumText)
	}
	fullText := strings.Join(parts, " ")

	mediaFileCue = &scene_audio_db_models.MediaFileCueMetadata{
		// 系统保留字段 (综合)
		ID:          fileMetadata.ID,
		CreatedAt:   fileMetadata.CreatedAt,
		UpdatedAt:   fileMetadata.UpdatedAt,
		FullText:    fullText,
		Path:        fileMetadata.FilePath,
		Suffix:      suffix,
		Size:        int(fileMetadata.Size),
		FileName:    fileMetadata.FileName,
		LibraryPath: fileMetadata.LibraryPath,
	}

	mediaFileCue.CueTracks = tracks

	mediaFileCue.Rem = scene_audio_db_models.CueREM{
		GENRE:   globalMeta["GENRE"],
		DATE:    globalMeta["DATE"],
		DISCID:  globalMeta["DISCID"],
		COMMENT: globalMeta["COMMENT"],
	}
	mediaFileCue.Performer = globalMeta["PERFORMER"]
	mediaFileCue.PerformerID = generateDeterministicID(globalMeta["PERFORMER"]).Hex()
	mediaFileCue.Title = globalMeta["TITLE"]
	mediaFileCue.File = scene_audio_db_models.CueFile{
		FilePath: globalMeta["FILE"],
	}
	mediaFileCue.Catalog = globalMeta["CATALOG"]
	mediaFileCue.SongWriter = globalMeta["SONGWRITER"]

	mediaFileCue.CueSampleRate = int(properties.SampleRate)
	mediaFileCue.CueDuration = float64(properties.Length)
	mediaFileCue.CueBitRate = int(properties.Bitrate)
	mediaFileCue.CueChannels = int(properties.Channels)

	return mediaFileCue, albumTag, formattedArtist, albumArtistTag, allArtistIDs
}

func (e *AudioMetadataExtractorTaglib) buildMediaFile(
	tags map[string][]string,
	properties taglib.Properties,
	fileMetadata *domain_file_entity.FileMetadata,
	artistID, albumID, albumArtistID primitive.ObjectID,
	suffix string,
	albumTag, artistTag, albumArtistTag string,
) (
	*scene_audio_db_models.MediaFileMetadata,
	bool,
	string, []scene_audio_db_models.ArtistIDPair,
	string, []scene_audio_db_models.ArtistIDPair,
	[]string, []string, []string,
) {
	titleTag := e.getTagString(tags, taglib.Title)

	currentTrack, totalTracks := e.getTagIntPair(tags, taglib.TrackNumber)
	currentDisc, totalDiscs := e.getTagIntPair(tags, taglib.DiscNumber)

	compilationArtist := e.hasMultipleArtists(artistTag)
	formattedArtist := artistTag
	var allArtistIDs []scene_audio_db_models.ArtistIDPair
	if compilationArtist {
		formattedArtist, allArtistIDs = formatMultipleArtists(artistTag)
	} else {
		allArtistIDs = append(allArtistIDs, scene_audio_db_models.ArtistIDPair{
			ArtistName: artistTag,
			ArtistID:   artistID.Hex(),
		})
	}

	compilationAlbumArtist := e.hasMultipleArtists(albumArtistTag)
	formattedAlbumArtist := albumArtistTag
	var allAlbumArtistIDs []scene_audio_db_models.ArtistIDPair
	if compilationAlbumArtist {
		formattedAlbumArtist, allAlbumArtistIDs = formatMultipleArtists(albumArtistTag)
	} else {
		allAlbumArtistIDs = append(allAlbumArtistIDs, scene_audio_db_models.ArtistIDPair{
			ArtistName: albumArtistTag,
			ArtistID:   albumArtistID.Hex(),
		})
	}

	titleText := e.cleanText(titleTag)
	artistText := e.cleanText(formattedArtist)
	albumText := e.cleanText(albumTag)
	var parts []string
	if titleText != "" {
		parts = append(parts, titleText)
	}
	if artistText != "" {
		parts = append(parts, artistText)
	}
	if albumText != "" {
		parts = append(parts, albumText)
	}
	fullText := strings.Join(parts, " ")

	titlePinyin := pinyin.LazyConvert(titleTag, nil)
	albumPinyin := pinyin.LazyConvert(albumTag, nil)
	artistPinyin := pinyin.LazyConvert(formattedArtist, nil)
	albumArtistPinyin := pinyin.LazyConvert(albumArtistTag, nil)

	return &scene_audio_db_models.MediaFileMetadata{
			// 系统保留字段 (综合)
			ID:          e.mediaID,
			CreatedAt:   fileMetadata.CreatedAt,
			UpdatedAt:   fileMetadata.UpdatedAt,
			FullText:    fullText,
			Path:        fileMetadata.FilePath,
			Suffix:      suffix,
			Size:        int(fileMetadata.Size),
			FileName:    fileMetadata.FileName,
			LibraryPath: fileMetadata.LibraryPath,

			// 基础元数据 (github.com/dhowden/tag、go.senan.xyz/taglib)
			Title:       titleTag,
			Artist:      formattedArtist,
			Album:       albumTag,
			AlbumArtist: formattedAlbumArtist,

			TitlePinyin:       titlePinyin,
			AlbumPinyin:       albumPinyin,
			ArtistPinyin:      artistPinyin,
			AlbumArtistPinyin: albumArtistPinyin,

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
		},
		compilationArtist,
		formattedArtist, allArtistIDs,
		formattedAlbumArtist, allAlbumArtistIDs,
		albumPinyin, artistPinyin, albumArtistPinyin
}

func (e *AudioMetadataExtractorTaglib) buildAlbum(
	tags map[string][]string,
	now time.Time,
	artistID, albumID, albumArtistID primitive.ObjectID,
	compilationArtist bool,
	formattedArtist string, allArtistIDs []scene_audio_db_models.ArtistIDPair,
	formattedAlbumArtist string, allAlbumArtistIDs []scene_audio_db_models.ArtistIDPair,
	albumPinyin, artistPinyin, albumArtistPinyin []string,
) *scene_audio_db_models.AlbumMetadata {
	albumTag := e.getTagString(tags, taglib.Album)

	return &scene_audio_db_models.AlbumMetadata{
		// 系统保留字段 (综合)
		ID:        albumID,
		CreatedAt: now,
		UpdatedAt: now,

		// 基础元数据 (综合)
		Name:              e.getTagString(tags, taglib.Album),
		Artist:            formattedArtist,
		AlbumArtist:       formattedAlbumArtist,
		NamePinyin:        albumPinyin,
		ArtistPinyin:      artistPinyin,
		AlbumArtistPinyin: albumArtistPinyin,
		Genre:             e.getTagString(tags, taglib.Genre),
		Comment:           e.getTagString(tags, taglib.Comment),
		SongCount:         0,
		Duration:          0,
		Size:              0,
		MinYear:           e.getTagInt(tags, taglib.Date),
		MaxYear:           e.getTagInt(tags, taglib.Date),
		Compilation:       compilationArtist,

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
	now time.Time,
	artistID primitive.ObjectID,
	artistName string,
	compilationArtist bool,
	formattedArtist string, allArtistIDs []scene_audio_db_models.ArtistIDPair,
	artistPinyin []string,
) *scene_audio_db_models.ArtistMetadata {
	var artistTag string
	if artistName != "" {
		artistTag = artistName
	} else {
		artistTag = formattedArtist
	}

	return &scene_audio_db_models.ArtistMetadata{
		// 系统保留字段 (综合)
		ID:        artistID,
		CreatedAt: now,
		UpdatedAt: now,

		// 基础元数据 (综合)
		Name:        artistTag,
		NamePinyin:  artistPinyin,
		AlbumCount:  0,
		SongCount:   0,
		Size:        0,
		Compilation: compilationArtist,

		// 关系ID索引(复合艺术家)
		AllArtistIDs: allArtistIDs,

		// 索引排序信息
		SortArtistName:  e.getSortArtistName(artistTag),
		OrderArtistName: e.getOrderArtistName(artistTag),
	}
}

type AudioMetadataExtractorTaglib struct {
	mediaID primitive.ObjectID
}

func generateDeterministicID(seed string) primitive.ObjectID {
	hash := sha256.Sum256([]byte(seed))
	return primitive.ObjectID(hash[:12])
}

func (e *AudioMetadataExtractorTaglib) hasMultipleArtists(artist string) bool {
	separators := []string{"|", "｜", "/", "//", ",", "，", "&", ";", "; ", "、"}
	artist = strings.TrimSpace(artist)
	for _, sep := range separators {
		if strings.Contains(artist, sep) {
			return true
		}
	}
	return false
}

func formatMultipleArtists(artistTag string) (string, []scene_audio_db_models.ArtistIDPair) {
	separators := []string{"|", "｜", "/", "//", ",", "，", "&", ";", "; ", "、"}
	currentList := []string{artistTag}

	for _, sep := range separators {
		var newList []string
		for _, item := range currentList {
			parts := strings.Split(item, sep)
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					newList = append(newList, trimmed)
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

	allArtistPairs := make([]scene_audio_db_models.ArtistIDPair, len(dedupedList))
	for i, artist := range dedupedList {
		artistID := generateDeterministicID(artist).Hex()
		allArtistPairs[i] = scene_audio_db_models.ArtistIDPair{
			ArtistName: artist,
			ArtistID:   artistID,
		}
	}

	return joinedArtists, allArtistPairs
}

func (e *AudioMetadataExtractorTaglib) cleanText(text string) string {
	reg := regexp.MustCompile(`[^\p{L}\p{N}\s]`)
	cleaned := reg.ReplaceAllString(text, "")
	return strings.TrimSpace(cleaned)
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
	nonAlphaPattern := regexp.MustCompile(`[^\p{L}]`)
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

func parseCueFile(cuePath string) (
	globalMeta map[string]string,
	tracks []scene_audio_db_models.CueTrack,
	err error,
) {
	data, err := os.ReadFile(cuePath)
	if err != nil {
		return nil, nil, fmt.Errorf("读取CUE文件失败: %w", err)
	}

	globalMeta = make(map[string]string)
	var currentTrack *scene_audio_db_models.CueTrack
	inTrackBlock := false

	// 自动检测编码并转换为UTF-8
	var finalData []byte
	if isGBK(data) {
		decoder := simplifiedchinese.GBK.NewDecoder()
		utf8Data, _, _ := transform.Bytes(decoder, data)
		finalData = utf8Data
	} else {
		finalData = data
	}

	lines := strings.Split(string(finalData), "\n")
	for _, rawLine := range lines {
		// 防御性检查：跳过空行
		if len(rawLine) == 0 {
			continue
		}

		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		// 1. TRACK行处理
		if strings.HasPrefix(line, "TRACK ") {
			inTrackBlock = true
			if currentTrack != nil {
				tracks = append(tracks, *currentTrack)
			}

			parts := strings.Fields(line)
			if len(parts) < 3 {
				continue
			}
			trackNum, _ := strconv.Atoi(parts[1])
			currentTrack = &scene_audio_db_models.CueTrack{
				TRACK:     trackNum,
				TYPE:      parts[2],
				INDEXES:   []scene_audio_db_models.CueIndex{},
				Title:     "",
				Performer: "",
			}
			continue
		}

		// 2. 全局元数据解析
		if !inTrackBlock {
			switch {
			case strings.HasPrefix(line, "REM "):
				// 支持多种REM类型：GENRE/DATE/DISCID/COMMENT
				remParts := strings.SplitN(line[4:], " ", 2)
				if len(remParts) == 2 {
					key := strings.TrimSpace(remParts[0])
					value := strings.Trim(strings.TrimSpace(remParts[1]), `"`)
					globalMeta[key] = value
				}
			case strings.HasPrefix(line, "PERFORMER "):
				globalMeta["PERFORMER"] = extractQuotedValueSimple(line[10:])
			case strings.HasPrefix(line, "TITLE "):
				globalMeta["TITLE"] = extractQuotedValueSimple(line[6:])
			case strings.HasPrefix(line, "FILE "):
				if value, ok := extractQuotedValue(rawLine, "FILE"); ok {
					globalMeta["FILE"] = value
				}
			case strings.HasPrefix(line, "CATALOG "):
				globalMeta["CATALOG"] = strings.TrimSpace(line[8:])
			case strings.HasPrefix(line, "SONGWRITER "):
				globalMeta["SONGWRITER"] = extractQuotedValueSimple(line[11:])
			}
			continue
		}

		// 3. 音轨元数据解析
		if rawLine[0] == ' ' || rawLine[0] == '\t' {
			trimmedLine := strings.TrimSpace(line)

			// 确保currentTrack非空指针[1,5](@ref)
			if currentTrack == nil {
				currentTrack = &scene_audio_db_models.CueTrack{
					INDEXES: make([]scene_audio_db_models.CueIndex, 0), // 初始化切片[6,7](@ref)
				}
			}

			switch {
			case strings.HasPrefix(trimmedLine, "TITLE "):
				if value, ok := extractQuotedValue(rawLine, "TITLE"); ok {
					currentTrack.Title = value
				}
			case strings.HasPrefix(trimmedLine, "PERFORMER "):
				if value, ok := extractQuotedValue(rawLine, "PERFORMER"); ok {
					currentTrack.Performer = value
					currentTrack.PerformerID = generateDeterministicID(value).Hex()
				}
			case strings.HasPrefix(trimmedLine, "FLAGS "):
				currentTrack.FLAGS = strings.TrimSpace(trimmedLine[6:])
			case strings.HasPrefix(trimmedLine, "INDEX "):
				parts := strings.Fields(trimmedLine)
				if len(parts) >= 3 {
					indexNum, _ := strconv.Atoi(parts[1])
					currentTrack.INDEXES = append(currentTrack.INDEXES,
						scene_audio_db_models.CueIndex{
							INDEX: indexNum,
							TIME:  parts[2],
						})
				}
			case strings.HasPrefix(trimmedLine, "ISRC "):
				currentTrack.ISRC = strings.TrimSpace(trimmedLine[5:])
			case strings.HasPrefix(trimmedLine, "REM REPLAYGAIN_TRACK_GAIN "):
				if gainStr := strings.TrimPrefix(trimmedLine, "REM REPLAYGAIN_TRACK_GAIN "); gainStr != "" {
					// 移除单位并转换[1](@ref)
					cleanGainStr := strings.TrimSuffix(gainStr, " dB")
					if gain, err := strconv.ParseFloat(cleanGainStr, 64); err == nil {
						currentTrack.GAIN = gain
					} else {
						log.Printf("无效GAIN值: %s", gainStr)
					}
				}
			case strings.HasPrefix(trimmedLine, "REM REPLAYGAIN_TRACK_PEAK "):
				if peakStr := strings.TrimPrefix(trimmedLine, "REM REPLAYGAIN_TRACK_PEAK "); peakStr != "" {
					if peak, err := strconv.ParseFloat(peakStr, 64); err == nil {
						currentTrack.PEAK = peak
					} else {
						log.Printf("无效PEAK值: %s", peakStr)
					}
				}
			}
		}
	}

	// 添加最后一个音轨
	if currentTrack != nil {
		tracks = append(tracks, *currentTrack)
	}

	return globalMeta, tracks, nil
}

// 检测是否为GBK编码
func isGBK(data []byte) bool {
	length := len(data)
	var i int
	for i < length {
		if data[i] <= 0x7f {
			i++
			continue
		}

		if i+1 >= length {
			return false
		}

		if data[i] >= 0x81 && data[i] <= 0xfe &&
			data[i+1] >= 0x40 && data[i+1] <= 0xfe && data[i+1] != 0x7f {
			i += 2
			continue
		}

		return false
	}
	return true
}

// 提取带引号的值（兼容中英文引号）
func extractQuotedValue(rawLine, key string) (string, bool) {
	keyIdx := strings.Index(rawLine, key)
	if keyIdx == -1 {
		return "", false
	}

	// 定位起始引号（兼容中英文引号）
	start := strings.IndexAny(rawLine[keyIdx:], `"'“”`)
	if start == -1 {
		return "", false
	}
	start += keyIdx + 1

	// 定位结束引号
	end := strings.IndexAny(rawLine[start:], `"'“”`)
	if end == -1 {
		return "", false
	}
	return rawLine[start : start+end], true
}

func extractQuotedValueSimple(s string) string {
	quotes := `"'“”`
	return strings.TrimSpace(strings.Trim(s, quotes))
}
