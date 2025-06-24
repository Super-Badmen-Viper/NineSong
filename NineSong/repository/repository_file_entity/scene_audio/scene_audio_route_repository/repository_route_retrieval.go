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

func (r *retrievalRepository) GetStreamPath(ctx context.Context, mediaFileId string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(mediaFileId)
	if err != nil {
		return "", errors.New("invalid media file id format")
	}

	collection := r.db.Collection(domain.CollectionFileEntityAudioSceneMediaFile)
	var result scene_audio_route_models.MediaFileMetadata
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("stream metadata not found: %w", err)
	}
	return result.Path, nil
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

func (r *retrievalRepository) GetCoverArt(ctx context.Context, fileType string, targetID string) (string, error) {
	if _, err := primitive.ObjectIDFromHex(targetID); err != nil {
		return "", errors.New("invalid target id format")
	}

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
