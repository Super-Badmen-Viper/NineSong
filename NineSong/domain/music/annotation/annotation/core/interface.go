package core

import (
	"context"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnnotationRepository 注释仓库接口
type AnnotationRepository interface {
	shared.BaseRepository[AnnotationMetadata] // 嵌入基础Repository接口

	// 特定方法
	GetByPath(ctx context.Context, path string) (*AnnotationMetadata, error)
	DeleteByPath(ctx context.Context, path string) error
}
