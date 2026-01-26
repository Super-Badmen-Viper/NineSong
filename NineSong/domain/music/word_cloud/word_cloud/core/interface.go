package core

import (
	"context"
)

// WordCloudRepository 词云仓库接口
type WordCloudRepository interface {
	BulkUpsert(ctx context.Context, files []*WordCloudMetadata) (int, error)
	AllDelete(ctx context.Context) error
	DropAllIndex(ctx context.Context) error
	CreateIndex(ctx context.Context, fieldName string, unique bool) error
	GetAll(ctx context.Context) ([]*WordCloudMetadata, error)
}
