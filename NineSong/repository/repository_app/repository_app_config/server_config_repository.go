package repository_app_config

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// AppServerConfigRepository is an alias for the generic BaseRepository.
type AppServerConfigRepository interface {
	domain.BaseRepository[domain_app_config.AppServerConfig]
}

// NewAppServerConfigRepository creates a new repository for app server configurations.
// It uses the generic BaseMongoRepository implementation.
func NewAppServerConfigRepository(db mongo.Database, collection string) AppServerConfigRepository {
	return repository.NewBaseMongoRepository[domain_app_config.AppServerConfig](db, collection)
}
