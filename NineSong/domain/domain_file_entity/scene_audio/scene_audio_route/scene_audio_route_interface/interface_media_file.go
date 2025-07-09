package scene_audio_route_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
)

type MediaFileRepository interface {
	GetMediaFileItems(
		ctx context.Context,
		start, end, sort, order,
		search, starred,
		albumId, artistId,
		year string,
	) ([]scene_audio_route_models.MediaFileMetadata, error)

	GetMediaFileItemsMultipleSorting(
		ctx context.Context,
		start, end string,
		sortOrder []domain.SortOrder,
		search, starred,
		albumId, artistId,
		year string,
	) ([]scene_audio_route_models.MediaFileMetadata, error)

	GetMediaFileFilterItemsCount(
		ctx context.Context,
		search, starred, albumId, artistId, year string,
	) (*scene_audio_route_models.MediaFileFilterCounts, error)
}
