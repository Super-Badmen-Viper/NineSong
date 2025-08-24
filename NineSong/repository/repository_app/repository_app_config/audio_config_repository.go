package repository_app_config

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// AppAudioConfigRepository is an alias for the generic ConfigRepository.
// It handles collections of audio configuration items.
type AppAudioConfigRepository interface {
	domain.ConfigRepository[domain_app_config.AppAudioConfig]
}

// NewAppAudioConfigRepository creates a new repository for app audio configurations.
// It uses the generic ConfigMongoRepository implementation.
func NewAppAudioConfigRepository(db mongo.Database, collection string) AppAudioConfigRepository {
	return repository.NewConfigMongoRepository[domain_app_config.AppAudioConfig](db, collection)
}
