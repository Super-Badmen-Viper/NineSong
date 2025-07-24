package domain_file_entity

import (
	"archive/zip"
	"bytes"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileTypeNo int

const (
	Audio FileTypeNo = iota + 1
	Video
	Image
	Text
	Document
	Archive
	Executable
	Database
	Unknown
)

const (
	AcoustIDFingerprint       = "ACOUSTID_FINGERPRINT"
	AcoustIDID                = "ACOUSTID_ID"
	Album                     = "ALBUM"
	AlbumArtist               = "ALBUMARTIST"
	AlbumArtistSort           = "ALBUMARTISTSORT"
	AlbumSort                 = "ALBUMSORT"
	Arranger                  = "ARRANGER"
	Artist                    = "ARTIST"
	Artists                   = "ARTISTS"
	ArtistSort                = "ARTISTSORT"
	ArtistWebpage             = "ARTISTWEBPAGE"
	ASIN                      = "ASIN"
	AudioSourceWebpage        = "AUDIOSOURCEWEBPAGE"
	Barcode                   = "BARCODE"
	BPM                       = "BPM"
	CatalogNumber             = "CATALOGNUMBER"
	Comment                   = "COMMENT"
	Compilation               = "COMPILATION"
	Composer                  = "COMPOSER"
	ComposerSort              = "COMPOSERSORT"
	Conductor                 = "CONDUCTOR"
	Copyright                 = "COPYRIGHT"
	CopyrightURL              = "COPYRIGHTURL"
	Date                      = "DATE"
	DiscNumber                = "DISCNUMBER"
	DiscSubtitle              = "DISCSUBTITLE"
	DJMixer                   = "DJMIXER"
	EncodedBy                 = "ENCODEDBY"
	Encoding                  = "ENCODING"
	EncodingTime              = "ENCODINGTIME"
	Engineer                  = "ENGINEER"
	FileType                  = "FILETYPE"
	FileWebpage               = "FILEWEBPAGE"
	GaplessPlayback           = "GAPLESSPLAYBACK"
	Genre                     = "GENRE"
	Grouping                  = "GROUPING"
	InitialKey                = "INITIALKEY"
	InvolvedPeople            = "INVOLVEDPEOPLE"
	ISRC                      = "ISRC"
	Label                     = "LABEL"
	Language                  = "LANGUAGE"
	Length                    = "LENGTH"
	License                   = "LICENSE"
	Lyricist                  = "LYRICIST"
	Lyrics                    = "LYRICS"
	Media                     = "MEDIA"
	Mixer                     = "MIXER"
	Mood                      = "MOOD"
	MovementCount             = "MOVEMENTCOUNT"
	MovementName              = "MOVEMENTNAME"
	MovementNumber            = "MOVEMENTNUMBER"
	MusicBrainzAlbumID        = "MUSICBRAINZ_ALBUMID"
	MusicBrainzAlbumArtistID  = "MUSICBRAINZ_ALBUMARTISTID"
	MusicBrainzArtistID       = "MUSICBRAINZ_ARTISTID"
	MusicBrainzReleaseGroupID = "MUSICBRAINZ_RELEASEGROUPID"
	MusicBrainzReleaseTrackID = "MUSICBRAINZ_RELEASETRACKID"
	MusicBrainzTrackID        = "MUSICBRAINZ_TRACKID"
	MusicBrainzWorkID         = "MUSICBRAINZ_WORKID"
	MusicianCredits           = "MUSICIANCREDITS"
	MusicIPPUID               = "MUSICIP_PUID"
	OriginalAlbum             = "ORIGINALALBUM"
	OriginalArtist            = "ORIGINALARTIST"
	OriginalDate              = "ORIGINALDATE"
	OriginalFilename          = "ORIGINALFILENAME"
	OriginalLyricist          = "ORIGINALLYRICIST"
	Owner                     = "OWNER"
	PaymentWebpage            = "PAYMENTWEBPAGE"
	Performer                 = "PERFORMER"
	PlaylistDelay             = "PLAYLISTDELAY"
	Podcast                   = "PODCAST"
	PodcastCategory           = "PODCASTCATEGORY"
	PodcastDesc               = "PODCASTDESC"
	PodcastID                 = "PODCASTID"
	PodcastURL                = "PODCASTURL"
	ProducedNotice            = "PRODUCEDNOTICE"
	Producer                  = "PRODUCER"
	PublisherWebpage          = "PUBLISHERWEBPAGE"
	RadioStation              = "RADIOSTATION"
	RadioStationOwner         = "RADIOSTATIONOWNER"
	RadioStationWebpage       = "RADIOSTATIONWEBPAGE"
	ReleaseCountry            = "RELEASECOUNTRY"
	ReleaseDate               = "RELEASEDATE"
	ReleaseStatus             = "RELEASESTATUS"
	ReleaseType               = "RELEASETYPE"
	Remixer                   = "REMIXER"
	Script                    = "SCRIPT"
	ShowSort                  = "SHOWSORT"
	ShowWorkMovement          = "SHOWWORKMOVEMENT"
	Subtitle                  = "SUBTITLE"
	TaggingDate               = "TAGGINGDATE"
	Title                     = "TITLE"
	TitleSort                 = "TITLESORT"
	TrackNumber               = "TRACKNUMBER"
	TVEpisode                 = "TVEPISODE"
	TVEpisodeID               = "TVEPISODEID"
	TVNetwork                 = "TVNETWORK"
	TVSeason                  = "TVSEASON"
	TVShow                    = "TVSHOW"
	URL                       = "URL"
	Work                      = "WORK"
)

type FileMetadata struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	FolderID         primitive.ObjectID `bson:"folder_id"`
	FileName         string             `bson:"file_name" validate:"required,filename"`
	FileNameNoSuffix string             `bson:"file_name_no_suffix" validate:"required,filename_no_suffix"`
	LibraryPath      string             `bson:"library_path" validate:"required,filepath"`
	FilePath         string             `bson:"file_path" validate:"filepath"`
	FileType         FileTypeNo         `bson:"file_type" validate:"min=1,max=8"`
	Size             int64              `bson:"size" validate:"min=0"`
	ModTime          time.Time          `bson:"mod_time" validate:"required"`
	Checksum         string             `bson:"checksum" validate:"sha256"`
	CreatedAt        time.Time          `bson:"created_at" validate:"required"`
	UpdatedAt        time.Time          `bson:"updated_at" validate:"required,gtfield=CreatedAt"`
}

type FileRepository interface {
	Upsert(ctx context.Context, file *FileMetadata) error
	FindByPath(ctx context.Context, path string) (*FileMetadata, error)
	DeleteByFolder(ctx context.Context, folderID primitive.ObjectID) error
	CountByFolderID(ctx context.Context, folderID primitive.ObjectID) (int64, error)
}

type FileDetector interface {
	DetectMediaType(filePath string) (FileTypeNo, error)
}

type FileDetectorImpl struct{}

func (fd *FileDetectorImpl) DetectMediaType(filePath string) (FileTypeNo, error) {
	// 扩展名检测[1](@ref)
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	// 音频类型（按普及度从高到低排序）
	case ".mp3", ".wav", ".aac", ".flac", // 第一梯队：绝对主流
		".m4a", ".wma", ".ogg", ".opus", // 第二梯队：常见有条件支持
		".aiff", ".alac", ".ape", ".cue", // 第二梯队：专业/生态限定
		".dsd", ".dff", ".dsdiff", ".dsf", // 第三梯队：DSD专业格式
		".amr", ".spx", // 第三梯队：语音编码
		".wv", ".tta", ".tak": // 第三梯队：冷门无损
		//".ra", ".vqf", ".shn", ".ofr", ".la":
		return Audio, nil

	// ISO特殊处理
	case ".iso":
		//if fd.isLosslessMusicISO(filePath) {
		//	return Audio, nil
		//}
		return Archive, nil

	// 视频类型
	case ".mp4", ".avi", ".mkv", ".mov", ".flv", ".webm", ".wmv", ".ts":
		return Video, nil

	// 图片类型
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff", ".svg":
		return Image, nil

	// 文本类型
	case ".txt", ".md", ".log", ".ini", ".cfg", ".conf", ".csv", ".xml", ".json":
		return Text, nil

	// 文档类型
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".rtf":
		return Document, nil

	// 压缩类型
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz":
		return Archive, nil

	// 可执行文件
	case ".exe", ".msi", ".bat", ".sh", ".apk", ".dmg":
		return Executable, nil

	default:
		if fd.isDatabaseFile(ext) {
			return Database, nil
		}
	}
	return Unknown, nil
}

// 无损音乐ISO检测[5](@ref)
func (fd *FileDetectorImpl) isLosslessMusicISO(filePath string) bool {
	// 方法1：检查ISO内部特征目录
	if fd.containsAudioFolders(filePath) {
		return true
	}

	// 方法2：尝试解压检测音频文件（示例伪代码）
	return fd.scanExtractedAudio(filePath)
}

// 检查ISO内部特征目录[5](@ref)
func (fd *FileDetectorImpl) containsAudioFolders(filePath string) bool {
	// 读取ISO前1MB内容（足够检测目录结构）
	file, _ := os.Open(filePath)
	defer file.Close()

	buf := make([]byte, 1024*1024)
	file.Read(buf)

	// 检测关键目录签名
	return bytes.Contains(buf, []byte("AUDIO_TS")) || // DVD-Audio
		bytes.Contains(buf, []byte("DSD")) || // SACD镜像
		bytes.Contains(buf, []byte("BDMV")) // 蓝光音频
}

// 解压扫描音频文件（实际实现需完整解压逻辑）
func (fd *FileDetectorImpl) scanExtractedAudio(filePath string) bool {
	// 重命名为.zip临时文件（ISO兼容ZIP结构）[5](@ref)
	tmpPath := filePath + ".zip"
	os.Rename(filePath, tmpPath)
	defer os.Rename(tmpPath, filePath)

	// 打开ZIP文件
	r, _ := zip.OpenReader(tmpPath) // 此代码存在错误异常
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			return
		}
	}(r)

	// 扫描压缩包内文件
	for _, f := range r.File {
		ext := strings.ToLower(filepath.Ext(f.Name))
		switch ext {
		case ".flac", ".wav", ".dsd", ".dff", ".dsf", ".aiff":
			return true // 发现无损音频文件
		}
	}
	return false
}

// 自定义文本检测
func (fd *FileDetectorImpl) isTextContent(buf []byte) bool {
	for _, b := range buf {
		if b < 9 || (b > 13 && b < 32) && b != 27 {
			return false // 发现非文本控制字符
		}
	}
	return true
}

// 可执行文件检测
func (fd *FileDetectorImpl) isExecutable(buf []byte) bool {
	// PE文件签名（Windows）
	if len(buf) > 40 && bytes.Equal(buf[:2], []byte{0x4D, 0x5A}) {
		return true
	}
	// ELF文件签名（Linux）
	if len(buf) > 4 && bytes.Equal(buf[:4], []byte{0x7F, 0x45, 0x4C, 0x46}) {
		return true
	}
	// Mach-O签名（macOS）
	if len(buf) > 4 && bytes.Equal(buf[:4], []byte{0xFE, 0xED, 0xFA, 0xCF}) {
		return true
	}
	return false
}

// 数据库文件检测
func (fd *FileDetectorImpl) isDatabaseFile(ext string) bool {
	switch ext {
	case ".db", ".sqlite", ".mdb", ".accdb", ".dbf":
		return true
	}
	return false
}
