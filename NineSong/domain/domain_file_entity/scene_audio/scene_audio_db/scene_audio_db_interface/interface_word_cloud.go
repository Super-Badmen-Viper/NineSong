package scene_audio_db_interface

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
)

type WordCloudRepository interface {
	BulkUpsert(ctx context.Context, files []*scene_audio_db_models.WordCloudMetadata) (int, error)
	AllDelete(ctx context.Context) error
	DropAllIndex(ctx context.Context) error
	CreateIndex(ctx context.Context, fieldName string, unique bool) error
	GetAll(ctx context.Context) ([]*scene_audio_db_models.WordCloudMetadata, error)
}
