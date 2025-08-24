package repository_app_config

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// AppConfigRepository is an alias for the generic ConfigRepository.
// It handles collections of configuration items.
type AppConfigRepository interface {
	domain.ConfigRepository[domain_app_config.AppConfig]
}

// NewAppConfigRepository creates a new repository for app configurations.
// It uses the generic ConfigMongoRepository implementation which supports GetAll and ReplaceAll.
func NewAppConfigRepository(db mongo.Database, collection string) AppConfigRepository {
	return repository.NewConfigMongoRepository[domain_app_config.AppConfig](db, collection)
}
