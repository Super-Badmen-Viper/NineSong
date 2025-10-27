package scene_audio_route_interface

import (
	"context"
)

type RecommendRouteRepository interface {
	GetRecommendAnnotationWordCloudItems(
		ctx context.Context,
		start, end, recommendType, randomSeed, offset string,
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
