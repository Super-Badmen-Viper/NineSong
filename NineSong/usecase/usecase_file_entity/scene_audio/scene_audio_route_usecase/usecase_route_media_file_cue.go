package scene_audio_route_usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"strconv"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mediaFileCueUsecase struct {
	mediaFileCueRepo scene_audio_route_interface.MediaFileCueRepository
	timeout          time.Duration
}

func NewMediaFileCueUsecase(repo scene_audio_route_interface.MediaFileCueRepository, timeout time.Duration) scene_audio_route_interface.MediaFileCueRepository {
	return &mediaFileCueUsecase{
		mediaFileCueRepo: repo,
		timeout:          timeout,
	}
}

func (uc *mediaFileCueUsecase) GetMediaFileCueItems(
	ctx context.Context,
	start, end, sort, order, search, starred, albumId, artistId, year string,
) ([]scene_audio_route_models.MediaFileCueMetadata, error) {
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

	return uc.mediaFileCueRepo.GetMediaFileCueItems(ctx, start, end, sort, order, search, starred, albumId, artistId, year)
}

func (uc *mediaFileCueUsecase) GetMediaFileCueItemsMultipleSorting(
	ctx context.Context,
	start, end string,
	sortOrder []domain.SortOrder,
	search, starred, albumId, artistId, year string,
) ([]scene_audio_route_models.MediaFileCueMetadata, error) {
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
				"title": true, "performer": true, "year": true, "rating": true,
				"starred_at": true, "genre": true, "play_count": true, "play_date": true,
				"size": true, "created_at": true, "updated_at": true, "cue_track_count": true,
				"bit_rate": true, "duration": true, "sample_rate": true, "track_title": true,
				"track_performer": true,
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

	return uc.mediaFileCueRepo.GetMediaFileCueItemsMultipleSorting(
		ctx, start, end, sortOrder, search, starred, albumId, artistId, year,
	)
}

func (uc *mediaFileCueUsecase) GetMediaFileCueFilterItemsCount(
	ctx context.Context,
	search, starred, albumId, artistId, year string,
) (*scene_audio_route_models.MediaFileCueFilterCounts, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	return uc.mediaFileCueRepo.GetMediaFileCueFilterItemsCount(ctx, search, starred, albumId, artistId, year)
}
