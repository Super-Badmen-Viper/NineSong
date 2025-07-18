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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"path/filepath"
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

func (r *retrievalRepository) GetLyricsLrcMetaData(ctx context.Context, mediaFileId string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(mediaFileId)
	if err != nil {
		return "", errors.New("invalid media file id format")
	}

	collection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
	var result scene_audio_route_models.RetrievalLyricsMetadata

	filter := bson.M{"_id": objID}

	// 执行查询
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("database query failed: %w", err)
	}

	// 返回歌词内容
	return result.Lyrics, nil
}

func (r *retrievalRepository) GetLyricsLrcFile(ctx context.Context, mediaFileId string) (string, error) {
	//TODO implement me
	panic("implement me")
}
