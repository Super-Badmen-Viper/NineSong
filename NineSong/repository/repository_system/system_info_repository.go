package repository_system

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_system"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// SystemInfoRepository is an alias for the generic ConfigRepository.
// No custom methods are needed for this configuration entity.
type SystemInfoRepository interface {
	domain.ConfigRepository[domain_system.SystemInfo]
}

// NewSystemInfoRepository creates a new repository for system information.
// It uses the generic ConfigMongoRepository implementation.
func NewSystemInfoRepository(db mongo.Database, collection string) SystemInfoRepository {
	return repository.NewConfigMongoRepository[domain_system.SystemInfo](db, collection)
}
