package scene_audio_route_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
)

type MediaFileRepository interface {
	GetMediaFileItems(
		ctx context.Context,
		start, end, sort, order,
		search, starred,
		albumId, artistId,
		year,
		suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter string,
	) ([]scene_audio_route_models.MediaFileMetadata, error)

	GetMediaFileItemsMultipleSorting(
		ctx context.Context,
		start, end string,
		sortOrder []domain_util.SortOrder,
		search, starred,
		albumId, artistId,
		year,
		suffix, minBitrate, maxBitrate, folderPath, folderPathSubFilter string,
	) ([]scene_audio_route_models.MediaFileMetadata, error)

	GetMediaFileFilterItemsCount(
		ctx context.Context,
	) (*scene_audio_route_models.MediaFileFilterCounts, error)

	GetMediaFileItemsIds(
		ctx context.Context, ids []string,
	) ([]scene_audio_route_models.MediaFileMetadata, error)
}
