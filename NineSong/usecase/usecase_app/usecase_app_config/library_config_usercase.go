package usecase_app_config

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_app/repository_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// AppLibraryConfigUsecase implements the usecase interface for app library configuration.
// It embeds the generic ConfigUsecase to handle the core GetAll/ReplaceAll logic.
type AppLibraryConfigUsecase struct {
	usecase.ConfigUsecase[domain_app_config.AppLibraryConfig]
}

// NewAppLibraryConfigUsecase creates a new usecase for app library configuration.
// It uses the generic NewConfigUsecase constructor for consistency.
func NewAppLibraryConfigUsecase(repo repository_app_config.AppLibraryConfigRepository, timeout time.Duration) domain_app_config.AppLibraryConfigUsecase {
	baseUsecase := usecase.NewConfigUsecase[domain_app_config.AppLibraryConfig](repo, timeout)
	return &AppLibraryConfigUsecase{
		ConfigUsecase: baseUsecase,
	}
}
