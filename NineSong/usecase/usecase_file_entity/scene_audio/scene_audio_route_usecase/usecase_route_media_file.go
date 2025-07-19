package scene_audio_route_usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"strconv"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mediaFileUsecase struct {
	mediaFileRepo scene_audio_route_interface.MediaFileRepository
	timeout       time.Duration
}

func NewMediaFileUsecase(repo scene_audio_route_interface.MediaFileRepository, timeout time.Duration) scene_audio_route_interface.MediaFileRepository {
	return &mediaFileUsecase{
		mediaFileRepo: repo,
		timeout:       timeout,
	}
}

func (uc *mediaFileUsecase) GetMediaFileItems(
	ctx context.Context,
	start, end, sort, order, search, starred, albumId, artistId, year,
	suffix, minBitrate, maxBitrate, folderPath string,
) ([]scene_audio_route_models.MediaFileMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 参数验证
	validations := []func() error{
		func() error {
			if _, err := strconv.Atoi(start); start != "" && err != nil {
				return errors.New("invalid start parameter")
			}
			return nil
		},
		func() error {
			if _, err := strconv.Atoi(end); end != "" && err != nil {
				return errors.New("invalid end parameter")
			}
			return nil
		},
		func() error {
			if albumId != "" {
				if _, err := primitive.ObjectIDFromHex(albumId); err != nil {
					return errors.New("invalid album id format")
				}
			}
			return nil
		},
		func() error {
			if artistId != "" {
				if _, err := primitive.ObjectIDFromHex(artistId); err != nil {
					return errors.New("invalid artist id format")
				}
			}
			return nil
		},
		func() error {
			if year != "" {
				if _, err := strconv.Atoi(year); err != nil {
					return errors.New("year must be integer")
				}
			}
			return nil
		},
	}

	for _, validate := range validations {
		if err := validate(); err != nil {
			return nil, err
		}
	}

	return uc.mediaFileRepo.GetMediaFileItems(
		ctx, start, end, sort, order,
		search, starred,
		albumId, artistId, year,
		suffix, minBitrate, maxBitrate, folderPath,
	)
}

func (uc *mediaFileUsecase) GetMediaFileItemsIds(
	ctx context.Context, ids []string,
) ([]scene_audio_route_models.MediaFileMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	return uc.mediaFileRepo.GetMediaFileItemsIds(ctx, ids)
}

func (uc *mediaFileUsecase) GetMediaFileItemsMultipleSorting(
	ctx context.Context,
	start, end string,
	sortOrder []domain_util.SortOrder,
	search, starred, albumId, artistId, year,
	suffix, minBitrate, maxBitrate, folderPath string,
) ([]scene_audio_route_models.MediaFileMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// 参数验证
	validations := []func() error{
		func() error {
			if _, err := strconv.Atoi(start); start != "" && err != nil {
				return errors.New("invalid start parameter")
			}
			return nil
		},
		func() error {
			if _, err := strconv.Atoi(end); end != "" && err != nil {
				return errors.New("invalid end parameter")
			}
			return nil
		},
		func() error {
			if albumId != "" {
				if _, err := primitive.ObjectIDFromHex(albumId); err != nil {
					return errors.New("invalid album id format")
				}
			}
			return nil
		},
		func() error {
			if artistId != "" {
				if _, err := primitive.ObjectIDFromHex(artistId); err != nil {
					return errors.New("invalid artist id format")
				}
			}
			return nil
		},
		func() error {
			if year != "" {
				if _, err := strconv.Atoi(year); err != nil {
					return errors.New("year must be integer")
				}
			}
			return nil
		},
		func() error {
			// 验证排序字段有效性
			validFields := map[string]bool{
				"title": true, "album": true, "artist": true, "album_artist": true,
				"year": true, "rating": true, "starred_at": true, "genre": true,
				"play_count": true, "play_date": true, "duration": true,
				"bit_rate": true, "size": true, "created_at": true, "updated_at": true,
			}
			for _, so := range sortOrder {
				if !validFields[strings.ToLower(so.Sort)] {
					return fmt.Errorf("invalid sort field: %s", so.Sort)
				}
				if so.Order != "asc" && so.Order != "desc" {
					return fmt.Errorf("invalid sort order: %s", so.Order)
				}
			}
			return nil
		},
	}

	for _, validate := range validations {
		if err := validate(); err != nil {
			return nil, err
		}
	}

	return uc.mediaFileRepo.GetMediaFileItemsMultipleSorting(
		ctx, start, end,
		sortOrder, search, starred,
		albumId, artistId, year,
		suffix, minBitrate, maxBitrate, folderPath,
	)
}

func (uc *mediaFileUsecase) GetMediaFileFilterItemsCount(
	ctx context.Context,
) (*scene_audio_route_models.MediaFileFilterCounts, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	return uc.mediaFileRepo.GetMediaFileFilterItemsCount(ctx)
}
