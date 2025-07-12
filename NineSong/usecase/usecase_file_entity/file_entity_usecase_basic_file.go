package usecase_file_entity

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/mozillazg/go-pinyin"
	ffmpeggo "github.com/u2takey/ffmpeg-go"
	driver "go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_models"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio/scene_audio_db_usecase"
	"github.com/dhowden/tag"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScanManager 全局状态管理器
type ScanManager struct {
	mu                  sync.RWMutex
	globalScanRunning   bool                          // 全局扫描任务状态
	concurrentScanCount int                           // 并发扫描任务计数
	cancelFuncs         map[string]context.CancelFunc // 所有扫描任务的取消函数
}

type taskProgress struct {
	id             string
	totalFiles     int32      // 改为原子类型
	walkedFiles    int32      // 原子计数：已遍历文件数
	processedFiles int32      // 原子计数：已处理文件数
	mu             sync.Mutex // 新增互斥锁保护非原子字段
	initialized    bool       // 新增：标记是否已初始化
	status         string     // 新增：任务状态
}

func NewScanManager() *ScanManager {
	return &ScanManager{
		cancelFuncs: make(map[string]context.CancelFunc),
	}
}

// TryStartGlobalScan 尝试启动全局扫描
func (sm *ScanManager) TryStartGlobalScan(taskID string) (bool, context.CancelFunc) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.globalScanRunning {
		return false, nil // 全局扫描已在运行
	}

	// 中断所有并发扫描任务
	for id, cancel := range sm.cancelFuncs {
		if id != taskID {
			cancel()
			delete(sm.cancelFuncs, id)
		}
	}

	sm.globalScanRunning = true
	sm.concurrentScanCount = 0
	return true, func() {
		sm.mu.Lock()
		sm.globalScanRunning = false
		delete(sm.cancelFuncs, taskID)
		sm.mu.Unlock()
	}
}

// TryStartConcurrentScan 尝试启动并发扫描
func (sm *ScanManager) TryStartConcurrentScan(taskID string) (bool, context.CancelFunc) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.globalScanRunning {
		return false, nil // 全局扫描运行时禁止并发扫描
	}

	sm.concurrentScanCount++
	return true, func() {
		sm.mu.Lock()
		sm.concurrentScanCount--
		delete(sm.cancelFuncs, taskID)
		sm.mu.Unlock()
	}
}

// RegisterCancelFunc 注册取消函数
func (sm *ScanManager) RegisterCancelFunc(taskID string, cancel context.CancelFunc) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.cancelFuncs[taskID] = cancel
}

// AddTotalFiles 新增：安全获取总文件数
func (tp *taskProgress) AddTotalFiles(count int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	atomic.AddInt32(&tp.totalFiles, int32(count))
	tp.initialized = true
}

type FileUsecase struct {
	db          mongo.Database
	fileRepo    domain_file_entity.FileRepository
	folderRepo  domain_file_entity.FolderRepository
	detector    domain_file_entity.FileDetector
	folderType  map[domain_file_entity.FileTypeNo]struct{}
	targetMutex sync.RWMutex
	workerPool  chan struct{}
	scanTimeout time.Duration

	// 进度跟踪相关字段
	activeTasks      map[string]*taskProgress // 任务ID -> 进度跟踪器
	activeTasksMu    sync.RWMutex             // 保护 activeTasks
	scanManager      *ScanManager             // 新增扫描状态管理器
	scanningPaths    map[string]bool          // 新增：记录正在扫描的路径
	scanningPathsMu  sync.RWMutex             // 保护scanningPaths的互斥锁
	activeScanCount  int                      // 新增：当前活跃的扫描任务数量
	scanProgress     float32
	scanMutex        sync.RWMutex
	lastScanStart    time.Time
	scanStageWeights struct {
		traversal   float32
		statistics  float32
		refactoring float32
	}

	audioExtractor scene_audio_db_usecase.AudioMetadataExtractorTaglib
	artistRepo     scene_audio_db_interface.ArtistRepository
	albumRepo      scene_audio_db_interface.AlbumRepository
	mediaRepo      scene_audio_db_interface.MediaFileRepository
	tempRepo       scene_audio_db_interface.TempRepository
	mediaCueRepo   scene_audio_db_interface.MediaFileCueRepository
}

func NewFileUsecase(
	db mongo.Database,
	fileRepo domain_file_entity.FileRepository,
	folderRepo domain_file_entity.FolderRepository,
	detector domain_file_entity.FileDetector,
	timeoutMinutes int,

	artistRepo scene_audio_db_interface.ArtistRepository,
	albumRepo scene_audio_db_interface.AlbumRepository,
	mediaRepo scene_audio_db_interface.MediaFileRepository,
	tempRepo scene_audio_db_interface.TempRepository,
	mediaCueRepo scene_audio_db_interface.MediaFileCueRepository,
) *FileUsecase {
	workerCount := runtime.NumCPU() * 2
	if workerCount < 4 {
		workerCount = 4
	}

	return &FileUsecase{
		db:            db,
		fileRepo:      fileRepo,
		folderRepo:    folderRepo,
		detector:      detector,
		workerPool:    make(chan struct{}, workerCount),
		scanTimeout:   time.Duration(timeoutMinutes) * time.Minute,
		scanningPaths: make(map[string]bool),
		scanManager:   NewScanManager(),               // 初始化扫描管理器
		activeTasks:   make(map[string]*taskProgress), // 新增初始化

		artistRepo:   artistRepo,
		albumRepo:    albumRepo,
		mediaRepo:    mediaRepo,
		tempRepo:     tempRepo,
		mediaCueRepo: mediaCueRepo,
	}
}

func (uc *FileUsecase) GetScanProgress() (float32, time.Time, int, map[string]string) {
	uc.scanMutex.RLock()
	defer uc.scanMutex.RUnlock()

	if len(uc.activeTasks) == 0 {
		return 0, uc.lastScanStart, 0, nil
	}

	totalProgress := float32(0)
	taskCount := 0
	taskStatuses := make(map[string]string)

	uc.activeTasksMu.RLock()
	defer uc.activeTasksMu.RUnlock()

	for id, task := range uc.activeTasks {
		taskStatuses[id] = task.status

		if task.status == "preparing" {
			totalProgress += 0.01 // 准备中状态返回1%
			taskCount++
			continue
		}

		if !task.initialized {
			totalProgress += 0.01
			taskCount++
			continue
		}

		walked := atomic.LoadInt32(&task.walkedFiles)
		processed := atomic.LoadInt32(&task.processedFiles)
		total := atomic.LoadInt32(&task.totalFiles)

		if total == 0 {
			continue
		}

		traversalRatio := float32(walked) / float32(2*total)
		processingRatio := float32(processed) / float32(total)
		taskProgress := (traversalRatio + processingRatio) * 0.5
		totalProgress += taskProgress
		taskCount++
	}

	if taskCount == 0 {
		return 0, uc.lastScanStart, 0, taskStatuses
	}

	avgProgress := totalProgress / float32(taskCount)
	return avgProgress, uc.lastScanStart, taskCount, taskStatuses
}

func (uc *FileUsecase) ProcessDirectory(
	ctx context.Context,
	dirPaths []string,
	folderType int,
	ScanModel int,
) error {
	// 生成唯一任务ID
	taskID := fmt.Sprintf("%d-%v", ScanModel, time.Now().UnixNano())
	// 检查扫描模式并获取执行权限
	var (
		allowed bool
		cancel  context.CancelFunc
	)
	if ScanModel == 3 { // 全局扫描模式
		allowed, cancel = uc.scanManager.TryStartGlobalScan(taskID)
		if !allowed {
			return errors.New("全局扫描任务已在运行，无法启动新任务")
		}
	} else { // 并发扫描模式
		allowed, cancel = uc.scanManager.TryStartConcurrentScan(taskID)
		if !allowed {
			return errors.New("全局扫描任务运行中，无法启动并发扫描")
		}
	}
	// 注册取消函数
	uc.scanManager.RegisterCancelFunc(taskID, cancel)
	defer cancel() // 任务结束时自动取消注册
	// 创建可取消的上下文
	ctx, cancelTask := context.WithCancel(ctx)
	defer cancelTask()

	// 1. 检查路径是否已在扫描中
	uc.scanningPathsMu.Lock()
	for _, path := range dirPaths {
		if uc.scanningPaths[path] {
			uc.scanningPathsMu.Unlock()
			return fmt.Errorf("路径 %s 已在扫描中", path)
		}
	}

	// 2. 标记路径为扫描中
	for _, path := range dirPaths {
		uc.scanningPaths[path] = true
	}
	uc.scanningPathsMu.Unlock()

	// 3. 确保扫描完成后解除路径锁定
	defer func() {
		uc.scanningPathsMu.Lock()
		for _, path := range dirPaths {
			delete(uc.scanningPaths, path)
		}
		uc.scanningPathsMu.Unlock()
	}()

	// 重置进度计数器
	uc.scanMutex.Lock()
	uc.activeScanCount++
	if uc.activeScanCount == 1 {
		uc.scanProgress = 0
		uc.lastScanStart = time.Now()
	}
	// 设置各阶段权重（根据实际耗时比例）
	switch ScanModel {
	case 0:
		uc.scanStageWeights = struct {
			traversal   float32
			statistics  float32
			refactoring float32
		}{
			traversal:   0.9, // 文件处理阶段权重
			statistics:  0.1, // 统计阶段权重
			refactoring: 0,   // 重构阶段权重
		}
	case 1:
		uc.scanStageWeights = struct {
			traversal   float32
			statistics  float32
			refactoring float32
		}{
			traversal:   0.9,  // 文件处理阶段权重
			statistics:  0.05, // 统计阶段权重
			refactoring: 0.05, // 重构阶段权重
		}
	case 2:
		uc.scanStageWeights = struct {
			traversal   float32
			statistics  float32
			refactoring float32
		}{
			traversal:   0.6, // 文件处理阶段权重
			statistics:  0.1, // 统计阶段权重
			refactoring: 0.3, // 重构阶段权重
		}
	case 3:
		uc.scanStageWeights = struct {
			traversal   float32
			statistics  float32
			refactoring float32
		}{
			traversal:   0,   // 文件处理阶段权重
			statistics:  0.2, // 统计阶段权重
			refactoring: 0.8, // 重构阶段权重
		}
	}
	uc.scanMutex.Unlock()

	// 扫描前删除索引
	mongo.DropAllIndexes(uc.db)

	// 在函数返回时，减少activeScanCount
	defer func() {
		uc.scanMutex.Lock()
		uc.activeScanCount--
		uc.scanMutex.Unlock()
		// 扫描结束后创建索引
		mongo.CreateIndexes(uc.db)
	}()

	// 修复：在正确位置初始化任务进度跟踪器
	taskProg := &taskProgress{
		id:     taskID,
		status: "preparing", // 初始状态
	}

	// 注册任务
	uc.activeTasks[taskID] = taskProg
	// 开始统计文件数
	taskProg.status = "counting_files"
	uc.activeTasks[taskID] = taskProg

	// 任务结束时清理
	defer func() {
		uc.activeTasksMu.Lock()
		delete(uc.activeTasks, taskID)
		uc.activeTasksMu.Unlock()
	}()

	var libraryFolders []*domain_file_entity.LibraryFolderMetadata

	if ScanModel == 0 || ScanModel == 2 {
		// 处理多个目录路径
		for _, dirPath := range dirPaths {
			if len(dirPath) == 0 {
				continue
			}
			folder, err := uc.folderRepo.FindLibrary(ctx, dirPath, folderType)
			if err != nil {
				log.Printf("文件夹查询失败: %v", err)
				return fmt.Errorf("folder query failed: %w", err)
			}

			if folder == nil {
				newFolder := &domain_file_entity.LibraryFolderMetadata{
					ID:          primitive.NewObjectID(),
					FolderPath:  dirPath,
					FolderType:  folderType,
					FileCount:   0,
					LastScanned: time.Now(),
				}
				if err := uc.folderRepo.Insert(ctx, newFolder); err != nil {
					log.Printf("文件夹创建失败: %v", err)
					return fmt.Errorf("folder creation failed: %w", err)
				}
				libraryFolders = append(libraryFolders, newFolder)
			} else {
				libraryFolders = append(libraryFolders, folder)
			}
		}
		if libraryFolders == nil {
			// 默认扫描所有媒体库目录，当传入路径为空且ScanModel为2时
			library, err := uc.folderRepo.GetAllByType(ctx, 1)
			if err != nil {
				log.Printf("文件夹查询失败: %v", err)
				return fmt.Errorf("folder query failed: %w", err)
			}
			libraryFolders = library
		}
	} else if ScanModel == 1 || ScanModel == 3 {
		if ScanModel == 3 && len(dirPaths) > 0 {
			folder, err := uc.folderRepo.FindLibrary(ctx, dirPaths[0], folderType)
			if err != nil {
				log.Printf("文件夹查询失败: %v", err)
				return fmt.Errorf("folder query failed: %w", err)
			}
			if folder != nil {
				err = uc.folderRepo.Delete(ctx, folder.ID)
				if err != nil {
					return err
				}
			}
		}
		// 3. 获取整个媒体库（修复模式使用）
		library, err := uc.folderRepo.GetAllByType(ctx, 1)
		if err != nil {
			log.Printf("文件夹查询失败: %v", err)
			return fmt.Errorf("folder query failed: %w", err)
		}
		libraryFolders = library
	}

	if ScanModel == 3 {
		if len(dirPaths) > 0 {
			delResult, err := uc.mediaRepo.DeleteByFolder(ctx, dirPaths[0])
			if err != nil {
				log.Printf("常规音频清理失败: %v", err)
				return fmt.Errorf("regular audio cleanup failed: %w", err)
			} else if delResult > 0 {
				log.Printf("正在删除媒体库：%v, 其中删除%d个常规音频文件", dirPaths[0], delResult)
			}
			delCueResult, err := uc.mediaCueRepo.DeleteByFolder(ctx, dirPaths[0])
			if err != nil {
				log.Printf("CUE音频清理失败: %v", err)
				return fmt.Errorf("regular audio cleanup failed: %w", err)
			} else if delCueResult > 0 {
				log.Printf("正在删除媒体库：%v, 其中删除%d个常规音频文件", dirPaths[0], delCueResult)
			}
			delArtistResult, err := uc.artistRepo.DeleteAll(ctx)
			if err != nil {
				log.Printf("艺术家清理失败: %v", err)
				return fmt.Errorf("artist cleanup failed: %w", err)
			} else if delArtistResult > 0 {
				log.Printf("正在删除媒体库：%v, 其中删除%d个艺术家", dirPaths[0], delArtistResult)
			}
			delAlbumResult, err := uc.albumRepo.DeleteAll(ctx)
			if err != nil {
				log.Printf("专辑清理失败: %v", err)
				return fmt.Errorf("album cleanup failed: %w", err)
			} else if delAlbumResult > 0 {
				log.Printf("正在删除媒体库：%v, 其中删除%d个专辑", dirPaths[0], delAlbumResult)
			}
		}
	}

	if libraryFolders == nil {
		libraryFolders = make([]*domain_file_entity.LibraryFolderMetadata, 0)
	}

	// 4. 根据扫描模式执行处理
	switch ScanModel {
	case 0: // 扫描新的和有修改的文件
		if folderType == 1 {
			err := uc.ProcessMusicDirectory(ctx, libraryFolders, taskProg, true, true, false)
			if err != nil {
				return err
			}
		}
	case 1: // 搜索缺失的元数据
		if folderType == 1 {
			err := uc.ProcessMusicDirectory(ctx, libraryFolders, taskProg, true, false, false)
			if err != nil {
				return err
			}
		}
	case 2: // 覆盖所有元数据
		if folderType == 1 {
			err := uc.ProcessMusicDirectory(ctx, libraryFolders, taskProg, true, true, true)
			if err != nil {
				return err
			}
		}
	case 3: // 重构媒体库
		if folderType == 1 {
			err := uc.ProcessMusicDirectory(ctx, libraryFolders, taskProg, true, true, true)
			if err != nil {
				return err
			}
		}
	}

	log.Printf("媒体库扫描完成，共处理%d个目录", len(dirPaths))

	return nil
}

func (uc *FileUsecase) ProcessMusicDirectory(
	ctx context.Context,
	libraryFolders []*domain_file_entity.LibraryFolderMetadata,
	taskProg *taskProgress,
	libraryTraversal bool, // 是否遍历传入的媒体库目录
	libraryStatistics bool, // 是否统计传入的媒体库目录
	libraryRefactoring bool, // 是否重构传入的媒体库目录（覆盖所有元数据）
) error {
	var finalErr error

	// 更新文件夹统计与状态
	for _, folderInfo := range libraryFolders {
		if updateErr := uc.folderRepo.UpdateStats(
			ctx, folderInfo.ID, folderInfo.FileCount, domain_file_entity.StatusActive,
		); updateErr != nil {
			log.Printf("媒体库：%s (ID %s) - 统计更新失败: %v", folderInfo.FolderPath, folderInfo.ID, updateErr)
			return fmt.Errorf("stats update failed: %w", updateErr)
		}
	}

	// 修复：统计总文件数并设置到任务进度
	totalFiles := 0
	for _, folder := range libraryFolders {
		count, err := fastCountFilesInFolder(folder.FolderPath)
		if err != nil {
			log.Printf("统计文件数失败: %v", err)
			continue
		}
		totalFiles += count
	}
	// 更新任务状态
	taskProg.status = "processing"
	taskProg.AddTotalFiles(totalFiles)

	coverTempPath, _ := uc.tempRepo.GetTempPath(ctx, "cover")

	var libraryFolderNewInfos []struct {
		libraryFolderID        primitive.ObjectID
		libraryFolderPath      string
		libraryFolderFileCount int
	}
	var regularAudioPaths []string
	var cueAudioPaths []string

	if libraryStatistics {
		// 扫描前重置数据库统计字段
		_, err := uc.albumRepo.ResetALLField(ctx)
		if err != nil {
			return err
		}
		_, err = uc.albumRepo.ResetField(ctx, "AllArtistIDs")
		if err != nil {
			return err
		}
		_, err = uc.artistRepo.ResetALLField(ctx)
		if err != nil {
			return err
		}
	}

	// 路径收集容器
	regularAudioPaths = make([]string, 0) // 常规音频路径
	cueAudioPaths = make([]string, 0)     // CUE音频路径
	// 路径去重容器
	addedCuePaths := make(map[string]bool)
	addedRegularPaths := make(map[string]bool)

	for _, folder := range libraryFolders {
		libraryFolderPath := strings.Replace(folder.FolderPath, "/", "\\", -1)
		if !strings.HasSuffix(libraryFolderPath, "\\") {
			libraryFolderPath += "\\"
		}

		folderInfo := struct {
			libraryFolderID        primitive.ObjectID
			libraryFolderPath      string
			libraryFolderFileCount int
		}{
			libraryFolderID:        folder.ID,
			libraryFolderPath:      libraryFolderPath,
			libraryFolderFileCount: 0,
		}

		// 并发处理管道
		var wgFile sync.WaitGroup
		errChan := make(chan error, 100)

		// 存储需要排除的.wav文件路径
		excludeWavs := make(map[string]struct{})
		cueResourcesMap := make(map[string]*scene_audio_db_models.CueConfig)

		// 遍历时统计总文件数
		for _, folderCount := range libraryFolders {
			total := 0
			err := filepath.Walk(folderCount.FolderPath, func(path string, info os.FileInfo, err error) error {
				select {
				case <-ctx.Done(): // 响应取消
					return ctx.Err()
				default:
				}
				if !info.IsDir() && uc.shouldProcess(path, 1) {
					total++
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		// 第一次遍历：收集.cue文件和关联资源
		err := filepath.Walk(folder.FolderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("访问路径 %s 出错: %v", path, err)
				return nil // 跳过错误继续遍历
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if info.IsDir() {
					return nil
				}

				// 修复：更新任务级别的遍历计数器
				atomic.AddInt32(&taskProg.walkedFiles, 1)

				ext := strings.ToLower(filepath.Ext(path))
				dir := filepath.Dir(path)
				baseName := strings.TrimSuffix(filepath.Base(path), ext)

				// 处理.cue文件
				if ext == ".cue" {
					res := &scene_audio_db_models.CueConfig{CuePath: path}

					// 查找同名音频文件
					audioFound := false
					audioExts := []string{".wav", ".ape", ".flac"} // 扩展支持的音频格式
					for _, audioExt := range audioExts {
						audioPath := filepath.Join(dir, baseName+audioExt)
						if _, err := os.Stat(audioPath); err == nil {
							res.AudioPath = audioPath
							excludeWavs[audioPath] = struct{}{} // 添加到排除列表
							audioFound = true
							break // 找到一种音频格式即停止
						}
					}

					// 未找到音频文件则跳过
					if !audioFound {
						return nil
					}

					// 收集相关资源文件[9](@ref)
					resourceFiles := []string{"back.jpg", "cover.jpg", "disc.jpg", "list.txt", "log.txt"}
					for _, f := range resourceFiles {
						p := filepath.Join(dir, f)
						if _, err := os.Stat(p); err == nil {
							switch f {
							case "back.jpg":
								res.BackImage = p
							case "cover.jpg":
								res.CoverImage = p
							case "disc.jpg":
								res.DiscImage = p
							case "list.txt":
								res.ListFile = p
							case "log.txt":
								res.LogFile = p
							}
						}
					}

					cueResourcesMap[path] = res
				}
				return nil
			}
		})
		if err != nil {
			log.Printf("文件夹%v，文件遍历错误: %v，请检查该文件夹是否存在", folder.FolderPath, err)
		}

		// 第二次遍历：处理文件
		err = filepath.Walk(folder.FolderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {

				log.Printf("访问路径 %s 出错: %v", path, err)
				return nil // 跳过错误继续遍历
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if info.IsDir() || !uc.shouldProcess(path, 1) {
					return nil
				}

				// 修复：更新任务级别的遍历计数器
				atomic.AddInt32(&taskProg.walkedFiles, 1)

				// 检查是否是被排除的.wav文件
				if _, excluded := excludeWavs[path]; excluded {
					return nil
				}

				ext := strings.ToLower(filepath.Ext(path))

				// 处理.cue文件
				if ext == ".cue" {
					res, exists := cueResourcesMap[path]
					if !exists {
						log.Printf("未找到cue文件对应的音频资源: %s", path)
						return nil
					}

					// 关键修复：路径收集不受libraryTraversal影响
					if res.AudioPath != "" {
						if _, existsAdded := addedCuePaths[res.AudioPath]; !existsAdded {
							cueAudioPaths = append(cueAudioPaths, res.AudioPath)
							addedCuePaths[res.AudioPath] = true
						}
					}

					// 仅当开启遍历时才执行文件处理和计数
					if libraryTraversal {
						wgFile.Add(1)
						folderInfo.libraryFolderFileCount++
						go uc.processFile(ctx, res, res.AudioPath, libraryFolderPath, coverTempPath, folder.ID, &wgFile, errChan, taskProg)
					}
				} else {
					// 关键修复：路径收集不受libraryTraversal影响
					if _, existsAdded := addedRegularPaths[path]; !existsAdded {
						regularAudioPaths = append(regularAudioPaths, path)
						addedRegularPaths[path] = true
					}

					// 仅当开启遍历时才执行文件处理和计数
					if libraryTraversal {
						wgFile.Add(1)
						folderInfo.libraryFolderFileCount++
						go uc.processFile(ctx, nil, path, libraryFolderPath, coverTempPath, folder.ID, &wgFile, errChan, taskProg)
					}
					return nil
				}
				return nil
			}
		})
		if err != nil {
			log.Printf("文件夹%v，文件遍历错误: %v，请检查该文件夹是否存在", folder.FolderPath, err)
		}

		// 等待所有任务完成
		go func() {
			wgFile.Wait()
			close(errChan)
		}()

		// 收集错误
		for err := range errChan {
			log.Printf("文件处理错误: %v", err)
			if finalErr == nil {
				finalErr = err
			} else {
				finalErr = fmt.Errorf("%v; %w", finalErr, err)
			}
		}

		libraryFolderNewInfos = append(libraryFolderNewInfos, folderInfo)
	}

	artistIDs, err := uc.artistRepo.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("获取艺术家ID列表失败: %w", err)
	}

	// 区域2: 媒体库统计（仅当libraryStatistics为true时执行）
	if libraryStatistics {
		maxConcurrency := 50
		sem := make(chan struct{}, maxConcurrency)
		var wgUpdate sync.WaitGroup
		counters := []struct {
			countMethod func(context.Context, string) (int64, error)
			counterName string
			countType   string
		}{
			{
				countMethod: uc.albumRepo.AlbumCountByArtist,
				counterName: "album_count",
				countType:   "专辑",
			},
			{
				countMethod: uc.albumRepo.GuestAlbumCountByArtist,
				counterName: "guest_album_count",
				countType:   "合作专辑",
			},
			{
				countMethod: uc.mediaRepo.MediaCountByArtist,
				counterName: "song_count",
				countType:   "单曲",
			},
			{
				countMethod: uc.mediaRepo.GuestMediaCountByArtist,
				counterName: "guest_song_count",
				countType:   "合作单曲",
			},
			{
				countMethod: uc.mediaCueRepo.MediaCueCountByArtist,
				counterName: "cue_count",
				countType:   "光盘",
			},
			{
				countMethod: uc.mediaCueRepo.GuestMediaCueCountByArtist,
				counterName: "guest_cue_count",
				countType:   "合作光盘",
			},
		}

		// 更新当前阶段为统计阶段
		uc.scanMutex.Lock()
		uc.scanProgress = uc.scanStageWeights.traversal // 文件处理已完成
		uc.scanMutex.Unlock()

		totalArtists := len(artistIDs)
		current := 0

		for _, artistID := range artistIDs {
			wgUpdate.Add(1)
			go func(id primitive.ObjectID) {
				defer wgUpdate.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				strID := id.Hex()
				ctx := context.Background()

				// 统一处理计数器更新
				for _, counter := range counters {
					count, err := counter.countMethod(ctx, strID)
					if err != nil {
						log.Printf("艺术家%s%s统计失败: %v", strID, counter.countType, err)
						continue
					}

					if _, err = uc.artistRepo.UpdateCounter(ctx, id, counter.counterName, int(count)); err != nil {
						log.Printf("艺术家%s%s计数更新失败: %v", strID, counter.countType, err)
					}
				}

				// 更新统计阶段进度
				current++
				uc.scanMutex.Lock()
				traversalProgress := uc.scanStageWeights.traversal
				statisticsProgress := uc.scanStageWeights.statistics * float32(current) / float32(totalArtists)
				uc.scanProgress = traversalProgress + statisticsProgress
				uc.scanMutex.Unlock()
			}(artistID)
		}
		wgUpdate.Wait()
	}

	// 区域3: 媒体库重构（仅当libraryRefactoring为true时执行）
	if libraryRefactoring {
		// 更新当前阶段为重构阶段
		uc.scanMutex.Lock()
		uc.scanProgress = uc.scanStageWeights.traversal + uc.scanStageWeights.statistics
		uc.scanMutex.Unlock()

		totalArtists := len(artistIDs)
		current := 0

		// 定义统一的结构体类型
		type artistStats struct {
			mediaCount         int
			guestMediaCount    int
			cueMediaCount      int
			guestCueMediaCount int
			albumCount         int
			guestAlbumCount    int
		}

		// 使用指针类型map
		invalidArtistCounts := make(map[string]*artistStats)
		deleteArtistCounts := make(map[string]int)
		invalidAlbumCounts := make(map[string]int)

		artistAlbums, err := uc.albumRepo.GetArtistAlbumsMap(ctx)
		if err != nil {
			return fmt.Errorf("获取专辑ID列表失败: %w", err)
		}

		artistGuestAlbums, err := uc.albumRepo.GetArtistGuestAlbumsMap(ctx)
		if err != nil {
			return fmt.Errorf("获取专辑ID列表失败: %w", err)
		}

		// 1. 处理常规音频的艺术家统计
		for _, id := range artistIDs {
			artistID := id.Hex()

			// 确保结构体已初始化[1,2](@ref)
			if _, exists := invalidArtistCounts[artistID]; !exists {
				invalidArtistCounts[artistID] = &artistStats{}
			}

			// 处理主艺术家统计
			mediaCount, err := uc.mediaRepo.InspectMediaCountByArtist(ctx, artistID, regularAudioPaths)
			if err != nil {
				log.Printf("常规音频主艺术家统计失败 (ID %s): %v", artistID, err)
			} else {
				if mediaCount >= 0 {
					invalidArtistCounts[artistID].mediaCount += mediaCount
				} else {
					invalidArtistCounts[artistID].mediaCount++
					deleteArtistCounts[artistID]++
				}
			}

			// 处理合作艺术家统计
			guestMediaCount, err := uc.mediaRepo.InspectGuestMediaCountByArtist(ctx, artistID, regularAudioPaths)
			if err != nil {
				log.Printf("常规音频合作艺术家统计失败 (ID %s): %v", artistID, err)
			} else {
				if guestMediaCount >= 0 {
					invalidArtistCounts[artistID].guestMediaCount += guestMediaCount
				} else {
					invalidArtistCounts[artistID].guestMediaCount++
					deleteArtistCounts[artistID]++
				}
			}

			// 更新重构阶段进度
			current++
			uc.scanMutex.Lock()
			progress := uc.scanStageWeights.traversal +
				uc.scanStageWeights.statistics +
				uc.scanStageWeights.refactoring*float32(current)/float32(totalArtists)
			uc.scanProgress = progress
			uc.scanMutex.Unlock()
		}

		// 2. 处理常规音频的专辑统计
		for artistID, albums := range artistAlbums {
			artistIDStr := artistID.Hex()

			// 确保结构体已初始化[1,2](@ref)
			if _, exists := invalidArtistCounts[artistIDStr]; !exists {
				invalidArtistCounts[artistIDStr] = &artistStats{}
			}

			for _, albumID := range albums {
				albumIDStr := albumID.Hex()
				count, err := uc.mediaRepo.InspectMediaCountByAlbum(ctx, albumIDStr, regularAudioPaths)
				if err != nil {
					log.Printf("常规音频专辑统计失败 (ID %s): %v", albumIDStr, err)
					continue
				}

				if count > 0 {
					// 直接修改指针指向的结构体字段[7](@ref)
					invalidArtistCounts[artistIDStr].albumCount += count
					//
					invalidAlbumCounts[albumIDStr] += count
				} else if count < 0 {
					// 删除艺术家统计
					invalidArtistCounts[artistIDStr].albumCount++
					// 标记为删除
					invalidAlbumCounts[albumIDStr] = -1
					//
					deleteArtistCounts[artistIDStr]++
				} else {
					delete(invalidAlbumCounts, albumIDStr)
				}
			}
		}

		// 2.1 处理常规音频的合作专辑统计
		for artistID, albums := range artistGuestAlbums {
			artistIDStr := artistID.Hex()

			// 确保结构体已初始化[1,2](@ref)
			if _, exists := invalidArtistCounts[artistIDStr]; !exists {
				invalidArtistCounts[artistIDStr] = &artistStats{}
			}

			for _, albumID := range albums {
				albumIDStr := albumID.Hex()
				count, err := uc.mediaRepo.InspectGuestMediaCountByAlbum(ctx, albumIDStr, regularAudioPaths)
				if err != nil {
					log.Printf("常规音频专辑统计失败 (ID %s): %v", albumIDStr, err)
					continue
				}

				if count > 0 {
					invalidArtistCounts[artistIDStr].guestAlbumCount += count
				} else if count < 0 {
					invalidArtistCounts[artistIDStr].guestAlbumCount++
					deleteArtistCounts[artistIDStr]++
				}
			}
		}

		// 3. 清除常规音频的无效专辑（删除统计：单曲数为0的专辑）
		invalidAlbumNums := 0
		for albumID, operand := range invalidAlbumCounts {
			deleted, err := uc.albumRepo.InspectAlbumMediaCountByAlbum(ctx, albumID, operand)
			if err != nil {
				log.Printf("专辑清理失败 (ID %s): %v", albumID, err)
			}
			if deleted {
				invalidAlbumNums++
			}
		}
		log.Printf("已删除 %v 个无效专辑\n", invalidAlbumNums)

		// 4. 处理CUE音频的艺术家统计
		for _, id := range artistIDs {
			artistID := id.Hex()

			// 确保结构体已初始化[1,2](@ref)
			if _, exists := invalidArtistCounts[artistID]; !exists {
				invalidArtistCounts[artistID] = &artistStats{}
			}

			// 处理CUE音频主艺术家统计
			cueMediaCount, err := uc.mediaCueRepo.InspectMediaCueCountByArtist(ctx, artistID, cueAudioPaths)
			if err != nil {
				log.Printf("CUE音频主艺术家统计失败 (ID %s): %v", artistID, err)
			} else {
				if cueMediaCount >= 0 {
					invalidArtistCounts[artistID].cueMediaCount += cueMediaCount
				} else {
					invalidArtistCounts[artistID].cueMediaCount++
					deleteArtistCounts[artistID]++
				}
			}

			// 处理CUE音频合作艺术家统计
			guestCueMediaCount, err := uc.mediaCueRepo.InspectGuestMediaCueCountByArtist(ctx, artistID, cueAudioPaths)
			if err != nil {
				log.Printf("CUE音频合作艺术家统计失败 (ID %s): %v", artistID, err)
			} else {
				if guestCueMediaCount >= 0 {
					invalidArtistCounts[artistID].guestCueMediaCount += guestCueMediaCount
				} else {
					invalidArtistCounts[artistID].guestCueMediaCount++
					deleteArtistCounts[artistID]++
				}
			}

			// 更新重构阶段进度
			current++
			uc.scanMutex.Lock()
			progress := uc.scanStageWeights.traversal +
				uc.scanStageWeights.statistics +
				uc.scanStageWeights.refactoring*float32(current)/float32(totalArtists)
			uc.scanProgress = progress
			uc.scanMutex.Unlock()
		}

		// 5. 清除无效的艺术家（删除统计：【单曲/合作单曲+CD/合作CD都为0+专辑/合作专辑都为0】的艺术家）
		artistIDStrCount := 0
		for _, artistID := range artistIDs {
			artistIDStr := artistID.Hex()
			if deleteCount, exists := deleteArtistCounts[artistIDStr]; exists && deleteCount >= 6 {
				err := uc.artistRepo.DeleteByID(ctx, artistID)
				if err != nil {
					log.Printf("艺术家删除失败 (ID %s): %v", artistIDStr, err)
				} else {
					artistIDStrCount++
				}
				delete(invalidArtistCounts, artistIDStr)
			}

			// 更新重构阶段进度
			current++
			uc.scanMutex.Lock()
			progress := uc.scanStageWeights.traversal +
				uc.scanStageWeights.statistics +
				uc.scanStageWeights.refactoring*float32(current)/float32(totalArtists)
			uc.scanProgress = progress
			uc.scanMutex.Unlock()
		}
		log.Printf("已删除 %d 个无效艺术家", artistIDStrCount)

		// 6. 清除无效的常规音频
		delResult, invalidMediaArtist, err := uc.mediaRepo.DeleteAllInvalid(ctx, regularAudioPaths)
		if err != nil {
			log.Printf("常规音频清理失败: %v", err)
			return fmt.Errorf("regular audio cleanup failed: %w", err)
		} else if delResult > 0 {
			for _, artist := range invalidMediaArtist {
				artistID := artist.ArtistID.Hex()
				// 确保结构体已初始化
				if _, exists := invalidArtistCounts[artistID]; !exists {
					invalidArtistCounts[artistID] = &artistStats{}
				}
				invalidArtistCounts[artistID].mediaCount = int(artist.Count)
			}
		}

		// 7. 清除无效的CUE音频
		delCueResult, invalidMediaCueArtist, err := uc.mediaCueRepo.DeleteAllInvalid(ctx, cueAudioPaths)
		if err != nil {
			log.Printf("CUE音频清理失败: %v", err)
			return fmt.Errorf("CUE audio cleanup failed: %w", err)
		} else if delCueResult > 0 {
			for _, artist := range invalidMediaCueArtist {
				artistID := artist.ArtistID.Hex()
				// 确保结构体已初始化
				if _, exists := invalidArtistCounts[artistID]; !exists {
					invalidArtistCounts[artistID] = &artistStats{}
				}
				invalidArtistCounts[artistID].cueMediaCount = int(artist.Count)
			}
		}

		// 8. 重构媒体库艺术家操作数
		artistOperands := make(map[string]struct {
			albumCount         int
			guestAlbumCount    int
			mediaCount         int
			guestMediaCount    int
			cueMediaCount      int
			cueGuestMediaCount int
		})

		// 收集常规、CUE音频艺术家操作数
		for artistID, stats := range invalidArtistCounts {
			op := artistOperands[artistID]
			op.albumCount = stats.albumCount
			op.guestAlbumCount = stats.guestAlbumCount
			op.mediaCount = stats.mediaCount
			op.guestMediaCount = stats.guestMediaCount
			op.cueMediaCount = stats.cueMediaCount
			op.cueGuestMediaCount = stats.guestCueMediaCount
			artistOperands[artistID] = op
		}

		// 应用艺术家操作数。不能为0，否则将删除艺术家
		for artistID, operands := range artistOperands {
			// 专辑计数处理
			if operands.albumCount > 0 {
				operands.albumCount, err = uc.artistRepo.InspectAlbumCountByArtist(ctx, artistID, operands.albumCount)
				if err != nil {
					log.Printf("艺术家专辑计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 合作专辑计数处理
			if operands.guestAlbumCount > 0 {
				operands.guestAlbumCount, err = uc.artistRepo.InspectGuestAlbumCountByArtist(ctx, artistID, operands.guestAlbumCount)
				if err != nil {
					log.Printf("艺术家合作专辑计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 单曲计数处理
			if operands.mediaCount > 0 {
				operands.mediaCount, err = uc.artistRepo.InspectMediaCountByArtist(ctx, artistID, operands.mediaCount)
				if err != nil {
					log.Printf("艺术家单曲计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 合作单曲计数处理
			if operands.guestMediaCount > 0 {
				operands.guestMediaCount, err = uc.artistRepo.InspectGuestMediaCountByArtist(ctx, artistID, operands.guestMediaCount)
				if err != nil {
					log.Printf("艺术家合作单曲计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 光盘计数处理
			if operands.cueMediaCount > 0 {
				operands.cueMediaCount, err = uc.artistRepo.InspectMediaCueCountByArtist(ctx, artistID, operands.cueMediaCount)
				if err != nil {
					log.Printf("艺术家光盘计数处理失败 (ID %s): %v", artistID, err)
				}
			}

			// 合作光盘计数处理
			if operands.cueGuestMediaCount > 0 {
				operands.cueGuestMediaCount, err = uc.artistRepo.InspectGuestMediaCueCountByArtist(ctx, artistID, operands.cueGuestMediaCount)
				if err != nil {
					log.Printf("艺术家合作光盘计数处理失败 (ID %s): %v", artistID, err)
				}
			}
		}

		// 清除无效的艺术家：应用艺术家操作数之后
		invalid, err := uc.artistRepo.DeleteAllInvalid(ctx)
		if err != nil {
			return err
		}
		if invalid > 0 {
			log.Printf("已删除 %d 个无效艺术家", invalid)
		}
	}

	// 更新文件夹统计（仅在执行遍历时更新）
	if libraryTraversal {
		for _, folderInfo := range libraryFolderNewInfos {
			if updateErr := uc.folderRepo.UpdateStats(
				ctx, folderInfo.libraryFolderID, folderInfo.libraryFolderFileCount, domain_file_entity.StatusActive,
			); updateErr != nil {
				log.Printf("媒体库：%s (ID %s) - 统计更新失败: %v", folderInfo.libraryFolderPath, folderInfo.libraryFolderID, updateErr)
				return fmt.Errorf("stats update failed: %w", updateErr)
			}
		}
	}

	return finalErr
}

// 使用快速目录统计函数
func fastCountFilesInFolder(rootPath string) (int, error) {
	count := 0
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

func (uc *FileUsecase) shouldProcess(path string, folderType int) bool {
	fileType, err := uc.detector.DetectMediaType(path)
	if err != nil {
		log.Printf("文件类型检测失败: %v", err)
		return false
	}
	if folderType == 1 && fileType == domain_file_entity.Audio {
		return true
	}
	if folderType == 2 && fileType == domain_file_entity.Video {
		return true
	}
	if folderType == 3 && fileType == domain_file_entity.Image {
		return true
	}
	if folderType == 4 && fileType == domain_file_entity.Document {
		return true
	}
	return false
}

func (uc *FileUsecase) processFile(
	ctx context.Context,
	res *scene_audio_db_models.CueConfig,
	path string,
	libraryPath string,
	coverTempPath string,
	libraryFolderID primitive.ObjectID,
	wg *sync.WaitGroup,
	errChan chan<- error,
	taskProg *taskProgress,
) {
	defer func() {
		// 修复：更新任务级别的处理计数器
		atomic.AddInt32(&taskProg.processedFiles, 1)
		wg.Done()
	}()

	// 上下文取消检查
	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return
	default:
	}

	// 获取工作槽
	select {
	case uc.workerPool <- struct{}{}:
		defer func() { <-uc.workerPool }()
	case <-ctx.Done():
		errChan <- ctx.Err()
		return
	}

	// 文件类型检测
	fileType, err := uc.detector.DetectMediaType(path)
	if err != nil {
		log.Printf("文件检测失败: %s | %v", path, err)
		errChan <- fmt.Errorf("文件检测失败 %s: %w", path, err)
		return
	}

	// 创建基础元数据
	metadata, err := uc.createMetadataBasicInfo(path, libraryFolderID)
	if err != nil {
		log.Printf("元数据创建失败: %s | %v", path, err)
		errChan <- fmt.Errorf("文件处理失败 %s: %w", path, err)
		return
	}

	// 保存基础文件信息
	if err := uc.fileRepo.Upsert(ctx, metadata); err != nil {
		log.Printf("文件写入失败: %s | %v", path, err)
		errChan <- fmt.Errorf("数据库写入失败 %s: %w", path, err)
		return
	}

	// 处理音频文件
	if fileType == domain_file_entity.Audio {
		mediaFile, album, artists, mediaFileCue, err := uc.audioExtractor.Extract(path, libraryPath, metadata, res)
		if err != nil {
			return
		}

		if mediaFile != nil && mediaFile.Title == "" {
			mediaFile.Title = "Unknown Title"
			mediaFile.OrderTitle = "Unknown Title"
			mediaFile.SortTitle = "Unknown Title"
		}
		if mediaFileCue != nil && mediaFileCue.Title == "" {
			mediaFileCue.Title = "Unknown Title"
		}

		if err := uc.processAudioHierarchy(ctx, artists, album, mediaFile, mediaFileCue); err != nil {
			return
		}

		if err := uc.processAudioMediaFilesAndAlbumCover(
			ctx,
			mediaFile,
			album,
			mediaFileCue,
			path,
			coverTempPath,
		); err != nil {
			errChan <- fmt.Errorf("文件存储失败 %s: %w", path, err)
			return
		}
	}
}

func (uc *FileUsecase) createMetadataBasicInfo(
	path string,
	libraryFolderID primitive.ObjectID,
) (*domain_file_entity.FileMetadata, error) {
	// 1. 先查询是否已存在该路径文件
	existingFile, err := uc.fileRepo.FindByPath(context.Background(), path)
	if err != nil && !errors.Is(err, driver.ErrNoDocuments) {
		log.Printf("路径查询失败: %s | %v", path, err)
		return nil, fmt.Errorf("路径查询失败: %w", err)
	}

	// 2. 已存在则直接返回
	if existingFile != nil {
		return existingFile, nil
	}

	// 3. 不存在时执行原流程
	file, err := os.Open(path)
	if err != nil {
		log.Printf("文件打开失败: %s | %v", path, err)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("文件关闭失败: %s | %v", path, err)
		}
	}(file)

	stat, err := file.Stat()
	if err != nil {
		log.Printf("文件信息获取失败: %s | %v", path, err)
		return nil, err
	}

	fileType, err := uc.detector.DetectMediaType(path)
	if err != nil {
		log.Printf("文件类型检测失败: %s | %v", path, err)
		return nil, err
	}

	hash := sha256.New()
	if _, err := io.CopyBuffer(hash, file, make([]byte, 32*1024)); err != nil {
		log.Printf("哈希计算失败: %s | %v", path, err)
		return nil, err
	}

	normalizedPath := filepath.ToSlash(filepath.Clean(path))

	return &domain_file_entity.FileMetadata{
		ID:        primitive.NewObjectID(),
		FolderID:  libraryFolderID,
		FilePath:  normalizedPath,
		FileType:  fileType,
		Size:      stat.Size(),
		ModTime:   stat.ModTime(),
		Checksum:  fmt.Sprintf("%x", hash.Sum(nil)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (uc *FileUsecase) processAudioMediaFilesAndAlbumCover(
	ctx context.Context,
	media *scene_audio_db_models.MediaFileMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	mediaCue *scene_audio_db_models.MediaFileCueMetadata,
	path string,
	coverBasePath string,
) error {
	if mediaCue != nil {
		// 处理CUE文件的封面逻辑（保持不变）
		mediaCueUpdate := bson.M{
			"$set": bson.M{
				"back_image_url":  mediaCue.CueResources.BackImage,
				"cover_image_url": mediaCue.CueResources.CoverImage,
				"disc_image_url":  mediaCue.CueResources.DiscImage,
				"has_cover_art":   mediaCue.CueResources.CoverImage != "",
			},
		}
		if _, err := uc.mediaCueRepo.UpdateByID(ctx, mediaCue.ID, mediaCueUpdate); err != nil {
			return fmt.Errorf("媒体更新失败: %w", err)
		}
	} else {
		// 创建封面存储目录
		mediaCoverDir := filepath.Join(coverBasePath, "media", media.ID.Hex())
		if err := os.MkdirAll(mediaCoverDir, 0755); err != nil {
			return fmt.Errorf("媒体目录创建失败: %w", err)
		}

		var coverPath string
		defer func() {
			if coverPath == "" {
				if err := os.RemoveAll(mediaCoverDir); err != nil {
					log.Printf("[WARN] 媒体目录删除失败 | 路径:%s | 错误:%v", mediaCoverDir, err)
				}
			}
		}()

		// 1. 优先检查本地已存在的封面文件[2,3](@ref)
		coverPath = uc.findLocalCover(path, mediaCoverDir)

		// 2. 如果未找到本地封面，尝试读取内嵌封面
		if coverPath == "" {
			file, err := os.Open(path)
			if err != nil {
				log.Printf("[WARN] 文件打开失败: %v", err)
			} else {
				defer func(file *os.File) {
					err := file.Close()
					if err != nil {
						return
					}
				}(file)
				if metadata, err := tag.ReadFrom(file); err == nil {
					coverPath = uc.extractEmbeddedCover(metadata, mediaCoverDir)
				}
			}
		}

		// 3. 如果仍未获取到封面，使用FFmpeg提取内嵌封面[1,3,6](@ref)
		if coverPath == "" {
			coverPath = uc.extractCoverWithFFmpeg(path, mediaCoverDir)
		}

		// 更新媒体封面信息
		mediaUpdate := bson.M{
			"$set": bson.M{
				"medium_image_url": coverPath,
				"has_cover_art":    coverPath != "",
			},
		}
		if _, err := uc.mediaRepo.UpdateByID(ctx, media.ID, mediaUpdate); err != nil {
			return fmt.Errorf("媒体更新失败: %w", err)
		}

		// 4. 处理专辑封面（如果存在）
		if album != nil && coverPath != "" {
			uc.processAlbumCover(album, coverPath, coverBasePath, ctx)
		}
	}

	return nil
}

// 在音频文件同目录查找封面文件[2](@ref)
func (uc *FileUsecase) findLocalCover(audioPath, targetDir string) string {
	dir := filepath.Dir(audioPath)
	baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))

	// 可能的封面文件名模式
	coverPatterns := []string{
		"cover.jpg", "cover.png", "cover.jpeg", // 通用封面名
		baseName + ".jpg", baseName + ".png", baseName + ".jpeg", // 与音频同名的封面
		"folder.jpg", "folder.png", // Windows常见的封面名
	}

	for _, pattern := range coverPatterns {
		srcPath := filepath.Join(dir, pattern)
		if _, err := os.Stat(srcPath); err == nil {
			destPath := filepath.Join(targetDir, "cover"+filepath.Ext(srcPath))
			if err := uc.copyCoverFile(srcPath, destPath); err == nil {
				return destPath
			}
		}
	}
	return ""
}

// 复制封面文件
func (uc *FileUsecase) copyCoverFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			return
		}
	}(srcFile)

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {
			return
		}
	}(destFile)

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}
	return nil
}

// 从标签元数据提取内嵌封面[1](@ref)
func (uc *FileUsecase) extractEmbeddedCover(metadata tag.Metadata, mediaCoverDir string) string {
	if pic := metadata.Picture(); pic != nil && len(pic.Data) > 0 {
		coverPath := filepath.Join(mediaCoverDir, "cover.jpg")
		if err := os.WriteFile(coverPath, pic.Data, 0644); err == nil {
			return coverPath
		}
	}
	return ""
}

// 使用FFmpeg提取封面（当其他方法失败时）[3,6](@ref)
func (uc *FileUsecase) extractCoverWithFFmpeg(audioPath, mediaCoverDir string) string {
	// 1. 设置输出路径为PNG格式（WAV封面多为PNG）
	coverPath := filepath.Join(mediaCoverDir, "cover.png")

	// 2. 构建优化的FFmpeg命令
	cmd := ffmpeggo.Input(audioPath).
		Output(coverPath, ffmpeggo.KwArgs{
			"an":     "",     // 丢弃音频流
			"vcodec": "copy", // 直接复制视频流（封面数据）
			"y":      "",     // 覆盖现有文件
		}).
		OverWriteOutput()

	// 3. 执行并记录详细命令
	compiledCmd := cmd.Compile()
	log.Printf("执行FFmpeg封面提取命令: %s", strings.Join(compiledCmd.Args, " "))

	// 4. 运行命令并处理错误
	if err := cmd.Run(); err != nil {
		// 5. 特定错误处理
		if strings.Contains(err.Error(), "No video stream") {
			log.Printf("[WARN] 文件无内嵌封面: %s", audioPath)
		} else {
			log.Printf("[ERROR] FFmpeg封面提取失败: %v", err)
		}
		return ""
	}

	// 6. 验证结果
	if info, err := os.Stat(coverPath); err != nil || info.Size() == 0 {
		log.Printf("[WARN] 封面生成失败 | 错误:%v", err)
		return ""
	}

	return coverPath
}

// 处理专辑封面[1](@ref)
func (uc *FileUsecase) processAlbumCover(album *scene_audio_db_models.AlbumMetadata,
	coverPath, coverBasePath string, ctx context.Context) {

	if album.MediumImageURL == "" {
		albumCoverDir := filepath.Join(coverBasePath, "album", album.ID.Hex())
		if err := os.MkdirAll(albumCoverDir, 0755); err != nil {
			log.Printf("[WARN] 专辑封面目录创建失败: %v", err)
			return
		}

		albumCoverPath := filepath.Join(albumCoverDir, "cover.jpg")
		if err := uc.copyCoverFile(coverPath, albumCoverPath); err != nil {
			log.Printf("[WARN] 专辑封面复制失败: %v", err)
			return
		}

		albumUpdate := bson.M{
			"$set": bson.M{
				"medium_image_url": albumCoverPath,
				"has_cover_art":    true,
				"updated_at":       time.Now().UTC(),
			},
		}

		if _, err := uc.albumRepo.UpdateByID(ctx, album.ID, albumUpdate); err != nil {
			log.Printf("[WARN] 专辑元数据更新失败: %v", err)
		}
	}
}

func (uc *FileUsecase) processAudioHierarchy(ctx context.Context,
	artists []*scene_audio_db_models.ArtistMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	mediaFile *scene_audio_db_models.MediaFileMetadata,
	mediaFileCue *scene_audio_db_models.MediaFileCueMetadata,
) (err error) { // 改为命名返回值，便于错误处理
	// 关键依赖检查
	if uc.mediaRepo == nil || uc.artistRepo == nil || uc.albumRepo == nil || uc.mediaCueRepo == nil {
		log.Print("音频仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	// 修复：检查mediaFile和mediaFileCue同时为nil的情况
	if mediaFile == nil && mediaFileCue == nil {
		log.Print("媒体文件元数据为空")
		return fmt.Errorf("mediaFile and mediaFileCue cannot both be nil")
	}

	// 直接保存无关联数据
	if artists == nil && album == nil {
		if mediaFile != nil {
			// 修复：移除短声明避免覆盖原变量
			if _, err := uc.mediaRepo.Upsert(ctx, mediaFile); err != nil {
				// 修复：直接使用原mediaFile避免NPE
				log.Printf("歌曲保存失败: %s | %v", mediaFile.Path, err)
				return fmt.Errorf("歌曲元数据保存失败 | 路径:%s | %w", mediaFile.Path, err)
			}
		}
		if mediaFileCue != nil {
			// 修复：移除短声明+添加安全访问
			if _, err := uc.mediaCueRepo.Upsert(ctx, mediaFileCue); err != nil {
				cuePath := "unknown path"
				if len(mediaFileCue.CueResources.CuePath) > 0 {
					cuePath = mediaFileCue.CueResources.CuePath
				}
				log.Printf("CUE文件保存失败: %s | %v", cuePath, err)
				// 修复：CUE错误应返回但不中断核心流程
			}
		}
		return nil
	}

	// 处理艺术家
	if artists != nil {
		for _, artist := range artists {
			// 修复：添加空指针检查
			if artist == nil {
				log.Print("艺术家元数据为空")
				continue
			}
			if artist.Name == "" {
				log.Print("艺术家名称为空")
				artist.Name = "Unknown"
			} else {
				artist.NamePinyin = pinyin.LazyConvert(artist.Name, nil)
				artist.NamePinyinFull = strings.Join(artist.NamePinyin, "")
			}
			if err := uc.updateAudioArtistMetadata(ctx, artist); err != nil {
				log.Printf("艺术家处理失败: %s | %v", artist.Name, err)
				return fmt.Errorf("艺术家处理失败 | 原因:%w", err)
			}
		}
	}

	// 处理专辑 - 修复：添加空专辑指针检查
	if album != nil {
		if album.Name == "" {
			log.Print("专辑名称为空")
			album.Name = "Unknown"
		}
		if album.Artist == "" {
			log.Print("专辑艺术家名称为空")
			album.Artist = "Unknown Artist"
		} else {
			album.ArtistPinyin = pinyin.LazyConvert(album.Artist, nil)
			album.ArtistPinyinFull = strings.Join(album.ArtistPinyin, "")
		}
		if err := uc.updateAudioAlbumMetadata(ctx, album); err != nil {
			log.Printf("专辑处理失败: %s | %v", album.Name, err)
			return fmt.Errorf("专辑处理失败 | 名称:%s | 原因:%w", album.Name, err)
		}
	}

	// 保存媒体文件
	if mediaFile != nil {
		if mediaFile.Artist == "" {
			log.Print("媒体文件艺术家名称为空")
			mediaFile.Artist = "Unknown Artist"
		} else {
			mediaFile.ArtistPinyin = pinyin.LazyConvert(mediaFile.Artist, nil)
			mediaFile.ArtistPinyinFull = strings.Join(mediaFile.ArtistPinyin, "")
		}
		// 修复：移除短声明避免NPE
		if _, err := uc.mediaRepo.Upsert(ctx, mediaFile); err != nil {
			errorInfo := fmt.Sprintf("路径:%s", mediaFile.Path)
			if album != nil {
				errorInfo += fmt.Sprintf(" 专辑:%s", album.Name)
			}
			log.Printf("最终保存失败: %s | %v", errorInfo, err)
			return fmt.Errorf("歌曲写入失败 %s | %w", errorInfo, err)
		}
	}

	// 修复：分离媒体文件保存逻辑并添加空指针保护
	if mediaFileCue != nil {
		if mediaFileCue.Performer == "" {
			log.Print("CUE文件艺术家名称为空")
			mediaFileCue.Performer = "Unknown Artist"
		} else {
			mediaFileCue.PerformerPinyin = pinyin.LazyConvert(mediaFileCue.Performer, nil)
			mediaFileCue.PerformerPinyinFull = strings.Join(mediaFileCue.PerformerPinyin, "")
		}
		if _, err := uc.mediaCueRepo.Upsert(ctx, mediaFileCue); err != nil {
			cuePath := "unknown cue path"
			if len(mediaFileCue.CueResources.CuePath) > 0 {
				cuePath = mediaFileCue.CueResources.CuePath
			}

			errorInfo := cuePath
			if album != nil {
				errorInfo += fmt.Sprintf(" 专辑:%s", album.Name)
			}

			log.Printf("CUE最终保存失败: %s | %v", errorInfo, err)
			// 修复：CUE错误应返回但不中断核心流程
		}
	}

	// 异步统计更新
	go uc.updateAudioArtistAndAlbumStatistics(artists, album, mediaFile, mediaFileCue)
	return nil
}

func (uc *FileUsecase) updateAudioArtistAndAlbumStatistics(
	artists []*scene_audio_db_models.ArtistMetadata,
	album *scene_audio_db_models.AlbumMetadata,
	mediaFile *scene_audio_db_models.MediaFileMetadata,
	mediaFileCue *scene_audio_db_models.MediaFileCueMetadata,
) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("统计更新发生panic: %v", r)
		}
	}()

	ctx := context.Background()
	var artistID, albumID primitive.ObjectID

	if artists != nil {
		for index, artist := range artists {
			if artist != nil && !artist.ID.IsZero() {
				artistID = artist.ID
			}
			if !artistID.IsZero() && index == 0 {
				if mediaFileCue != nil {
					if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "size", mediaFileCue.Size); err != nil {
						log.Printf("专辑大小统计更新失败: %v", err)
					}
					if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "duration", int(mediaFileCue.CueDuration)); err != nil {
						log.Printf("专辑播放时间统计更新失败: %v", err)
					}
				}
				if mediaFile != nil {
					if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "size", mediaFile.Size); err != nil {
						log.Printf("专辑大小统计更新失败: %v", err)
					}
					if _, err := uc.artistRepo.UpdateCounter(ctx, artistID, "duration", int(mediaFile.Duration)); err != nil {
						log.Printf("专辑播放时间统计更新失败: %v", err)
					}
				}
			}
		}
	}

	if album != nil && !album.ID.IsZero() {
		albumID = album.ID
	}
	if !albumID.IsZero() {
		if mediaFile != nil {
			if _, err := uc.albumRepo.UpdateCounter(ctx, albumID, "song_count", 1); err != nil {
				log.Printf("专辑单曲数量统计更新失败: %v", err)
			}
			if _, err := uc.albumRepo.UpdateCounter(ctx, albumID, "size", mediaFile.Size); err != nil {
				log.Printf("专辑大小统计更新失败: %v", err)
			}
			if _, err := uc.albumRepo.UpdateCounter(ctx, albumID, "duration", int(mediaFile.Duration)); err != nil {
				log.Printf("专辑播放时间统计更新失败: %v", err)
			}
		}
	}
}

func (uc *FileUsecase) updateAudioArtistMetadata(ctx context.Context, artist *scene_audio_db_models.ArtistMetadata) error {
	if uc.artistRepo == nil {
		log.Print("艺术家仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	existing, err := uc.artistRepo.GetByName(ctx, artist.Name)
	if err != nil {
		log.Printf("名称查询错误: %v", err)
	}
	// 仅同步ID，因为album会随着更新而新增字段
	if existing != nil {
		artist.ID = existing.ID
	}
	// 再次插入，将版本更新的字段覆盖到现有文档
	if err := uc.artistRepo.Upsert(ctx, artist); err != nil {
		log.Printf("艺术家创建失败: %s | %v", artist.Name, err)
		return err
	}

	return nil
}

func (uc *FileUsecase) updateAudioAlbumMetadata(ctx context.Context, album *scene_audio_db_models.AlbumMetadata) error {
	if uc.albumRepo == nil {
		log.Print("专辑仓库未初始化")
		return fmt.Errorf("系统服务异常")
	}

	filter := bson.M{
		"_id":       album.ID,
		"artist_id": album.ArtistID,
		"name":      album.Name,
	}

	existing, err := uc.albumRepo.GetByFilter(ctx, filter)
	if err != nil {
		log.Printf("组合查询错误: %v", err)
	}
	// 仅同步ID，因为album会随着更新而新增字段
	if existing != nil {
		album.ID = existing.ID
	}
	// 再次插入，将版本更新的字段覆盖到现有文档
	if err := uc.albumRepo.Upsert(ctx, album); err != nil {
		log.Printf("专辑创建失败: %s | %v", album.Name, err)
		return err
	}

	return nil
}
