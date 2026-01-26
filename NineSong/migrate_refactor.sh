#!/bin/bash

# 项目重构迁移脚本
# 此脚本用于批量迁移文件到新的目录结构

# 设置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}开始项目重构迁移...${NC}"

# 迁移domain层的media_library_audio
echo -e "${YELLOW}迁移 domain media_library_audio...${NC}"
mkdir -p domain/music/media_library/media_library_audio/core
mkdir -p domain/music/media_library/media_library_audio/utils

# 迁移模型文件
if [ -f "domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models/model_media_library_audio.go" ]; then
    cp domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models/model_media_library_audio.go \
       domain/music/media_library/media_library_audio/core/model.go
    sed -i '' 's/package scene_audio_db_models/package core/g' \
       domain/music/media_library/media_library_audio/core/model.go
    echo -e "${GREEN}✓ 迁移 model_media_library_audio.go${NC}"
fi

# 迁移接口文件
if [ -f "domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface/interface_media_library_audio.go" ]; then
    cp domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface/interface_media_library_audio.go \
       domain/music/media_library/media_library_audio/core/interface.go
    sed -i '' 's/package scene_audio_db_interface/package core/g' \
       domain/music/media_library/media_library_audio/core/interface.go
    # 更新import路径
    sed -i '' 's|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain|github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared|g' \
       domain/music/media_library/media_library_audio/core/interface.go
    sed -i '' 's|scene_audio_db_models|core|g' \
       domain/music/media_library/media_library_audio/core/interface.go
    echo -e "${GREEN}✓ 迁移 interface_media_library_audio.go${NC}"
fi

# 创建统一导出文件
cat > domain/music/media_library/media_library_audio/media_library_audio.go << 'EOF'
package media_library_audio

// 导出core中的所有接口和类型
import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/media_library_audio/core"
)

// 重新导出核心类型
type MediaLibraryAudio = core.MediaLibraryAudio
type MediaLibraryAudioRepository = core.MediaLibraryAudioRepository

// 重新导出核心函数
var Validate = core.Validate
var SetTimestamps = core.SetTimestamps
EOF

echo -e "${GREEN}✓ 创建统一导出文件${NC}"

echo -e "${GREEN}迁移完成！${NC}"
echo -e "${YELLOW}注意：请手动检查并更新所有import路径${NC}"
