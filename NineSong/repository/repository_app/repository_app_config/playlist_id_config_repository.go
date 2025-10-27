package repository_app_config

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// AppPlaylistIDConfigRepository is an alias for the generic ConfigRepository.
// It handles collections of playlist ID configuration items.
type AppPlaylistIDConfigRepository interface {
	domain.ConfigRepository[domain_app_config.AppPlaylistIDConfig]
}

// NewAppPlaylistIDConfigRepository creates a new repository for app playlist ID configurations.
// It uses the generic ConfigMongoRepository implementation.
func NewAppPlaylistIDConfigRepository(db mongo.Database, collection string) AppPlaylistIDConfigRepository {
	return repository.NewConfigMongoRepository[domain_app_config.AppPlaylistIDConfig](db, collection)
}
