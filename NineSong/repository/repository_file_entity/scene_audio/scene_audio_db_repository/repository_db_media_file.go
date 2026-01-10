package scene_audio_db_repository

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/yanyiwu/gojieba"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mediaFileRepository struct {
	db                 mongo.Database
	collection         string
	jieba              *gojieba.Jieba
	stopWords          map[string]bool
	wordCloudMutex     sync.Mutex
	wordCloudCancel    context.CancelFunc
	isWordCloudRunning bool
}

func NewMediaFileRepository(db mongo.Database, collection string) scene_audio_db_interface.MediaFileRepository {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		jieba := gojieba.NewJieba()
		stopWords := domain_util.LoadCombinedStopWords()

		return &mediaFileRepository{
			db:         db,
			collection: collection,
			jieba:      jieba,
			stopWords:  stopWords,
		}
	} else {
		// MacOS Build
		dictPath := os.Getenv("JIEBA_DICT_PATH")
		if dictPath == "" {
			dictPath = "/app/jieba-dict"
		}

		log.Printf("正在使用jieba词典路径: %s", dictPath)

		jieba := gojieba.NewJieba(
			dictPath+"/jieba.dict.utf8",
			dictPath+"/hmm_model.utf8",
			dictPath+"/user.dict.utf8",
			dictPath+"/idf.utf8",
			dictPath+"/stop_words.utf8",
		)

		stopWords := domain_util.LoadCombinedStopWords()

		return &mediaFileRepository{
			db:         db,
			collection: collection,
			jieba:      jieba,
			stopWords:  stopWords,
		}
	}
}

func (r *mediaFileRepository) GetAllGenre(ctx context.Context) ([]scene_audio_db_models.WordCloudMetadata, error) {
	coll := r.db.Collection(r.collection)

	// 构建聚合管道获取所有流派及其计数
	pipeline := bson.A{
		bson.D{{"$match", bson.D{
			{"genre", bson.D{{"$ne", ""}}}, // 过滤掉空流派
		}}},
		bson.D{{"$project", bson.D{
			{"genres", bson.D{
				{"$split", bson.A{"$genre", ";"}}, // 拆分多流派字符串
			}},
		}}},
		bson.D{{"$unwind", "$genres"}}, // 展开流派数组
		bson.D{{"$group", bson.D{
			{"_id", bson.D{
				{"$trim", bson.D{
					{"input", "$genres"},
					{"chars", " "}, // 去除前后空格
				}},
			}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		bson.D{{"$sort", bson.D{{"count", -1}}}}, // 按数量降序
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("流派聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析聚合结果
	var rawResults []struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}
	if err := cursor.All(ctx, &rawResults); err != nil {
		return nil, fmt.Errorf("解析流派数据失败: %w", err)
	}

	// 转换为WordCloudMetadata结构
	results := make([]scene_audio_db_models.WordCloudMetadata, len(rawResults))
	for i, item := range rawResults {
		results[i] = scene_audio_db_models.WordCloudMetadata{
			ID:    primitive.NewObjectID(),
			Name:  item.ID,
			Count: item.Count,
			Type:  "genre", // 明确标记为流派类型
			Rank:  i + 1,   // 根据排序设置排名
		}
	}

	return results, nil
}

func (r *mediaFileRepository) GetHighFrequencyWords(
	ctx context.Context,
	limit int,
) ([]scene_audio_db_models.WordCloudMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Concurrency and Cancellation Control
	r.wordCloudMutex.Lock()
	if r.isWordCloudRunning {
		if r.wordCloudCancel != nil {
			r.wordCloudCancel() // Cancel the previous running task
		}
	}

	r.wordCloudCancel = cancel
	r.isWordCloudRunning = true
	r.wordCloudMutex.Unlock()

	defer func() {
		r.wordCloudMutex.Lock()
		r.isWordCloudRunning = false
		r.wordCloudMutex.Unlock()
		cancel() // Ensure the context is cancelled on function exit
	}()

	// 2. Database Aggregation (Single Hit)
	coll := r.db.Collection(r.collection)
	pipeline := bson.A{
		bson.D{{"$project", bson.D{
			{"textField", bson.D{
				{"$concat", bson.A{
					bson.D{{"$ifNull", bson.A{"$title", ""}}}, " ",
					bson.D{{"$ifNull", bson.A{"$artist", ""}}}, " ",
					bson.D{{"$ifNull", bson.A{"$album", ""}}}, " ",
					bson.D{{"$ifNull", bson.A{"$lyrics", ""}}},
				}},
			}},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("文本字段聚合失败: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []struct{ TextField string }
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("读取聚合结果失败: %w", err)
	}

	// 3. Concurrent Worker Pool for Processing
	numWorkers := runtime.NumCPU()
	jobs := make(chan string, len(documents))
	results := make(chan map[string]int, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localWordMap := make(map[string]int)
			numericRegex := regexp.MustCompile(`^[0-9.]+$`)

			for docText := range jobs {
				select {
				case <-ctx.Done(): // Check for cancellation signal
					return
				default:
					words := r.jieba.Cut(docText, true)
					for _, word := range words {
						// Normalize word to lowercase
						lowerWord := strings.ToLower(word)
						// Filter stopwords, short words, and numeric strings
						if !r.stopWords[lowerWord] && utf8.RuneCountInString(lowerWord) > 1 && !numericRegex.MatchString(lowerWord) {
							localWordMap[lowerWord]++
						}
					}
				}
			}
			results <- localWordMap
		}()
	}

	for _, doc := range documents {
		jobs <- doc.TextField
	}
	close(jobs)

	wg.Wait()
	close(results)

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 4. Aggregate final results
	globalWordMap := make(map[string]int)
	for localMap := range results {
		for word, count := range localMap {
			globalWordMap[word] += count
		}
	}

	// 5. Extract Top-K using MinHeap
	h := &domain_util.MinHeap{}
	heap.Init(h)
	for word, count := range globalWordMap {
		if h.Len() < limit {
			heap.Push(h, domain_util.WordCount{Word: word, Count: count})
		} else if count > (*h)[0].Count {
			heap.Pop(h)
			heap.Push(h, domain_util.WordCount{Word: word, Count: count})
		}
	}

	// 6. Format and sort the final list
	wordCounts := make([]scene_audio_db_models.WordCloudMetadata, h.Len())
	tempList := make([]domain_util.WordCount, h.Len())
	for i := 0; h.Len() > 0; i++ {
		tempList[i] = heap.Pop(h).(domain_util.WordCount)
	}

	sort.Slice(tempList, func(i, j int) bool {
		return tempList[i].Count > tempList[j].Count
	})

	for i, wc := range tempList {
		wordCounts[i] = scene_audio_db_models.WordCloudMetadata{
			ID:    primitive.NewObjectID(),
			Name:  wc.Word,
			Count: wc.Count,
			Type:  "media_file",
			Rank:  i + 1,
		}
	}

	return wordCounts, nil
}

func (r *mediaFileRepository) GetRecommendedByKeywords(
	ctx context.Context,
	keywords []string,
	limit int,
) ([]scene_audio_db_models.WordCloudRecommendation, error) {
	coll := r.db.Collection(r.collection)
	if len(keywords) == 0 {
		return []scene_audio_db_models.WordCloudRecommendation{}, nil
	}

	// 1. 预过滤阶段（性能优化）
	orConditions := make([]bson.M, 0, len(keywords)*4)
	for _, kw := range keywords {
		regex := bson.M{"$regex": kw, "$options": "i"}
		orConditions = append(orConditions, bson.M{"title": regex})
		orConditions = append(orConditions, bson.M{"artist": regex})
		orConditions = append(orConditions, bson.M{"album": regex})
		orConditions = append(orConditions, bson.M{"genre": regex})
		orConditions = append(orConditions, bson.M{"lyrics": regex})
	}
	matchStage := bson.D{{"$match", bson.M{"$or": orConditions}}}

	// 2. 加权评分阶段
	weightedConditions := bson.A{}
	weights := map[string]int{"title": 3, "artist": 2, "album": 2, "lyrics": 1}
	for field, weight := range weights {
		for _, kw := range keywords {
			weightedConditions = append(weightedConditions, bson.M{
				"$cond": bson.A{
					bson.M{"$regexMatch": bson.M{
						"input":   "$" + field,
						"regex":   kw, // 关键修正：使用regex而非pattern
						"options": "i",
					}},
					weight, // 字段权重
					0,
				},
			})
		}
	}
	scoreStage := bson.D{{"$addFields", bson.M{
		"relevance_score": bson.M{"$sum": weightedConditions},
	}}}

	// 3. 排序与分页
	sortStage := bson.D{{"$sort", bson.D{
		{"relevance_score", -1},
		{"created_at", -1},
	}}}
	limitStage := bson.D{{"$limit", limit}}

	// 4. 执行聚合管道
	pipeline := driver.Pipeline{matchStage, scoreStage, sortStage, limitStage}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("推荐聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 5. 结果解码（包含评分）
	var files []struct {
		ID    primitive.ObjectID `bson:"_id"`
		Title string             `bson:"title"`
		Score float64            `bson:"relevance_score"`
	}
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("推荐结果解析失败: %w", err)
	}

	// 6. 转换为响应模型
	results := make([]scene_audio_db_models.WordCloudRecommendation, len(files))
	for i, f := range files {
		results[i] = scene_audio_db_models.WordCloudRecommendation{
			ID:    f.ID,
			Type:  "media_file",
			Name:  f.Title,
			Score: f.Score, // 保留原始评分
		}
	}
	return results, nil
}

func (r *mediaFileRepository) GetAllCounts(ctx context.Context) ([]scene_audio_db_models.MediaFileCounts, error) {
	coll := r.db.Collection(r.collection)

	filter := bson.M{}
	projection := bson.M{
		"_id":        1,
		"updated_at": 1,
	}

	cursor, err := coll.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []scene_audio_db_models.MediaFileCounts
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解码失败: %w", err)
	}

	return results, nil
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
) (int64, []struct {
	ArtistID primitive.ObjectID
	Count    int64
}, error) {
	coll := r.db.Collection(r.collection)
	deletedArtists := make([]struct {
		ArtistID primitive.ObjectID
		Count    int64
	}, 0)

	// 场景1：当filePaths为空时，不执行删除操作，避免误删所有文件
	if len(filePaths) == 0 {
		log.Printf("DeleteAllInvalid: filePaths为空，跳过删除操作")
		return 0, deletedArtists, nil
	}

	// 构建全局有效文件路径集合，使用正斜杠规范化
	validFilePaths := make(map[string]bool)
	for _, path := range filePaths {
		// 规范化文件路径，使用正斜杠格式，与数据库中存储的格式一致
		normalizedPath := filepath.ToSlash(filepath.Clean(path))
		validFilePaths[normalizedPath] = true
	}

	cur, err := coll.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1, "path": 1, "artist_id": 1}))
	if err != nil {
		return 0, deletedArtists, fmt.Errorf("查询失败: %w", err)
	}
	defer cur.Close(ctx)

	artistToIDs := make(map[primitive.ObjectID][]primitive.ObjectID)
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

		// 检查文件路径是否在有效文件路径集合中，或者是否实际存在
		isValid := false

		// 检查文件是否在有效文件路径集合中
		if validFilePaths[doc.Path] {
			isValid = true
		} else {
			// 检查文件是否实际存在于文件系统中
			absPath := doc.Path
			// 将正斜杠转换为反斜杠，适配Windows系统
			absPath = strings.ReplaceAll(absPath, "/", "\\")
			if _, err := os.Stat(absPath); err == nil {
				// 文件实际存在，是有效文件
				isValid = true
			}
		}

		// 如果文件既不在有效文件路径集合中，也不存在于文件系统中，才是无效文件
		if !isValid {
			toDelete = append(toDelete, doc.ID)
			artistToIDs[doc.ArtistID] = append(artistToIDs[doc.ArtistID], doc.ID)
		}
	}

	// 批量删除并统计（优化性能）
	totalDeleted := int64(0)
	const batchSize = 1000
	artistCounts := make(map[primitive.ObjectID]int64)

	// 使用预统计避免嵌套循环
	for artistID, ids := range artistToIDs {
		artistCounts[artistID] = int64(len(ids))
	}

	// 批量删除
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

func (r *mediaFileRepository) DeleteByFolder(ctx context.Context, folderPath string) (int64, error) {
	coll := r.db.Collection(r.collection)

	// 标准化路径格式（确保以反斜杠结尾）
	normalizedFolderPath := strings.Replace(folderPath, "/", "\\", -1)
	if !strings.HasSuffix(normalizedFolderPath, "\\") {
		normalizedFolderPath += "\\"
	}

	// 构建精确匹配library_path的正则表达式
	regexPattern := regexp.QuoteMeta(normalizedFolderPath)
	filter := bson.M{
		"library_path": bson.M{
			"$regex":   "^" + regexPattern,
			"$options": "i", // 不区分大小写
		},
	}

	// 执行删除操作
	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文件夹内容失败: %w", err)
	}

	return result, nil
}

// DeleteAll 删除所有媒体文件
func (r *mediaFileRepository) DeleteAll(ctx context.Context) (int64, error) {
	coll := r.db.Collection(r.collection)
	result, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("删除所有媒体文件失败: %w", err)
	}
	return result, nil
}

func (r *mediaFileRepository) GetAll(ctx context.Context) ([]*scene_audio_db_models.MediaFileMetadata, error) {
	coll := r.db.Collection(r.collection)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("获取所有媒体文件失败: %w", err)
	}
	defer cursor.Close(ctx)

	var files []*scene_audio_db_models.MediaFileMetadata
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("解析媒体文件失败: %w", err)
	}
	return files, nil
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

func (r *mediaFileRepository) GetByFolder(ctx context.Context, folderPath string) ([]string, error) {
	coll := r.db.Collection(r.collection)

	// 规范化文件夹路径，使用正斜杠格式
	normalizedFolderPath := filepath.ToSlash(folderPath)
	if !strings.HasSuffix(normalizedFolderPath, "/") {
		normalizedFolderPath += "/"
	}

	// 使用path字段进行前缀查询
	filter := bson.M{
		"path": bson.M{
			"$regex":   "^" + regexp.QuoteMeta(normalizedFolderPath),
			"$options": "i",
		},
	}

	// 只返回path字段
	opts := options.Find().SetProjection(bson.M{"path": 1})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询文件夹内容失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 提取路径结果
	var results []string
	for cursor.Next(ctx) {
		var item struct {
			Path string `bson:"path"`
		}
		if err := cursor.Decode(&item); err != nil {
			log.Printf("解码路径失败: %v", err)
			continue
		}
		// 规范化路径，确保使用正斜杠格式
		normalizedPath := filepath.ToSlash(item.Path)
		results = append(results, normalizedPath)
	}

	return results, nil
}

func (r *mediaFileRepository) GetFilesWithMissingMetadata(ctx context.Context, folderPath string, folderType int) ([]*scene_audio_db_models.MediaFileMetadata, error) {
	coll := r.db.Collection(r.collection)

	// 1. 获取数据库中所有文件的最新更新时间
	var latestUpdatedAt time.Time
	// 按updated_at降序排序，获取第一条记录
	findOpts := options.Find().SetSort(bson.D{{"updated_at", -1}}).SetProjection(bson.D{{"updated_at", 1}}).SetLimit(1)
	var latestRecord struct {
		UpdatedAt time.Time `bson:"updated_at"`
	}
	cursor, err := coll.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		if errors.Is(err, driver.ErrNoDocuments) {
			// 数据库中没有记录，返回空列表
			return []*scene_audio_db_models.MediaFileMetadata{}, nil
		} else {
			return nil, fmt.Errorf("查询最新更新时间失败: %w", err)
		}
	} else {
		// 遍历游标获取第一条记录
		if cursor.Next(ctx) {
			if err := cursor.Decode(&latestRecord); err != nil {
				return nil, fmt.Errorf("解码最新更新时间失败: %w", err)
			}
			latestUpdatedAt = latestRecord.UpdatedAt
		} else {
			// 数据库中没有记录，返回空列表
			return []*scene_audio_db_models.MediaFileMetadata{}, nil
		}
		cursor.Close(ctx)
	}

	log.Printf("DEBUG - 数据库最新更新时间: %v", latestUpdatedAt)

	// 2. 遍历文件夹，收集所有符合类型的文件及其修改时间
	fileModTimeMap := make(map[string]time.Time)

	// 根据folderType确定要处理的文件扩展名
	var allowedExts map[string]bool
	switch folderType {
	case 1: // 音频文件夹
		allowedExts = map[string]bool{".mp3": true, ".flac": true, ".wav": true, ".ape": true, ".m4a": true, ".ogg": true, ".aac": true, ".wma": true, ".opus": true, ".alac": true}
	case 2: // 视频文件夹
		allowedExts = map[string]bool{".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true}
	case 3: // 图像文件夹
		allowedExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true, ".tiff": true}
	case 4: // 文档文件夹
		allowedExts = map[string]bool{".txt": true, ".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".md": true}
	default: // 默认只处理音频文件
		allowedExts = map[string]bool{".mp3": true, ".flac": true, ".wav": true, ".ape": true, ".m4a": true, ".ogg": true, ".aac": true}
	}

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("访问路径 %s 出错: %v", path, err)
			return nil // 跳过错误继续遍历
		}
		if info.IsDir() {
			return nil
		}

		// 只处理符合类型的文件
		ext := strings.ToLower(filepath.Ext(path))
		if allowedExts[ext] {
			// 规范化文件路径，转换为正斜杠格式，与数据库存储格式一致
			normalizedPath := filepath.ToSlash(filepath.Clean(path))
			// 将文件修改时间转换为UTC时间，与数据库存储格式一致
			fileModTime := info.ModTime().UTC()
			// 只收集修改时间大于数据库最新更新时间的文件
			if fileModTime.After(latestUpdatedAt) {
				fileModTimeMap[normalizedPath] = fileModTime
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("遍历文件夹失败: %w", err)
	}

	log.Printf("DEBUG - 发现需要更新的文件数量: %d", len(fileModTimeMap))

	// 如果没有需要更新的文件，直接返回
	if len(fileModTimeMap) == 0 {
		return []*scene_audio_db_models.MediaFileMetadata{}, nil
	}

	// 3. 从数据库中获取这些文件的更新时间，进一步验证
	filePaths := make([]string, 0, len(fileModTimeMap))
	for path := range fileModTimeMap {
		filePaths = append(filePaths, path)
	}

	// 查询数据库中这些文件的更新时间
	filter := bson.M{
		"path": bson.M{"$in": filePaths},
	}
	findOpts = options.Find().SetProjection(bson.M{"path": 1, "updated_at": 1})

	cursor, err = coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("查询文件更新时间失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 构建数据库文件更新时间映射
	dbModTimeMap := make(map[string]time.Time)
	for cursor.Next(ctx) {
		var item struct {
			Path      string    `bson:"path"`
			UpdatedAt time.Time `bson:"updated_at"`
		}
		if err := cursor.Decode(&item); err != nil {
			log.Printf("解码文件更新时间失败: %v", err)
			continue
		}
		// 确保数据库中的路径也使用正斜杠格式
		normalizedPath := filepath.ToSlash(item.Path)
		dbModTimeMap[normalizedPath] = item.UpdatedAt
	}

	// 4. 比对时间，找出真正需要更新的文件
	var filesToUpdate []*scene_audio_db_models.MediaFileMetadata
	// 添加1秒的时间误差范围，处理不同系统间的时间精度差异
	timeError := 1 * time.Second
	for path, fileModTime := range fileModTimeMap {
		if dbModTime, exists := dbModTimeMap[path]; !exists || fileModTime.After(dbModTime.Add(timeError)) {
			// 文件不存在于数据库或文件修改时间大于数据库更新时间（考虑1秒误差），需要更新元数据
			filesToUpdate = append(filesToUpdate, &scene_audio_db_models.MediaFileMetadata{
				Path: path,
			})
		}
	}

	log.Printf("DEBUG - 最终需要更新的文件数量: %d", len(filesToUpdate))

	return filesToUpdate, nil
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

func (r *mediaFileRepository) inspectMedia(
	ctx context.Context,
	filter bson.M,
	validFilePaths []string, // 全局有效路径集合
	collection string,
	clean bool, // 是否清理无效路径
) (int, error) {
	// 构建全局有效目录集合，使用正斜杠规范化
	validDirectories := make([]string, 0, len(validFilePaths))
	for _, path := range validFilePaths {
		// 规范化文件路径，使用正斜杠格式，与数据库中存储的格式一致
		normalizedPath := filepath.ToSlash(filepath.Clean(path))
		// 确保目录路径以斜杠结尾
		if !strings.HasSuffix(normalizedPath, "/") {
			normalizedPath += "/"
		}
		validDirectories = append(validDirectories, normalizedPath)
	}

	// 没有有效路径时返回0，表示没有无效文件，避免误删
	if len(validDirectories) == 0 {
		log.Printf("inspectMedia: validFilePaths为空，返回0表示没有无效文件")
		return 0, nil
	}

	coll := r.db.Collection(collection)

	// 批量获取数据库中所有文件的路径（仅获取ID和Path字段）
	allFilesCur, err := coll.Find(
		ctx,
		filter,
		options.Find().SetProjection(bson.M{"_id": 1, "path": 1}),
	)
	if err != nil {
		return 0, fmt.Errorf("查询所有文档失败: %w", err)
	}
	defer allFilesCur.Close(ctx)

	// 批量收集无效文档ID和路径
	var toDelete []primitive.ObjectID
	var invalidPaths []string

	// 检查文件路径是否位于有效目录下
	for allFilesCur.Next(ctx) {
		var doc struct {
			ID   primitive.ObjectID `bson:"_id"`
			Path string             `bson:"path"`
		}
		if err := allFilesCur.Decode(&doc); err != nil {
			continue
		}

		// 检查文件路径是否位于任何一个有效目录下
		isValid := false
		for _, dir := range validDirectories {
			if strings.HasPrefix(doc.Path, dir) {
				// 文件路径位于有效目录下，是有效文件
				isValid = true
				break
			}
		}

		// 如果文件路径不位于任何有效目录下，检查文件是否实际存在
		if !isValid {
			// 文件可能不在有效目录下，但可能是通过其他方式添加的有效文件
			// 进行最终验证：检查文件是否实际存在于文件系统中
			absPath := doc.Path
			// 将正斜杠转换为反斜杠，适配Windows系统
			absPath = strings.ReplaceAll(absPath, "/", "\\")
			if _, err := os.Stat(absPath); err == nil {
				// 文件实际存在，是有效文件
				isValid = true
			}
		}

		// 如果文件确实无效，添加到删除列表
		if !isValid {
			if clean {
				toDelete = append(toDelete, doc.ID)
			}
			invalidPaths = append(invalidPaths, doc.Path)
		}
	}

	// 只有当确实有无效文件时才记录日志
	if len(invalidPaths) > 0 {
		// 优化日志输出，只在调试模式下输出详细信息
		// 使用更简洁的日志格式，避免过多信息
		log.Printf("检测到 %d 个无效媒体项，示例: %v...",
			len(invalidPaths),
			invalidPaths[:min(3, len(invalidPaths))])
	} else {
		return 0, nil // 没有无效项
	}

	// 批量删除无效文档
	if clean {
		if len(toDelete) > 0 {
			batchSize := 1000
			totalDeleted := 0

			for i := 0; i < len(toDelete); i += batchSize {
				end := i + batchSize
				if end > len(toDelete) {
					end = len(toDelete)
				}

				batch := toDelete[i:end]
				delResult, err := coll.DeleteMany(
					ctx,
					bson.M{"_id": bson.M{"$in": batch}},
				)
				if err != nil {
					log.Printf("部分删除失败: %v", err)
				} else {
					totalDeleted += int(delResult)
				}
			}
			return totalDeleted, nil
		}
	} else {
		if len(invalidPaths) > 0 {
			return len(invalidPaths), nil
		}
	}
	return 0, nil // 没有无效项
}

func (r *mediaFileRepository) InspectMediaCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
) (int, error) {
	filter := bson.M{"artist_id": artistID}
	return r.inspectMedia(ctx, filter, filePaths, r.collection, false)
}

func (r *mediaFileRepository) InspectGuestMediaCountByArtist(
	ctx context.Context,
	artistID string,
	filePaths []string,
) (int, error) {
	filter := bson.M{
		"artist_id":      bson.M{"$ne": artistID},
		"all_artist_ids": bson.M{"$elemMatch": bson.M{"artist_id": artistID}},
	}
	return r.inspectMedia(ctx, filter, filePaths, r.collection, false)
}

func (r *mediaFileRepository) InspectMediaCountByAlbum(
	ctx context.Context,
	albumID string,
	filePaths []string,
) (int, error) {
	filter := bson.M{"album_id": albumID}
	return r.inspectMedia(ctx, filter, filePaths, r.collection, false)
}

func (r *mediaFileRepository) InspectGuestMediaCountByAlbum(
	ctx context.Context,
	artistID string,
	filePaths []string,
) (int, error) {
	filter := bson.M{
		"artist_id":            bson.M{"$ne": artistID},
		"all_album_artist_ids": bson.M{"$elemMatch": bson.M{"artist_id": artistID}},
	}
	return r.inspectMedia(ctx, filter, filePaths, r.collection, false)
}
