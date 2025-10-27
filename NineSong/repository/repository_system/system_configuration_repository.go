package repository_system

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_system"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// SystemConfigurationRepository is an alias for the generic ConfigRepository.
// No custom methods are needed for this configuration entity.
type SystemConfigurationRepository interface {
	domain.ConfigRepository[domain_system.SystemConfiguration]
}

// NewSystemConfigurationRepository creates a new repository for system configurations.
// It uses the generic ConfigMongoRepository implementation.
func NewSystemConfigurationRepository(db mongo.Database, collection string) SystemConfigurationRepository {
	return repository.NewConfigMongoRepository[domain_system.SystemConfiguration](db, collection)
}
