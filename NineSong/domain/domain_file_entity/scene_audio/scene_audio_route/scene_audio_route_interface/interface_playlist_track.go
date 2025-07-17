package scene_audio_route_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
)

type PlaylistTrackRepository interface {
	GetPlaylistTrackItems(
		ctx context.Context,
		start, end, sort, order,
		search, starred,
		albumId, artistId,
		year,
		playlistId string,
	) ([]scene_audio_route_models.MediaFileMetadata, error)

	GetPlaylistTrackItemsMultipleSorting(
		ctx context.Context,
		start, end string,
		sortOrder []domain_util.SortOrder,
		search, starred,
		albumId, artistId,
		year,
		playlistId string,
	) ([]scene_audio_route_models.MediaFileMetadata, error)

	GetPlaylistTrackFilterItemsCount(
		ctx context.Context,
		search, albumId, artistId, year string,
	) (*scene_audio_route_models.MediaFileFilterCounts, error)

	AddPlaylistTrackItems(
		ctx context.Context,
		playlistId string,
		mediaFileIds string,
	) (bool, error)

	RemovePlaylistTrackItems(
		ctx context.Context,
		playlistId string,
		mediaFileIds string,
	) (bool, error)

	SortPlaylistTrackItems(
		ctx context.Context,
		playlistId string,
		mediaFileIds string,
	) (bool, error)
}
