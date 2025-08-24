package usecase_app_config

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_app/domain_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_app/repository_app_config"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// AppPlaylistIDConfigUsecase implements the usecase interface for app playlist ID configuration.
// It embeds the generic ConfigUsecase to handle the core GetAll/ReplaceAll logic.
type AppPlaylistIDConfigUsecase struct {
	usecase.ConfigUsecase[domain_app_config.AppPlaylistIDConfig]
}

// NewAppPlaylistIDConfigUsecase creates a new usecase for app playlist ID configuration.
// It uses the generic NewConfigUsecase constructor for consistency.
func NewAppPlaylistIDConfigUsecase(repo repository_app_config.AppPlaylistIDConfigRepository, timeout time.Duration) domain_app_config.AppPlaylistIDConfigUsecase {
	baseUsecase := usecase.NewConfigUsecase[domain_app_config.AppPlaylistIDConfig](repo, timeout)
	return &AppPlaylistIDConfigUsecase{
		ConfigUsecase: baseUsecase,
	}
}
