package usecase_app_config

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_app/repository_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// AppAudioConfigUsecase implements the usecase interface for app audio configuration.
// It embeds the generic ConfigUsecase to handle the core GetAll/ReplaceAll logic.
type AppAudioConfigUsecase struct {
	usecase.ConfigUsecase[domain_app_config.AppAudioConfig]
}

// NewAppAudioConfigUsecase creates a new usecase for app audio configuration.
// It uses the generic NewConfigUsecase constructor for consistency.
func NewAppAudioConfigUsecase(repo repository_app_config.AppAudioConfigRepository, timeout time.Duration) domain_app_config.AppAudioConfigUsecase {
	baseUsecase := usecase.NewConfigUsecase[domain_app_config.AppAudioConfig](repo, timeout)
	return &AppAudioConfigUsecase{
		ConfigUsecase: baseUsecase,
	}
}
