package core

import (
	"context"
	"fmt"

	annotationCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/annotation/annotation/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/shared"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type annotationRepository struct {
	db         mongo.Database
	collection string
}

func NewAnnotationRepository(db mongo.Database, collection string) annotationCore.AnnotationRepository {
	return &annotationRepository{
		db:         db,
		collection: collection,
	}
}

// GetByPath 根据路径获取注释
func (r *annotationRepository) GetByPath(ctx context.Context, path string) (*annotationCore.AnnotationMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"path": path})

	var annotation annotationCore.AnnotationMetadata
	if err := result.Decode(&annotation); err != nil {
		if shared.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get annotation by path failed: %w", err)
	}
	return &annotation, nil
}

// DeleteByPath 根据路径删除注释
func (r *annotationRepository) DeleteByPath(ctx context.Context, path string) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"path": path})
	if err != nil {
		return fmt.Errorf("delete annotation by path failed: %w", err)
	}
	return nil
}
