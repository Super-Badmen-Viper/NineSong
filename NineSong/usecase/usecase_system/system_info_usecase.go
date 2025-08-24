package usecase_system

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_system"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_system"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
)

// systemInfoUsecase implements the usecase interface for system information.
// It embeds the generic ConfigUsecase to handle the core Get/Update logic.
type systemInfoUsecase struct {
	usecase.ConfigUsecase[domain_system.SystemInfo]
}

// NewSystemInfoUsecase creates a new usecase for system information.
// It uses the generic NewConfigUsecase constructor for consistency and code reuse.
func NewSystemInfoUsecase(repo repository_system.SystemInfoRepository, timeout time.Duration) domain_system.SystemInfoUsecase {
	baseUsecase := usecase.NewConfigUsecase[domain_system.SystemInfo](repo, timeout)
	return &systemInfoUsecase{
		ConfigUsecase: baseUsecase,
	}
}
