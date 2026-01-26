package scene_audio_route_repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/siongui/gojianfan"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
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
		"media": true, "album": true, "artist": true, "playlist": true,
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

		var typePath string
		if fileType == "playlist" {
			// 播放列表封面路径：playlist/playlist_{playlistID}.jpg
			coverFileName := fmt.Sprintf("playlist_%s.jpg", targetID)
			typePath, err = r.checkCoverFile(filepath.Join(tempMeta.FolderPath, "playlist"), coverFileName)
		} else {
			// 其他类型：{type}/{targetID}/cover.jpg
			typePath, err = r.checkCoverFile(filepath.Join(tempMeta.FolderPath, fileType, targetID), "cover.jpg")
			if err != nil && errors.Is(err, os.ErrNotExist) {
				typePath, err = r.checkCoverFile(filepath.Join(tempMeta.FolderPath, fileType, targetID), "cover.png")
			}
		}

		if err != nil {
			return "", fmt.Errorf("cover art not found: %w", err)
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
	// 处理fileType="embedded"的情况：直接从media_file表读取
	if fileType == "embedded" {
		if len(mediaFileId) == 0 {
			return "", errors.New("mediaFileId is required for embedded lyrics")
		}
		objID, err := primitive.ObjectIDFromHex(mediaFileId)
		if err != nil {
			return "", errors.New("invalid media file id format")
		}

		collection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
		var result scene_audio_db_models.MediaFileMetadata

		filter := bson.M{"_id": objID}
		err = collection.FindOne(ctx, filter).Decode(&result)
		if err != nil {
			return "", fmt.Errorf("database query failed: %w", err)
		}

		if result.Lyrics == "" {
			return "", errors.New("no embedded lyrics found")
		}

		return result.Lyrics, nil
	}

	// 如果fileType为空，默认为auto模式
	if fileType == "" {
		fileType = "auto"
	}

	// 处理auto模式的优先级查询
	if fileType == "auto" {
		priorityTypes := []string{"krc", "qrc", "lrc"}

		// 如果提供了mediaFileId，优先根据id获取文件路径，然后查找同名文件
		if len(mediaFileId) > 0 {
			// 1. 根据mediaFileId获取文件路径，从原始文件夹查找同名文件（krc > qrc > lrc）
			mediaPath, err := r.getMediaFilePath(ctx, mediaFileId)
			if err == nil {
				content, err := r.findLyricsInDirectory(mediaPath, priorityTypes)
				if err == nil {
					return content, nil
				}
			}

			// 2. 如果原始文件夹找不到，尝试从缓存查询（基于artist和title的模糊匹配）
			// 先获取媒体文件的artist和title
			objID, err := primitive.ObjectIDFromHex(mediaFileId)
			if err == nil {
				mediaCollection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
				var mediaResult scene_audio_db_models.MediaFileMetadata
				filter := bson.M{"_id": objID}
				if err := mediaCollection.FindOne(ctx, filter).Decode(&mediaResult); err == nil {
					// 使用媒体文件的artist和title从缓存查询（使用模糊匹配）
					cacheArtist := mediaResult.Artist
					cacheTitle := mediaResult.Title

					// 使用模糊匹配逻辑查询缓存
					collection := r.db.Collection(domain.CollectionFileEntityAudioSceneLyricsFile)
					specialChars := []string{"|", "｜", "/", "//", ",", "，", "&", ";", "; ", "、"}

					removeSpecialChars := func(s string) string {
						for _, char := range specialChars {
							s = strings.ReplaceAll(s, char, "")
						}
						return s
					}

					createFTSRegexPattern := func(input string) string {
						cleaned := removeSpecialChars(input)
						if cleaned == "" {
							return ""
						}
						t2s := gojianfan.T2S(cleaned)
						s2t := gojianfan.S2T(cleaned)
						doubleConvert := gojianfan.S2T(t2s)
						return regexp.QuoteMeta(cleaned) + "|" +
							regexp.QuoteMeta(t2s) + "|" +
							regexp.QuoteMeta(s2t) + "|" +
							regexp.QuoteMeta(doubleConvert)
					}

					baseFilter := bson.M{}
					if cacheArtist == cacheTitle && cacheTitle != "" {
						pattern := createFTSRegexPattern(cacheTitle)
						if pattern != "" {
							baseFilter["title"] = primitive.Regex{
								Pattern: pattern,
								Options: "i",
							}
						}
					} else {
						if cacheArtist != "" {
							hasSeparator := false
							for _, char := range specialChars {
								if strings.Contains(cacheArtist, char) {
									hasSeparator = true
									break
								}
							}

							if hasSeparator {
								separatorPattern := regexp.QuoteMeta(strings.Join(specialChars, "|"))
								re := regexp.MustCompile("[" + separatorPattern + "]+")
								artists := re.Split(cacheArtist, -1)

								orPatterns := make([]string, 0)
								for _, a := range artists {
									if trimmed := strings.TrimSpace(a); trimmed != "" {
										if pattern := createFTSRegexPattern(trimmed); pattern != "" {
											orPatterns = append(orPatterns, pattern)
										}
									}
								}

								if len(orPatterns) > 0 {
									baseFilter["artist"] = primitive.Regex{
										Pattern: "(" + strings.Join(orPatterns, "|") + ")",
										Options: "i",
									}
								}
							} else {
								pattern := createFTSRegexPattern(cacheArtist)
								if pattern != "" {
									baseFilter["artist"] = primitive.Regex{
										Pattern: pattern,
										Options: "i",
									}
								}
							}
						}

						if cacheTitle != "" {
							pattern := createFTSRegexPattern(cacheTitle)
							if pattern != "" {
								baseFilter["title"] = primitive.Regex{
									Pattern: pattern,
									Options: "i",
								}
							}
						}
					}

					// 从缓存查询
					for _, priorityType := range priorityTypes {
						priorityFilter := bson.M{}
						for k, v := range baseFilter {
							priorityFilter[k] = v
						}
						priorityFilter["fileType"] = primitive.Regex{
							Pattern: "^" + regexp.QuoteMeta(priorityType) + "$",
							Options: "i",
						}

						var cacheResult scene_audio_db_models.LyricsFileMetadata
						if err := collection.FindOne(ctx, priorityFilter).Decode(&cacheResult); err == nil {
							// 检查缓存文件是否存在
							content, err := os.ReadFile(cacheResult.Path)
							if err == nil {
								return string(content), nil
							}
						}
					}

					// 3. 如果缓存也找不到，从media_file表读取Lyrics字段
					if mediaResult.Lyrics != "" {
						return mediaResult.Lyrics, nil
					}
				}
			}
		} else {
			// 如果没有提供mediaFileId，使用artist和title从缓存查询（原有逻辑）
			collection := r.db.Collection(domain.CollectionFileEntityAudioSceneLyricsFile)
			specialChars := []string{"|", "｜", "/", "//", ",", "，", "&", ";", "; ", "、"}

			removeSpecialChars := func(s string) string {
				for _, char := range specialChars {
					s = strings.ReplaceAll(s, char, "")
				}
				return s
			}

			createFTSRegexPattern := func(input string) string {
				cleaned := removeSpecialChars(input)
				if cleaned == "" {
					return ""
				}
				t2s := gojianfan.T2S(cleaned)
				s2t := gojianfan.S2T(cleaned)
				doubleConvert := gojianfan.S2T(t2s)
				return regexp.QuoteMeta(cleaned) + "|" +
					regexp.QuoteMeta(t2s) + "|" +
					regexp.QuoteMeta(s2t) + "|" +
					regexp.QuoteMeta(doubleConvert)
			}

			baseFilter := bson.M{}
			if artist == title && title != "" {
				pattern := createFTSRegexPattern(title)
				if pattern != "" {
					baseFilter["title"] = primitive.Regex{
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
							baseFilter["artist"] = primitive.Regex{
								Pattern: "(" + strings.Join(orPatterns, "|") + ")",
								Options: "i",
							}
						}
					} else {
						pattern := createFTSRegexPattern(artist)
						if pattern != "" {
							baseFilter["artist"] = primitive.Regex{
								Pattern: pattern,
								Options: "i",
							}
						}
					}
				}

				if title != "" {
					pattern := createFTSRegexPattern(title)
					if pattern != "" {
						baseFilter["title"] = primitive.Regex{
							Pattern: pattern,
							Options: "i",
						}
					}
				}
			}

			// 尝试从缓存查询
			for _, priorityType := range priorityTypes {
				priorityFilter := bson.M{}
				for k, v := range baseFilter {
					priorityFilter[k] = v
				}
				priorityFilter["fileType"] = primitive.Regex{
					Pattern: "^" + regexp.QuoteMeta(priorityType) + "$",
					Options: "i",
				}

				var priorityResult scene_audio_db_models.LyricsFileMetadata
				if err := collection.FindOne(ctx, priorityFilter).Decode(&priorityResult); err != nil {
					if !errors.Is(err, driver.ErrNoDocuments) {
						// 非"未找到文档"的错误，记录但继续
					}
					continue
				}

				// 检查缓存文件是否存在
				content, err := os.ReadFile(priorityResult.Path)
				if err == nil {
					return string(content), nil
				}
			}
		}

		// 所有优先级都未找到
		return "", errors.New("no matching lyrics found")
	}

	// 处理mediaFileId直接查询的情况（原有逻辑）
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

		// 处理非auto类型的查询逻辑
		if fileType != "" {
			cleanedFileType := removeSpecialChars(fileType)
			if cleanedFileType != "" && cleanedFileType != "auto" {
				filter["fileType"] = primitive.Regex{
					Pattern: "^" + regexp.QuoteMeta(cleanedFileType) + "$",
					Options: "i",
				}
			}
		}

		// 从缓存查询
		if err := collection.FindOne(ctx, filter).Decode(&result); err != nil {
			// 如果缓存找不到且提供了mediaFileId，尝试从原始文件夹查找
			if errors.Is(err, driver.ErrNoDocuments) && len(mediaFileId) > 0 {
				mediaPath, pathErr := r.getMediaFilePath(ctx, mediaFileId)
				if pathErr == nil {
					// 根据fileType确定要查找的格式
					var searchFormats []string
					if fileType != "" {
						cleanedFileType := removeSpecialChars(fileType)
						if cleanedFileType != "" && cleanedFileType != "auto" {
							searchFormats = []string{cleanedFileType}
						} else {
							searchFormats = []string{"lrc"} // 默认查找lrc
						}
					} else {
						searchFormats = []string{"lrc"} // 默认查找lrc
					}

					content, findErr := r.findLyricsInDirectory(mediaPath, searchFormats)
					if findErr == nil {
						return content, nil
					}
				}
			}

			if errors.Is(err, driver.ErrNoDocuments) {
				return "", errors.New("no matching lyrics found")
			}
			return "", fmt.Errorf("lyrics query failed: %w", err)
		}

		// 检查缓存文件是否存在
		content, err := os.ReadFile(result.Path)
		if err != nil {
			// 缓存文件不存在，如果提供了mediaFileId，尝试从原始文件夹查找
			if len(mediaFileId) > 0 {
				mediaPath, pathErr := r.getMediaFilePath(ctx, mediaFileId)
				if pathErr == nil {
					var searchFormats []string
					if fileType != "" {
						cleanedFileType := removeSpecialChars(fileType)
						if cleanedFileType != "" && cleanedFileType != "auto" {
							searchFormats = []string{cleanedFileType}
						} else {
							searchFormats = []string{"lrc"}
						}
					} else {
						searchFormats = []string{"lrc"}
					}

					content, findErr := r.findLyricsInDirectory(mediaPath, searchFormats)
					if findErr == nil {
						return content, nil
					}
				}
			}
			return "", fmt.Errorf("failed to read lyrics file at %s: %w", result.Path, err)
		}
		return string(content), nil
	}
}

// findLyricsInDirectory 在指定目录查找与音频文件同名的歌词文件
func (r *retrievalRepository) findLyricsInDirectory(audioPath string, priorityTypes []string) (string, error) {
	audioDir := filepath.Dir(audioPath)
	audioName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))

	for _, format := range priorityTypes {
		lyricsPath := filepath.Join(audioDir, audioName+"."+format)
		if _, err := os.Stat(lyricsPath); err == nil {
			content, err := os.ReadFile(lyricsPath)
			if err != nil {
				continue // 读取失败，继续下一个
			}
			return string(content), nil
		}
	}
	return "", errors.New("no lyrics file found in directory")
}

// getMediaFilePath 根据mediaFileId获取媒体文件路径
func (r *retrievalRepository) getMediaFilePath(ctx context.Context, mediaFileId string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(mediaFileId)
	if err != nil {
		return "", errors.New("invalid media file id format")
	}

	collection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
	var result scene_audio_route_models.MediaFileMetadata

	filter := bson.M{"_id": objID}
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("media file not found: %w", err)
	}

	return result.Path, nil
}

func (r *retrievalRepository) GetLyricsLrcFile(ctx context.Context, mediaFileId string) (string, error) {
	//TODO implement me
	panic("implement me")
}
