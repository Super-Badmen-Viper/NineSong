package scene_audio_route_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
)

type WordCloudMetadata interface {
	GetAllWordCloudSearch(
		ctx context.Context,
	) ([]scene_audio_db_models.WordCloudMetadata, error)

	GetAllGenreSearch(
		ctx context.Context,
	) ([]scene_audio_db_models.WordCloudMetadata, error)

	GetHighFrequencyWordCloudSearch(
		ctx context.Context, wordLimit int,
	) ([]scene_audio_db_models.WordCloudMetadata, error)

	GetRecommendedWordCloudSearch(
		ctx context.Context, keywords []string,
	) ([]scene_audio_db_models.Recommendation, error)
}
