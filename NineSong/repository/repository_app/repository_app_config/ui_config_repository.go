package repository_app_config

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// AppUIConfigRepository is an alias for the generic ConfigRepository.
// It handles collections of UI configuration items.
type AppUIConfigRepository interface {
	domain.ConfigRepository[domain_app_config.AppUIConfig]
}

// NewAppUIConfigRepository creates a new repository for app UI configurations.
// It uses the generic ConfigMongoRepository implementation.
func NewAppUIConfigRepository(db mongo.Database, collection string) AppUIConfigRepository {
	return repository.NewConfigMongoRepository[domain_app_config.AppUIConfig](db, collection)
}
