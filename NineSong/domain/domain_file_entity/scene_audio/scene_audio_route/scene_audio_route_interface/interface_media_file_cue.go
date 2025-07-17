package scene_audio_route_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
)

type MediaFileCueRepository interface {
	GetMediaFileCueItems(
		ctx context.Context,
		start, end, sort, order,
		search, starred,
		albumId, artistId,
		year string,
	) ([]scene_audio_route_models.MediaFileCueMetadata, error)

	GetMediaFileCueItemsMultipleSorting(
		ctx context.Context,
		start, end string,
		sortOrder []domain_util.SortOrder,
		search, starred,
		albumId, artistId,
		year string,
	) ([]scene_audio_route_models.MediaFileCueMetadata, error)

	GetMediaFileCueFilterItemsCount(
		ctx context.Context,
		search, starred, albumId, artistId, year string,
	) (*scene_audio_route_models.MediaFileCueFilterCounts, error)
}
