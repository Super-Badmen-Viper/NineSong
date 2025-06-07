package scene_audio_db_repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type mediaFileRepository struct {
	db         mongo.Database
	collection string
}

func NewMediaFileRepository(db mongo.Database, collection string) scene_audio_db_interface.MediaFileRepository {
	return &mediaFileRepository{
		db:         db,
		collection: collection,
	}
}

func (r *mediaFileRepository) Upsert(ctx context.Context, file *scene_audio_db_models.MediaFileMetadata) (*scene_audio_db_models.MediaFileMetadata, error) {
	coll := r.db.Collection(r.collection)
	now := time.Now().UTC()

	filter := bson.M{
		"path": file.Path,
	}

	update := file.ToUpdateDoc()
	update["$setOnInsert"] = bson.M{
		"created_at": now,
	}

	opts := options.Update().SetUpsert(true)
	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("upsert操作失败: %w", err)
	}

	if result.UpsertedID != nil {
		file.ID = result.UpsertedID.(primitive.ObjectID)
		file.CreatedAt = now
	} else {
		var existing struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		err := coll.FindOne(
			ctx,
			bson.M{"path": file.Path},
		).Decode(&existing)

		if err == nil {
			file.ID = existing.ID
		} else if errors.Is(err, driver.ErrNoDocuments) {
			return nil, fmt.Errorf("文档既未插入也未更新: %w", err)
		} else {
			return nil, fmt.Errorf("ID查询失败: %w", err)
		}
	}

	file.UpdatedAt = now
	return file, nil
}

func (r *mediaFileRepository) BulkUpsert(ctx context.Context, files []*scene_audio_db_models.MediaFileMetadata) (int, error) {
	coll := r.db.Collection(r.collection)
	var successCount int

	for _, file := range files {
		filter := bson.M{"_id": file.ID}
		update := bson.M{"$set": file}

		_, err := coll.UpdateOne(
			ctx,
			filter,
			update,
			options.Update().SetUpsert(true),
		)

		if err != nil {
			return successCount, fmt.Errorf("bulk upsert失败于索引%d: %w", successCount, err)
		}
		successCount++
	}
	return successCount, nil
}

func (r *mediaFileRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete media file failed: %w", err)
	}
	return nil
}

func (r *mediaFileRepository) DeleteByPath(ctx context.Context, path string) error {
	coll := r.db.Collection(r.collection)
	_, err := coll.DeleteOne(ctx, bson.M{"path": path})
	if err != nil {
		return fmt.Errorf("delete by path failed: %w", err)
	}
	return nil
}

func (r *mediaFileRepository) DeleteAllInvalid(
	ctx context.Context,
	filePaths []string,
	folderPath string,
) (int64, []struct {
	ArtistID primitive.ObjectID
	Count    int64
}, error) {
	coll := r.db.Collection(r.collection)
	// 创建返回结构：艺术家ID与删除数量的键值对数组
	deletedArtists := make([]struct {
		ArtistID primitive.ObjectID
		Count    int64
	}, 0)

	// 处理全量删除场景
	if filePaths == nil || len(filePaths) == 0 {
		filter := bson.M{"library_path": folderPath}
		// 新增：查询待删除文档的艺术家ID
		artistCounts := make(map[primitive.ObjectID]int64)
		cur, err := coll.Find(ctx, filter, options.Find().SetProjection(bson.M{"artist_id": 1}))
		if err == nil {
			defer cur.Close(ctx)
			for cur.Next(ctx) {
				var doc struct {
					ArtistID primitive.ObjectID `bson:"artist_id"`
				}
				if err := cur.Decode(&doc); err == nil {
					artistCounts[doc.ArtistID]++ // 统计艺术家关联文档数
				}
			}
		}

		// 执行删除
		delResult, err := coll.DeleteMany(ctx, filter)
		if err != nil {
			return 0, deletedArtists, fmt.Errorf("全量删除失败: %w", err)
		}

		// 构建艺术家删除统计
		for artistID, count := range artistCounts {
			deletedArtists = append(deletedArtists, struct {
				ArtistID primitive.ObjectID
				Count    int64
			}{ArtistID: artistID, Count: count})
		}
		return delResult, deletedArtists, nil
	}

	// 构建有效路径集合
	validPathSet := make(map[string]struct{})
	for _, path := range filePaths {
		absPath := filepath.Clean(path)
		if strings.HasPrefix(absPath, folderPath) {
			validPathSet[absPath] = struct{}{}
		}
	}

	// 查询需删除文档（包含艺术家ID）
	filter := bson.M{"library_path": folderPath}
	opts := options.Find().SetProjection(bson.M{"_id": 1, "path": 1, "artist_id": 1})
	cur, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return 0, deletedArtists, fmt.Errorf("查询失败: %w", err)
	}
	defer cur.Close(ctx)

	// 按艺术家分组待删除项
	artistToDelete := make(map[primitive.ObjectID][]primitive.ObjectID)
	var toDelete []primitive.ObjectID

	for cur.Next(ctx) {
		var doc struct {
			ID       primitive.ObjectID `bson:"_id"`
			Path     string             `bson:"path"`
			ArtistID primitive.ObjectID `bson:"artist_id"`
		}
		if err := cur.Decode(&doc); err != nil {
			continue
		}

		cleanPath := filepath.Clean(doc.Path)
		if _, exists := validPathSet[cleanPath]; !exists {
			toDelete = append(toDelete, doc.ID)
			artistToDelete[doc.ArtistID] = append(artistToDelete[doc.ArtistID], doc.ID)
		}
	}

	// 批量删除并统计
	totalDeleted := int64(0)
	batchSize := 1000
	artistCounts := make(map[primitive.ObjectID]int64)

	for i := 0; i < len(toDelete); i += batchSize {
		end := i + batchSize
		if end > len(toDelete) {
			end = len(toDelete)
		}

		batch := toDelete[i:end]
		delResult, err := coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": batch}})
		if err != nil {
			return totalDeleted, deletedArtists, fmt.Errorf("批量删除失败: %w", err)
		}
		totalDeleted += delResult

		// 统计本批次中各艺术家的删除数量
		for _, id := range batch {
			for artistID, ids := range artistToDelete {
				for _, artistDocID := range ids {
					if artistDocID == id {
						artistCounts[artistID]++
					}
				}
			}
		}
	}

	// 构建艺术家删除统计
	for artistID, count := range artistCounts {
		deletedArtists = append(deletedArtists, struct {
			ArtistID primitive.ObjectID
			Count    int64
		}{ArtistID: artistID, Count: count})
	}

	return totalDeleted, deletedArtists, nil
}

func (r *mediaFileRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*scene_audio_db_models.MediaFileMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"_id": id})

	var file scene_audio_db_models.MediaFileMetadata
	if err := result.Decode(&file); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get media file failed: %w", err)
	}
	return &file, nil
}

func (r *mediaFileRepository) GetByPath(ctx context.Context, path string) (*scene_audio_db_models.MediaFileMetadata, error) {
	coll := r.db.Collection(r.collection)
	result := coll.FindOne(ctx, bson.M{"path": path})

	var file scene_audio_db_models.MediaFileMetadata
	if err := result.Decode(&file); err != nil {
		if domain.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get by path failed: %w", err)
	}
	return &file, nil
}

func (r *mediaFileRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (bool, error) {
	coll := r.db.Collection(r.collection)

	// 构建原子更新操作
	result, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
		options.Update().SetUpsert(false),
	)

	if err != nil {
		return false, fmt.Errorf("媒体文件更新失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return false, nil
	}

	return true, nil
}

func (r *mediaFileRepository) MediaCountByArtist(
	ctx context.Context,
	artistID string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{
		"$or": []bson.M{
			{"artist_id": artistID},
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家单曲数量失败: %w", err)
	}

	return count, nil
}

func (r *mediaFileRepository) GuestMediaCountByArtist(
	ctx context.Context,
	artistID string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	// 构造复合查询条件
	filter := bson.M{
		"$and": []bson.M{
			{"artist_id": bson.M{"$ne": artistID}}, // 排除主导者[3](@ref)
			{"all_artist_ids": bson.M{ // 匹配合作者[4](@ref)
				"$elemMatch": bson.M{
					"artist_id": artistID,
				},
			}},
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计艺术家合作单曲失败: %w", err)
	}

	return count, nil
}

func (r *mediaFileRepository) MediaCountByAlbum(
	ctx context.Context,
	albumID string,
) (int64, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{
		"$or": []bson.M{
			{"album_id": albumID},
		},
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("统计专辑单曲数量失败: %w", err)
	}

	return count, nil
}

func takeMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *mediaFileRepository) inspectMedia(
	ctx context.Context,
	filter bson.M,
	validFilePaths []string,
	folderPath string,
	collection string,
) (int, error) {
	// 构建有效路径集合（只包含指定目录下的文件）
	validPathSet := make(map[string]struct{})
	for _, path := range validFilePaths {
		if strings.HasPrefix(path, folderPath) {
			validPathSet[path] = struct{}{}
		}
	}

	if validPathSet != nil && len(validPathSet) > 0 {
		coll := r.db.Collection(collection)

		// 精确查询条件（三重过滤）
		invalidDocsFilter := bson.M{
			"$and": []bson.M{
				filter, // 原始过滤条件
				{"path": bson.M{"$regex": "^" + regexp.QuoteMeta(folderPath)}}, // 路径必须在指定目录下
				{"path": bson.M{"$nin": validFilePaths}},                       // 路径不在有效列表中
			},
		}

		// 直接获取需要删除的文档ID
		cur, err := coll.Find(
			ctx,
			invalidDocsFilter,
			options.Find().SetProjection(bson.M{"_id": 1, "path": 1}),
		)
		if err != nil {
			return 0, fmt.Errorf("查询无效文档失败: %w", err)
		}
		defer cur.Close(ctx)

		// 批量收集无效文档ID
		var toDelete []primitive.ObjectID
		var invalidPaths []string
		for cur.Next(ctx) {
			var doc struct {
				ID   primitive.ObjectID `bson:"_id"`
				Path string             `bson:"path"`
			}
			if err := cur.Decode(&doc); err != nil {
				continue
			}
			toDelete = append(toDelete, doc.ID)
			invalidPaths = append(invalidPaths, doc.Path)
		}

		// 记录无效路径日志
		if len(invalidPaths) > 0 {
			log.Printf("在目录[%s]下检测到 %d 个无效媒体项: %v...",
				folderPath,
				len(invalidPaths),
				invalidPaths[:takeMin(5, len(invalidPaths))])
		}

		// 批量删除无效文档
		if len(toDelete) > 0 {
			batchSize := 1000
			for i := 0; i < len(toDelete); i += batchSize {
				end := i + batchSize
				if end > len(toDelete) {
					end = len(toDelete)
				}
				_, err = coll.DeleteMany(
					ctx,
					bson.M{"_id": bson.M{"$in": toDelete[i:end]}},
				)
				if err != nil {
					log.Printf("部分删除失败: %v", err)
				}
			}
			return len(toDelete), nil
		} else {
			return 0, nil // 没有无效项
		}
	}
	return -1, nil // 无效路径集合为空，返回-1表示直接标记删除
}

func (r *mediaFileRepository) InspectMediaCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
	folderPath string,
) (int, error) {
	filter := bson.M{"artist_id": artistID}
	return r.inspectMedia(ctx, filter, filePaths, folderPath, r.collection)
}

func (r *mediaFileRepository) InspectGuestMediaCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
	folderPath string,
) (int, error) {
	filter := bson.M{
		"artist_id":      bson.M{"$ne": artistID},
		"all_artist_ids": bson.M{"$elemMatch": bson.M{"artist_id": artistID}},
	}
	return r.inspectMedia(ctx, filter, filePaths, folderPath, r.collection)
}

func (r *mediaFileRepository) InspectMediaCountByAlbum(
	ctx context.Context,
	albumID string,
	filePaths []string,
	folderPath string,
) (int, error) {
	filter := bson.M{"album_id": albumID}
	return r.inspectMedia(ctx, filter, filePaths, folderPath, r.collection)
}

func (r *mediaFileRepository) InspectGuestMediaCountByAlbum(
	ctx context.Context,
	artistID string,
	filePaths []string,
	folderPath string,
) (int, error) {
	filter := bson.M{
		"artist_id":            bson.M{"$ne": artistID},
		"all_album_artist_ids": bson.M{"$elemMatch": bson.M{"artist_id": artistID}},
	}
	return r.inspectMedia(ctx, filter, filePaths, folderPath, r.collection)
}
