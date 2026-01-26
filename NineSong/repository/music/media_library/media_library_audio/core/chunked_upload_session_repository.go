package core

import (
	"context"
	"time"

	domainCore "github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/music/media_library/media_library_audio/core"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/infrastructure/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type chunkedUploadSessionRepository struct {
	db         mongo.Database
	collection string
}

func NewChunkedUploadSessionRepository(db mongo.Database, collection string) domainCore.ChunkedUploadSessionRepository {
	return &chunkedUploadSessionRepository{
		db:         db,
		collection: collection,
	}
}

// BaseRepository methods
func (r *chunkedUploadSessionRepository) Create(ctx context.Context, entity *domainCore.ChunkedUploadSession) error {
	entity.SetTimestamps()
	_, err := r.db.Collection(r.collection).InsertOne(ctx, entity)
	return err
}

func (r *chunkedUploadSessionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domainCore.ChunkedUploadSession, error) {
	var entity domainCore.ChunkedUploadSession
	err := r.db.Collection(r.collection).FindOne(ctx, bson.M{"_id": id}).Decode(&entity)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r *chunkedUploadSessionRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(r.collection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *chunkedUploadSessionRepository) GetAll(ctx context.Context) ([]*domainCore.ChunkedUploadSession, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*domainCore.ChunkedUploadSession
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}

// Custom methods
func (r *chunkedUploadSessionRepository) GetByUploadID(ctx context.Context, uploadID string) (*domainCore.ChunkedUploadSession, error) {
	var entity domainCore.ChunkedUploadSession
	err := r.db.Collection(r.collection).FindOne(ctx, bson.M{"upload_id": uploadID}).Decode(&entity)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r *chunkedUploadSessionRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *chunkedUploadSessionRepository) UpdateUploadedChunks(ctx context.Context, id primitive.ObjectID, uploadedChunks int) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"uploaded_chunks": uploadedChunks,
			"updated_at":      primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *chunkedUploadSessionRepository) UpdateFilePath(ctx context.Context, id primitive.ObjectID, filePath string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"file_path":  filePath,
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *chunkedUploadSessionRepository) UpdateChecksum(ctx context.Context, id primitive.ObjectID, checksum string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"checksum":   checksum,
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *chunkedUploadSessionRepository) UpdateProgress(ctx context.Context, id primitive.ObjectID, progress float64) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"progress":   progress,
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *chunkedUploadSessionRepository) UpdateUploadedBytes(ctx context.Context, id primitive.ObjectID, uploadedBytes int64) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"uploaded_bytes": uploadedBytes,
			"updated_at":     primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

// BaseRepository implementation methods
func (r *chunkedUploadSessionRepository) Upsert(ctx context.Context, entity *domainCore.ChunkedUploadSession) error {
	entity.SetTimestamps()
	filter := bson.M{"upload_id": entity.UploadID}
	update := bson.M{"$set": entity}
	opts := options.Update().SetUpsert(true)
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *chunkedUploadSessionRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	filter := bson.M{"_id": id}
	update["$set"].(bson.M)["updated_at"] = primitive.NewDateTimeFromTime(time.Now())
	result, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

func (r *chunkedUploadSessionRepository) CreateMany(ctx context.Context, entities []*domainCore.ChunkedUploadSession) error {
	if len(entities) == 0 {
		return nil
	}

	for _, entity := range entities {
		entity.SetTimestamps()
	}

	documents := make([]interface{}, len(entities))
	for i, entity := range entities {
		documents[i] = entity
	}

	_, err := r.db.Collection(r.collection).InsertMany(ctx, documents)
	return err
}

func (r *chunkedUploadSessionRepository) BulkUpsert(ctx context.Context, entities []*domainCore.ChunkedUploadSession) (int, error) {
	coll := r.db.Collection(r.collection)
	var successCount int

	for _, entity := range entities {
		now := primitive.NewDateTimeFromTime(time.Now())
		if entity.ID.IsZero() {
			entity.CreatedAt = now
		}
		entity.UpdatedAt = now

		filter := bson.M{"_id": entity.ID}
		update := bson.M{"$set": entity}

		_, err := coll.UpdateOne(
			ctx,
			filter,
			update,
		)
		if err != nil {
			continue // 继续处理下一个实体，而不是返回错误
		}
		successCount++
	}

	return successCount, nil
}

func (r *chunkedUploadSessionRepository) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	result, err := r.db.Collection(r.collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *chunkedUploadSessionRepository) GetByFilter(ctx context.Context, filter interface{}) ([]*domainCore.ChunkedUploadSession, error) {
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*domainCore.ChunkedUploadSession
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *chunkedUploadSessionRepository) GetOneByFilter(ctx context.Context, filter interface{}) (*domainCore.ChunkedUploadSession, error) {
	var entity domainCore.ChunkedUploadSession
	err := r.db.Collection(r.collection).FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == driver.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r *chunkedUploadSessionRepository) Count(ctx context.Context, filter interface{}) (int64, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	return count, err
}

func (r *chunkedUploadSessionRepository) GetPaginated(ctx context.Context, filter interface{}, skip, limit int64) ([]*domainCore.ChunkedUploadSession, error) {
	opts := options.Find().SetSkip(skip).SetLimit(limit)
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*domainCore.ChunkedUploadSession
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *chunkedUploadSessionRepository) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*driver.UpdateResult, error) {
	return r.db.Collection(r.collection).UpdateMany(ctx, filter, update, opts...)
}

func (r *chunkedUploadSessionRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *chunkedUploadSessionRepository) ExistsByFilter(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
