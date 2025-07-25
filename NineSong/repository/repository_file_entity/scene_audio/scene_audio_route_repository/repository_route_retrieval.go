package scene_audio_route_repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/siongui/gojianfan"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type retrievalRepository struct {
	db mongo.Database
}

func NewRetrievalRepository(db mongo.Database) scene_audio_route_interface.RetrievalRepository {
	return &retrievalRepository{db: db}
}

func (r *retrievalRepository) GetStreamPath(ctx context.Context, mediaFileId string, cueModel bool) (string, error) {
	objID, err := primitive.ObjectIDFromHex(mediaFileId)
	if err != nil {
		return "", errors.New("invalid media file id format")
	}
	if cueModel {
		collection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFileCue)
		var result scene_audio_route_models.MediaFileCueMetadata
		err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
		if err != nil {
			return "", fmt.Errorf("stream metadata not found: %w", err)
		}
		return result.Path, nil
	} else {
		collection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
		var result scene_audio_route_models.MediaFileMetadata
		err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
		if err != nil {
			return "", fmt.Errorf("stream metadata not found: %w", err)
		}
		return result.Path, nil
	}
}

func (r *retrievalRepository) GetStreamTempPath(ctx context.Context, metadataType string) (string, error) {
	collection := r.db.Collection(domain.CollectionFileEntityAudioSceneTempMetadata)
	filter := bson.M{"metadata_type": metadataType}

	var result scene_audio_db_models.ExternalResource
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("路径查询失败: %w | 类型: %s", err, metadataType)
	}

	if result.FolderPath == "" {
		return "", errors.New("数据库返回空路径")
	}

	if _, err := os.Stat(result.FolderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(result.FolderPath, os.ModePerm); err != nil {
			return "", fmt.Errorf("创建目录失败: %w | 路径: %s", err, result.FolderPath)
		}
	} else if err != nil {
		return "", fmt.Errorf("目录检测失败: %w | 路径: %s", err, result.FolderPath)
	}

	return result.FolderPath, nil
}

func (r *retrievalRepository) GetDownloadPath(ctx context.Context, mediaFileId string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(mediaFileId)
	if err != nil {
		return "", errors.New("invalid media file id format")
	}

	collection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
	var result scene_audio_route_models.MediaFileMetadata
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("download metadata not found: %w", err)
	}
	return result.Path, nil
}

func (r *retrievalRepository) GetCoverArtID(ctx context.Context, fileType string, targetID string) (string, error) {
	if _, err := primitive.ObjectIDFromHex(targetID); err != nil {
		return "", errors.New("invalid target id format")
	}

	// 扩展参数校验
	allowedTypes := map[string]bool{
		"media": true, "album": true, "artist": true,
	}
	if !allowedTypes[fileType] {
		// 处理back/cover/disc类型的图片路径查询
		tempCollection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFileCue)

		cueID, _ := primitive.ObjectIDFromHex(targetID)
		filter := bson.M{
			"_id": cueID,
		}

		var result scene_audio_db_models.MediaFileCueMetadata

		err := tempCollection.FindOne(
			ctx,
			filter,
		).Decode(&result)

		if err != nil {
			return "", fmt.Errorf("database query error: %w", err)
		}

		switch fileType {
		case "back":
			return result.BackImageURL, nil
		case "cover":
			return result.CoverImageURL, nil
		case "disc":
			return result.DiscImageURL, nil
		default:
			return "", fmt.Errorf("unsupported file type: %s", fileType)
		}
	} else {
		tempCollection := r.db.Collection(domain.CollectionFileEntityAudioSceneTempMetadata)
		var tempMeta scene_audio_db_models.ExternalResource
		err := tempCollection.FindOne(
			ctx,
			bson.M{"metadata_type": "cover"},
		).Decode(&tempMeta)
		if err != nil {
			return "", fmt.Errorf("cover storage config not found: %w", err)
		}

		typePath, err := r.checkCoverFile(filepath.Join(tempMeta.FolderPath, fileType, targetID), "cover.jpg")
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				typePath, err = r.checkCoverFile(filepath.Join(tempMeta.FolderPath, fileType, targetID), "cover.png")
				if err != nil {
					return "", fmt.Errorf("cover art not found: %w", err)
				}
				return typePath, nil
			}
			return "", fmt.Errorf("file system error: %w", err)
		}

		return typePath, nil
	}
}

func (r *retrievalRepository) checkCoverFile(basePath string, fileName string) (string, error) {
	typePath := filepath.Join(basePath, fileName)
	fileInfo, err := os.Stat(typePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return typePath, err
		}
		return typePath, fmt.Errorf("file system error: %w", err)
	}

	if fileInfo.Size() == 0 {
		return "", errors.New("empty cover art file")
	}

	return typePath, nil
}

func (r *retrievalRepository) GetLyricsLrcMetaData(ctx context.Context, mediaFileId, artist, title, fileType string) (string, error) {
	if len(mediaFileId) > 0 {
		objID, err := primitive.ObjectIDFromHex(mediaFileId)
		if err != nil {
			return "", errors.New("invalid media file id format")
		}

		collection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
		var result scene_audio_route_models.RetrievalLyricsMetadata

		filter := bson.M{"_id": objID}

		err = collection.FindOne(ctx, filter).Decode(&result)
		if err != nil {
			return "", fmt.Errorf("database query failed: %w", err)
		}

		return result.Lyrics, nil
	} else {
		collection := r.db.Collection(domain.CollectionFileEntityAudioSceneLyricsFile)
		var result scene_audio_db_models.LyricsFileMetadata

		specialChars := []string{"|", "｜", "/", "//", ",", "，", "&", ";", "; ", "、"}

		removeSpecialChars := func(s string) string {
			for _, char := range specialChars {
				s = strings.ReplaceAll(s, char, "")
			}
			return s
		}

		// 创建支持简繁体同时匹配的正则模式
		createFTSRegexPattern := func(input string) string {
			cleaned := removeSpecialChars(input)
			if cleaned == "" {
				return ""
			}

			// 生成四种编码形式（覆盖所有简繁体组合）
			t2s := gojianfan.T2S(cleaned)       // 繁体转简体
			s2t := gojianfan.S2T(cleaned)       // 简体转繁体
			doubleConvert := gojianfan.S2T(t2s) // 双重转换

			// 构建组合正则表达式
			return regexp.QuoteMeta(cleaned) + "|" +
				regexp.QuoteMeta(t2s) + "|" +
				regexp.QuoteMeta(s2t) + "|" +
				regexp.QuoteMeta(doubleConvert)
		}

		filter := bson.M{}

		if artist == title && title != "" {
			pattern := createFTSRegexPattern(title)
			if pattern != "" {
				filter["title"] = primitive.Regex{
					Pattern: pattern,
					Options: "i",
				}
			}
		} else {
			if artist != "" {
				hasSeparator := false
				for _, char := range specialChars {
					if strings.Contains(artist, char) {
						hasSeparator = true
						break
					}
				}

				if hasSeparator {
					separatorPattern := regexp.QuoteMeta(strings.Join(specialChars, "|"))
					re := regexp.MustCompile("[" + separatorPattern + "]+")
					artists := re.Split(artist, -1)

					orPatterns := make([]string, 0)
					for _, a := range artists {
						if trimmed := strings.TrimSpace(a); trimmed != "" {
							if pattern := createFTSRegexPattern(trimmed); pattern != "" {
								orPatterns = append(orPatterns, pattern)
							}
						}
					}

					if len(orPatterns) > 0 {
						filter["artist"] = primitive.Regex{
							Pattern: "(" + strings.Join(orPatterns, "|") + ")",
							Options: "i",
						}
					}
				} else {
					pattern := createFTSRegexPattern(artist)
					if pattern != "" {
						filter["artist"] = primitive.Regex{
							Pattern: pattern,
							Options: "i",
						}
					}
				}
			}

			if title != "" {
				pattern := createFTSRegexPattern(title)
				if pattern != "" {
					filter["title"] = primitive.Regex{
						Pattern: pattern,
						Options: "i",
					}
				}
			}
		}

		if fileType != "" {
			cleanedFileType := removeSpecialChars(fileType)
			if cleanedFileType != "" {
				filter["fileType"] = primitive.Regex{
					Pattern: "^" + regexp.QuoteMeta(cleanedFileType) + "$",
					Options: "i",
				}
			}
		}

		if err := collection.FindOne(ctx, filter).Decode(&result); err != nil {
			if errors.Is(err, driver.ErrNoDocuments) {
				return "", errors.New("no matching lyrics found")
			}
			return "", fmt.Errorf("lyrics query failed: %w", err)
		}

		content, err := os.ReadFile(result.Path)
		if err != nil {
			return "", fmt.Errorf("failed to read lyrics file at %s: %w", result.Path, err)
		}
		return string(content), nil
	}
}

func (r *retrievalRepository) GetLyricsLrcFile(ctx context.Context, mediaFileId string) (string, error) {
	//TODO implement me
	panic("implement me")
}
