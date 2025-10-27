package repository_app_library

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_library"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
)

// AppMediaFileLibraryRepository is an alias for the generic ConfigRepository.
// It handles the collection of media file library documents.
type AppMediaFileLibraryRepository interface {
	domain.ConfigRepository[domain_app_library.AppMediaFileLibrary]
}

// NewAppMediaFileLibraryRepository creates a new repository for the media file library.
// It uses the generic ConfigMongoRepository implementation.
func NewAppMediaFileLibraryRepository(db mongo.Database, collection string) AppMediaFileLibraryRepository {
	return repository.NewConfigMongoRepository[domain_app_library.AppMediaFileLibrary](db, collection)
}
