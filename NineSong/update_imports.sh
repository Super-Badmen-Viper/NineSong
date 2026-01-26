#!/bin/bash

# 批量更新import路径脚本

echo "开始更新import路径..."

# 更新domain导入路径
find . -name "*.go" -type f ! -path "./domain/shared/*" ! -path "./repository/shared/*" ! -path "./usecase/shared/*" ! -path "./infrastructure/*" -exec sed -i '' \
  -e 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/media_library_audio/core|g' \
  -e 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/media_library_audio/core|g' \
  -e 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared|g' \
  {} \;

# 更新repository导入路径
find . -name "*.go" -type f ! -path "./repository/shared/*" ! -path "./infrastructure/*" -exec sed -i '' \
  -e 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/repository_file_entity/scene_audio/scene_audio_db_repository|github.com/amitshekhariitbhu/go-backend-clean-architecture/repository/music/media_library/media_library_audio/core|g' \
  {} \;

# 更新usecase导入路径
find . -name "*.go" -type f ! -path "./usecase/shared/*" ! -path "./infrastructure/*" -exec sed -i '' \
  -e 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_route_usecase|github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/music/media_library/media_library_audio/core|g' \
  {} \;

# 更新controller导入路径
find . -name "*.go" -type f ! -path "./infrastructure/*" -exec sed -i '' \
  -e 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/controller_file_entity/scene_audio_route_api_controller|github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller/music/media_library/media_library_audio|g' \
  {} \;

# 更新mongo导入路径
find . -name "*.go" -type f ! -path "./infrastructure/*" -exec sed -i '' \
  -e 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo|github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo|g' \
  {} \;

# 更新internal导入路径
find . -name "*.go" -type f ! -path "./infrastructure/*" -exec sed -i '' \
  -e 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/internal|github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/internal|g' \
  {} \;

# 更新scene_audio_db_models引用
find . -name "*.go" -type f -exec sed -i '' \
  -e 's|scene_audio_db_models\.|core\.|g' \
  -e 's|scene_audio_db_interface\.|core\.|g' \
  {} \;

echo "import路径更新完成！"
