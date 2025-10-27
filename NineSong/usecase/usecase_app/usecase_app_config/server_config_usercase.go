package usecase_app_config

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_app/repository_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// AppServerConfigUsecase implements the usecase interface for app server configuration.
// It embeds the generic BaseUsecase to handle the core CRUD logic.
type AppServerConfigUsecase struct {
	usecase.BaseUsecase[domain_app_config.AppServerConfig]
}

// NewAppServerConfigUsecase creates a new usecase for app server configuration.
// It uses the generic NewBaseUsecase constructor for consistency.
func NewAppServerConfigUsecase(repo repository_app_config.AppServerConfigRepository, timeout time.Duration) domain_app_config.AppServerConfigUsecase {
	baseUsecase := usecase.NewBaseUsecase[domain_app_config.AppServerConfig](repo, timeout)
	return &AppServerConfigUsecase{
		BaseUsecase: baseUsecase,
	}
}
