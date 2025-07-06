package scene_audio_route_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_route/scene_audio_route_models"
)

type AnnotationRepository interface {
	UpdateStarred(ctx context.Context, itemId string, itemType string) (bool, error)
	UpdateUnStarred(ctx context.Context, itemId string, itemType string) (bool, error)
	UpdateRating(ctx context.Context, itemId string, itemType string, rating int) (bool, error)
	UpdateScrobble(ctx context.Context, itemId string, itemType string) (bool, error)
	UpdateCompleteScrobble(ctx context.Context, itemId string, itemType string) (bool, error)

	UpdateTagSource(ctx context.Context, itemId string, itemType string, tags []scene_audio_route_models.TagSource) (bool, error)
	UpdateWeightedTag(ctx context.Context, itemId string, itemType string, tags []scene_audio_route_models.WeightedTag) (bool, error)
}
