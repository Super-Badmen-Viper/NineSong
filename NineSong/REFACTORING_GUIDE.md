# 项目重构迁移指南

## 已完成的工作

### 1. 目录结构创建 ✓
- 已创建所有新的目录结构（架构/场景/功能/元数据类型/core|utils）
- API层、Domain层、Repository层、Usecase层的目录结构已就绪

### 2. 共享代码迁移 ✓
- `domain/shared/` - 基础Repository接口、collections、error/success响应
- `domain/shared/util/` - 工具函数（min_heap、query_options、sort_order、stop_words、task_progress、time_stamp）
- `repository/shared/` - 基础MongoDB Repository实现
- `usecase/shared/` - 基础Usecase实现
- `infrastructure/` - mongo和internal目录已迁移

## 待完成的工作

### 3. 音乐场景代码迁移

#### 3.1 Media Library功能
需要迁移的文件：
- Domain: `domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models/model_media_library_audio.go` 
  → `domain/music/media_library/media_library_audio/core/model.go`
- Domain Interface: `domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface/interface_media_library_audio.go`
  → `domain/music/media_library/media_library_audio/core/interface.go`
- Repository: `repository/repository_file_entity/scene_audio/scene_audio_db_repository/repository_db_media_library_audio.go`
  → `repository/music/media_library/media_library_audio/core/repository.go`
- Usecase: `usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase/usecase_route_media_library_audio.go`
  → `usecase/music/media_library/media_library_audio/core/usecase.go`
- Controller: `api/controller/controller_file_entity/scene_audio_route_api_controller/scene_audio_media_library_audio_controller.go`
  → `api/controller/music/media_library/media_library_audio/media_library_audio_controller.go`
- Route: `api/route/route_file_entity/scene_audio_route_api_route/scene_audio_media_library_audio_route.go`
  → `api/route/music/media_library/media_library_audio/media_library_audio_route.go`

#### 3.2 其他音乐场景功能
按照相同的模式迁移：
- file_entity (file, folder)
- artist
- album
- playlist (playlist, playlist_track)
- media_file (media_file, media_file_cue)
- retrieval
- recommend
- word_cloud
- annotation
- home
- sync_record

### 4. 其他场景代码迁移

#### 4.1 App场景
- config (app_config, audio_config, library_config, playlist_id_config, server_config, ui_config)
- library (media_file_library)

#### 4.2 Auth场景
- user, login, signup, refresh_token, profile, update_user, task

#### 4.3 System场景
- system_info, system_configuration

## 迁移步骤

### 步骤1: 迁移Domain层
1. 将模型文件迁移到 `domain/music/{功能}/{元数据类型}/core/model.go`
2. 将接口文件迁移到 `domain/music/{功能}/{元数据类型}/core/interface.go`
3. 将工具函数迁移到 `domain/music/{功能}/{元数据类型}/utils/`
4. 创建统一导出文件 `domain/music/{功能}/{元数据类型}/{元数据类型}.go`

### 步骤2: 迁移Repository层
1. 将Repository实现迁移到 `repository/music/{功能}/{元数据类型}/core/repository.go`
2. 更新import路径，引用新的domain路径

### 步骤3: 迁移Usecase层
1. 将Usecase实现迁移到 `usecase/music/{功能}/{元数据类型}/core/usecase.go`
2. 将工具函数迁移到 `usecase/music/{功能}/{元数据类型}/utils/`
3. 更新import路径

### 步骤4: 迁移API层
1. 将Controller迁移到 `api/controller/music/{功能}/{元数据类型}/{元数据类型}_controller.go`
2. 将Route迁移到 `api/route/music/{功能}/{元数据类型}/{元数据类型}_route.go`
3. 更新import路径

### 步骤5: 更新所有import路径
使用以下命令批量更新import路径：

```bash
# 更新domain导入
find . -name "*.go" -type f -exec sed -i '' 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared|g' {} \;

# 更新mongo导入
find . -name "*.go" -type f -exec sed -i '' 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo|github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo|g' {} \;

# 更新internal导入
find . -name "*.go" -type f -exec sed -i '' 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/internal|github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/internal|g' {} \;
```

## 注意事项

1. **保持向后兼容**: 在迁移过程中，可能需要保留旧的导入路径一段时间
2. **测试覆盖**: 确保所有测试用例在迁移后仍然通过
3. **Git历史**: 考虑使用 `git mv` 保留文件历史
4. **依赖关系**: 注意模块间的依赖关系，按依赖顺序迁移
5. **Package名称**: 确保每个文件的package名称与目录结构匹配

## 统一导出文件示例

每个元数据类型目录下应有一个统一导出文件，例如：

```go
// domain/music/media_library/media_library_audio/media_library_audio.go
package media_library_audio

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/media_library_audio/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/media_library_audio/utils"
)

// 重新导出核心类型
type MediaLibraryAudio = core.MediaLibraryAudio
type MediaLibraryAudioRepository = core.MediaLibraryAudioRepository

// 重新导出工具函数
var ProcessMetadata = utils.ProcessMetadata
```

这样外部只需要 `import "domain/music/media_library/media_library_audio"` 即可使用所有功能。

## 验证步骤

1. 运行 `go mod tidy` 更新依赖
2. 运行 `go build ./...` 检查编译错误
3. 运行所有测试 `go test ./...`
4. 检查linter错误 `golangci-lint run`
