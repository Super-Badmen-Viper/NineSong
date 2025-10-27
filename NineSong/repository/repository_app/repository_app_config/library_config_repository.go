package repository_app_config

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// AppLibraryConfigRepository is an alias for the generic ConfigRepository.
// It handles collections of library configuration items.
type AppLibraryConfigRepository interface {
	domain.ConfigRepository[domain_app_config.AppLibraryConfig]
}

// NewAppLibraryConfigRepository creates a new repository for app library configurations.
// It uses the generic ConfigMongoRepository implementation.
func NewAppLibraryConfigRepository(db mongo.Database, collection string) AppLibraryConfigRepository {
	return repository.NewConfigMongoRepository[domain_app_config.AppLibraryConfig](db, collection)
}
