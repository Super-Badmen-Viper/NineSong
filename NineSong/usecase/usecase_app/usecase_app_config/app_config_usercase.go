package usecase_app_config

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_app/repository_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// AppConfigUsecase implements the usecase interface for app configuration.
// It embeds the generic ConfigUsecase to handle the core GetAll/ReplaceAll logic.
type AppConfigUsecase struct {
	usecase.ConfigUsecase[domain_app_config.AppConfig]
}

// NewAppConfigUsecase creates a new usecase for app configuration.
// It uses the generic NewConfigUsecase constructor for consistency.
func NewAppConfigUsecase(repo repository_app_config.AppConfigRepository, timeout time.Duration) domain_app_config.AppConfigUsecase {
	baseUsecase := usecase.NewConfigUsecase[domain_app_config.AppConfig](repo, timeout)
	return &AppConfigUsecase{
		ConfigUsecase: baseUsecase,
	}
}
