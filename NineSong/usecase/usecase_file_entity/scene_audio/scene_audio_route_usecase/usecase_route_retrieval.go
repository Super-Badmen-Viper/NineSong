package scene_audio_route_usecase

import (
	"context"
	"errors"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type retrievalUsecase struct {
	repo    scene_audio_route_interface.RetrievalRepository
	timeout time.Duration
}

func NewRetrievalUsecase(repo scene_audio_route_interface.RetrievalRepository, timeout time.Duration) scene_audio_route_interface.RetrievalRepository {
	return &retrievalUsecase{
		repo:    repo,
		timeout: timeout,
	}
}

func (uc *retrievalUsecase) GetStreamPath(ctx context.Context, mediaFileId string, cueModel bool) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if _, err := primitive.ObjectIDFromHex(mediaFileId); err != nil {
		return "", errors.New("invalid media file id format")
	}
	return uc.repo.GetStreamPath(ctx, mediaFileId, cueModel)
}

func (uc *retrievalUsecase) GetStreamTempPath(ctx context.Context, metadataType string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	return uc.repo.GetStreamTempPath(ctx, metadataType)
}

func (uc *retrievalUsecase) GetDownloadPath(ctx context.Context, mediaFileId string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	if _, err := primitive.ObjectIDFromHex(mediaFileId); err != nil {
		return "", errors.New("invalid media file id format")
	}
	return uc.repo.GetDownloadPath(ctx, mediaFileId)
}

func (uc *retrievalUsecase) GetCoverArtID(ctx context.Context, fileType string, targetID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 扩展参数校验
	allowedTypes := map[string]bool{
		"media": true, "album": true, "artist": true,
		"back": true, "cover": true, "disc": true,
	}
	if !allowedTypes[fileType] {
		return "", errors.New("invalid file type parameter")
	}

	return uc.repo.GetCoverArtID(ctx, fileType, targetID)
}

func (uc *retrievalUsecase) GetLyricsLrcMetaData(ctx context.Context, mediaFileId string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 参数格式验证
	if _, err := primitive.ObjectIDFromHex(mediaFileId); err != nil {
		return "", errors.New("invalid media file id format")
	}

	// 添加业务规则验证（示例）
	if len(mediaFileId) != 24 {
		return "", errors.New("media file id must be 24 hex characters")
	}

	return uc.repo.GetLyricsLrcMetaData(ctx, mediaFileId)
}

func (uc *retrievalUsecase) GetLyricsLrcFile(ctx context.Context, mediaFileId string) (string, error) {
	//TODO implement me
	panic("implement me")
}
