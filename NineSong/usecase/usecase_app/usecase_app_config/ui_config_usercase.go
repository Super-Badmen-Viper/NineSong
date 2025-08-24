package usecase_app_config

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_app/repository_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// AppUIConfigUsecase implements the usecase interface for app UI configuration.
// It embeds the generic ConfigUsecase to handle the core GetAll/ReplaceAll logic.
type AppUIConfigUsecase struct {
	usecase.ConfigUsecase[domain_app_config.AppUIConfig]
}

// NewAppUIConfigUsecase creates a new usecase for app UI configuration.
// It uses the generic NewConfigUsecase constructor for consistency.
func NewAppUIConfigUsecase(repo repository_app_config.AppUIConfigRepository, timeout time.Duration) domain_app_config.AppUIConfigUsecase {
	baseUsecase := usecase.NewConfigUsecase[domain_app_config.AppUIConfig](repo, timeout)
	return &AppUIConfigUsecase{
		ConfigUsecase: baseUsecase,
	}
}
