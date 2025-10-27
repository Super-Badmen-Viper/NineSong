package domain_app_config

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppPlaylistIDConfig struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ConfigKey   string             `bson:"config_key"`
	ConfigValue string             `bson:"config_value"`
}

// AppPlaylistIDConfigUsecase defines the usecase interface for app playlist ID configuration.
// It embeds the generic ConfigUsecase to provide standard GetAll/ReplaceAll operations.
type AppPlaylistIDConfigUsecase interface {
	usecase.ConfigUsecase[AppPlaylistIDConfig]
}
