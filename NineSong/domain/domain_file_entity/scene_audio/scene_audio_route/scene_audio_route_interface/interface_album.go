package scene_audio_route_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
)

type AlbumRepository interface {
	GetAlbumItems(
		ctx context.Context,
		start, end, sort, order,
		search, starred,
		artistId,
		minYear, maxYear string,
	) ([]scene_audio_route_models.AlbumMetadata, error)

	GetAlbumItemsMultipleSorting(
		ctx context.Context,
		start, end string,
		sortOrder []domain_util.SortOrder,
		search, starred,
		artistId,
		minYear, maxYear string,
	) ([]scene_audio_route_models.AlbumMetadata, error)

	GetAlbumFilterItemsCount(
		ctx context.Context,
	) (*scene_audio_route_models.AlbumFilterCounts, error)
}
