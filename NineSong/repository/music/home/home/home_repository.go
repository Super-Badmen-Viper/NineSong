package home

// 导出core中的所有接口和类型
import (
	homeDomainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/home/home/core"
	homeRepoCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/home/home/core"
)

// 重新导出核心类型
type HomeRepository = homeDomainCore.HomeRepository

// 重新导出构造函数
var NewHomeRepository = homeRepoCore.NewHomeRepository
