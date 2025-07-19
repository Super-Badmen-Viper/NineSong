package scene_audio_route_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
)

type ArtistRepository interface {
	GetArtistItems(
		ctx context.Context,
		start, end, sort, order,
		search, starred string,
	) ([]scene_audio_route_models.ArtistMetadata, error)

	GetArtistItemsMultipleSorting(
		ctx context.Context,
		start, end string,
		sortOrder []domain_util.SortOrder,
		search, starred string,
	) ([]scene_audio_route_models.ArtistMetadata, error)

	GetArtistFilterItemsCount(
		ctx context.Context,
	) (*scene_audio_route_models.ArtistFilterCounts, error)
}
