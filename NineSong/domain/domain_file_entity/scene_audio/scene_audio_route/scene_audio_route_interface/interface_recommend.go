package scene_audio_route_interface

import (
	"context"
)

type RecommendRouteRepository interface {
	GetGeneralRecommendations(
		ctx context.Context,
		recommendType string,
		limit int,
		randomSeed string,
		recommendOffset string,
	) ([]interface{}, error)

	GetPersonalizedRecommendations(
		ctx context.Context,
		userId string,
		recommendType string,
		limit int,
	) ([]interface{}, error)

	GetPopularRecommendations(
		ctx context.Context,
		recommendType string,
		limit int,
	) ([]interface{}, error)
}
