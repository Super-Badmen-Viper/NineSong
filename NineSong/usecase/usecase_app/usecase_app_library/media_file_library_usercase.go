package usecase_app_library

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_library"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_app/repository_app_library"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// AppMediaFileLibraryUsecase implements the usecase interface for the media file library.
// It embeds the generic ConfigUsecase to handle the core GetAll/ReplaceAll logic.
type AppMediaFileLibraryUsecase struct {
	usecase.ConfigUsecase[domain_app_library.AppMediaFileLibrary]
}

// NewAppMediaFileLibraryUsecase creates a new usecase for the media file library.
// It uses the generic NewConfigUsecase constructor for consistency.
func NewAppMediaFileLibraryUsecase(repo repository_app_library.AppMediaFileLibraryRepository, timeout time.Duration) domain_app_library.AppMediaFileLibraryUsecase {
	baseUsecase := usecase.NewConfigUsecase[domain_app_library.AppMediaFileLibrary](repo, timeout)
	return &AppMediaFileLibraryUsecase{
		ConfigUsecase: baseUsecase,
	}
}
