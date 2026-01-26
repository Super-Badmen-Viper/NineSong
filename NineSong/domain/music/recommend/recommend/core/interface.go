package core

import (
	"context"
)

// RecommendRepository 推荐仓库接口
type RecommendRepository interface {
	GetGeneralRecommendations(
		ctx context.Context,
		recommendType string,
		limit int,
		randomSeed string,
		recommendOffset string,
		logShow bool,
		refresh bool,
	) ([]interface{}, error)

	GetPersonalizedRecommendations(
		ctx context.Context,
		userId string,
		recommendType string,
		limit int,
		logShow bool,
		refresh bool,
	) ([]interface{}, error)

	GetPopularRecommendations(
		ctx context.Context,
		recommendType string,
		limit int,
		logShow bool,
		refresh bool,
	) ([]interface{}, error)
}
