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
)

type ArtistUsecase struct {
	repo    scene_audio_route_interface.ArtistRepository
	timeout time.Duration
}

func NewArtistUsecase(repo scene_audio_route_interface.ArtistRepository, timeout time.Duration) *ArtistUsecase {
	return &ArtistUsecase{
		repo:    repo,
		timeout: timeout,
	}
}

func (uc *ArtistUsecase) GetArtistItems(
	ctx context.Context,
	start, end, sort, order, search, starred string,
) ([]scene_audio_route_models.ArtistMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

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
			if starred != "" {
				if _, err := strconv.ParseBool(starred); err != nil {
					return errors.New("invalid starred parameter")
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

	return uc.repo.GetArtistItems(ctx, start, end, sort, order, search, starred)
}

func (uc *ArtistUsecase) GetArtistItemsMultipleSorting(
	ctx context.Context,
	start, end string,
	sortOrder []domain_util.SortOrder,
	search, starred string,
) ([]scene_audio_route_models.ArtistMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	validations := []func() error{
		// 分页参数验证
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
		// Starred参数验证
		func() error {
			if starred != "" {
				if _, err := strconv.ParseBool(starred); err != nil {
					return errors.New("invalid starred parameter")
				}
			}
			return nil
		},
		// 排序参数验证
		func() error {
			validSortFields := map[string]bool{
				"name":        true,
				"album_count": true,
				"song_count":  true,
				"play_count":  true,
				"play_date":   true,
				"rating":      true,
				"starred_at":  true,
				"size":        true,
				"created_at":  true,
				"updated_at":  true,
			}

			for _, so := range sortOrder {
				if !validSortFields[strings.ToLower(so.Sort)] {
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

	return uc.repo.GetArtistItemsMultipleSorting(
		ctx, start, end, sortOrder, search, starred,
	)
}

func (uc *ArtistUsecase) GetArtistFilterItemsCount(
	ctx context.Context,
	search, starred string,
) (*scene_audio_route_models.ArtistFilterCounts, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	// Starred参数验证
	if starred != "" {
		if _, err := strconv.ParseBool(starred); err != nil {
			return nil, errors.New("invalid starred parameter")
		}
	}

	return uc.repo.GetArtistFilterItemsCount(ctx, search, starred)
}
