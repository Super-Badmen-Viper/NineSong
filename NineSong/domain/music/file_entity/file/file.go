package file

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/file_entity/file/core"
)

// 重新导出核心类型
type FileTypeNo = core.FileTypeNo
type FileMetadata = core.FileMetadata
type FileRepository = core.FileRepository
type FileDetector = core.FileDetector
type FileDetectorImpl = core.FileDetectorImpl

// 重新导出常量
const (
	Audio      = core.Audio
	Video      = core.Video
	Image      = core.Image
	Text       = core.Text
	Document   = core.Document
	Archive    = core.Archive
	Executable = core.Executable
	Database   = core.Database
	Unknown    = core.Unknown
)

// 重新导出元数据常量
const (
	AcoustIDFingerprint       = core.AcoustIDFingerprint
	AcoustIDID                = core.AcoustIDID
	Album                     = core.Album
	AlbumArtist               = core.AlbumArtist
	AlbumArtistSort           = core.AlbumArtistSort
	AlbumSort                 = core.AlbumSort
	Arranger                  = core.Arranger
	Artist                    = core.Artist
	Artists                   = core.Artists
	ArtistSort                = core.ArtistSort
	ArtistWebpage             = core.ArtistWebpage
	ASIN                      = core.ASIN
	AudioSourceWebpage        = core.AudioSourceWebpage
	Barcode                   = core.Barcode
	BPM                       = core.BPM
	CatalogNumber             = core.CatalogNumber
	Comment                   = core.Comment
	Compilation               = core.Compilation
	Composer                  = core.Composer
	ComposerSort              = core.ComposerSort
	Conductor                 = core.Conductor
	Copyright                 = core.Copyright
	CopyrightURL              = core.CopyrightURL
	Date                      = core.Date
	DiscNumber                = core.DiscNumber
	DiscSubtitle              = core.DiscSubtitle
	DJMixer                   = core.DJMixer
	EncodedBy                 = core.EncodedBy
	Encoding                  = core.Encoding
	EncodingTime              = core.EncodingTime
	Engineer                  = core.Engineer
	FileType                  = core.FileType
	FileWebpage               = core.FileWebpage
	GaplessPlayback           = core.GaplessPlayback
	Genre                     = core.Genre
	Grouping                  = core.Grouping
	InitialKey                = core.InitialKey
	InvolvedPeople            = core.InvolvedPeople
	ISRC                      = core.ISRC
	Label                     = core.Label
	Language                  = core.Language
	Length                    = core.Length
	License                   = core.License
	Lyricist                  = core.Lyricist
	Lyrics                    = core.Lyrics
	Media                     = core.Media
	Mixer                     = core.Mixer
	Mood                      = core.Mood
	MovementCount             = core.MovementCount
	MovementName              = core.MovementName
	MovementNumber            = core.MovementNumber
	MusicBrainzAlbumID        = core.MusicBrainzAlbumID
	MusicBrainzAlbumArtistID  = core.MusicBrainzAlbumArtistID
	MusicBrainzArtistID       = core.MusicBrainzArtistID
	MusicBrainzReleaseGroupID = core.MusicBrainzReleaseGroupID
	MusicBrainzReleaseTrackID = core.MusicBrainzReleaseTrackID
	MusicBrainzTrackID        = core.MusicBrainzTrackID
	MusicBrainzWorkID         = core.MusicBrainzWorkID
	MusicianCredits           = core.MusicianCredits
	MusicIPPUID               = core.MusicIPPUID
	OriginalAlbum             = core.OriginalAlbum
	OriginalArtist            = core.OriginalArtist
	OriginalDate              = core.OriginalDate
	OriginalFilename          = core.OriginalFilename
	OriginalLyricist          = core.OriginalLyricist
	Owner                     = core.Owner
	PaymentWebpage            = core.PaymentWebpage
	Performer                 = core.Performer
	PlaylistDelay             = core.PlaylistDelay
	Podcast                   = core.Podcast
	PodcastCategory           = core.PodcastCategory
	PodcastDesc               = core.PodcastDesc
	PodcastID                 = core.PodcastID
	PodcastURL                = core.PodcastURL
	ProducedNotice            = core.ProducedNotice
	Producer                  = core.Producer
	PublisherWebpage          = core.PublisherWebpage
	RadioStation              = core.RadioStation
	RadioStationOwner         = core.RadioStationOwner
	RadioStationWebpage       = core.RadioStationWebpage
	ReleaseCountry            = core.ReleaseCountry
	ReleaseDate               = core.ReleaseDate
	ReleaseStatus             = core.ReleaseStatus
	ReleaseType               = core.ReleaseType
	Remixer                   = core.Remixer
	Script                    = core.Script
	ShowSort                  = core.ShowSort
	ShowWorkMovement          = core.ShowWorkMovement
	Subtitle                  = core.Subtitle
	TaggingDate               = core.TaggingDate
	Title                     = core.Title
	TitleSort                 = core.TitleSort
	TrackNumber               = core.TrackNumber
	TVEpisode                 = core.TVEpisode
	TVEpisodeID               = core.TVEpisodeID
	TVNetwork                 = core.TVNetwork
	TVSeason                  = core.TVSeason
	TVShow                    = core.TVShow
	URL                       = core.URL
	Work                      = core.Work
)
