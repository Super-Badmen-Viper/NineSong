package usecase_system

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_system"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_system"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// systemConfigurationUsecase implements the usecase interface for system configuration.
// It embeds the generic ConfigUsecase to handle the core Get/Update logic.
// This approach reduces boilerplate code and centralizes the common logic in base_usecase.go.
type systemConfigurationUsecase struct {
	usecase.ConfigUsecase[domain_system.SystemConfiguration]
}

// NewSystemConfigurationUsecase creates a new usecase for system configuration.
// It leverages the generic NewConfigUsecase constructor to wire up the repository
// and timeout, promoting code reuse and consistency.
func NewSystemConfigurationUsecase(repo repository_system.SystemConfigurationRepository, timeout time.Duration) domain_system.SystemConfigurationUsecase {
	// We use the generic constructor from the usecase package to create the base usecase.
	// This ensures that all standard CRUD logic is handled consistently.
	baseUsecase := usecase.NewConfigUsecase[domain_system.SystemConfiguration](repo, timeout)

	return &systemConfigurationUsecase{
		ConfigUsecase: baseUsecase,
	}
}
