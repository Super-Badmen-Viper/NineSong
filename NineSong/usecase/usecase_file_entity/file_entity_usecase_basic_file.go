package usecase_file_entity

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase/usecase_file_entity/scene_audio"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity/scene_audio/scene_audio_db/scene_audio_db_interface"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScanManager 全局状态管理器
type ScanManager struct {
	mu                  sync.RWMutex
	globalScanRunning   bool                          // 全局扫描任务状态
	concurrentScanCount int                           // 并发扫描任务计数
	cancelFuncs         map[string]context.CancelFunc // 所有扫描任务的取消函数
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

type FileUsecase struct {
	db               mongo.Database
	fileRepo         domain_file_entity.FileRepository
	folderRepo       domain_file_entity.FolderRepository
	detector         domain_file_entity.FileDetector
	folderType       map[domain_file_entity.FileTypeNo]struct{}
	targetMutex      sync.RWMutex
	workerPool       chan struct{}
	scanTimeout      time.Duration
	activeTasks      map[string]*domain_util.TaskProgress // 任务ID -> 进度跟踪器
	activeTasksMu    sync.RWMutex                         // 保护 activeTasks
	scanManager      *ScanManager                         // 新增扫描状态管理器
	scanningPaths    map[string]bool                      // 新增：记录正在扫描的路径
	scanningPathsMu  sync.RWMutex                         // 保护scanningPaths的互斥锁
	activeScanCount  int                                  // 新增：当前活跃的扫描任务数量
	scanProgress     float32
	scanMutex        sync.RWMutex
	lastScanStart    time.Time
	scanStageWeights struct {
		traversal   float32
		statistics  float32
		refactoring float32
	}
	audioProcessingUsecase *scene_audio.AudioProcessingUsecase
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
	wordCloudRepo scene_audio_db_interface.WordCloudDBRepository,
	lyricsFileRepo scene_audio_db_interface.LyricsFileRepository,
) *FileUsecase {
	workerCount := runtime.NumCPU() * 2
	if workerCount < 4 {
		workerCount = 4
	}

	audioProcessingUsecase := scene_audio.NewAudioProcessingUsecase(
		db,
		fileRepo,
		folderRepo,
		detector,
		timeoutMinutes,
		artistRepo,
		albumRepo,
		mediaRepo,
		tempRepo,
		mediaCueRepo,
		wordCloudRepo,
		lyricsFileRepo,
	)

	return &FileUsecase{
		db:                     db,
		fileRepo:               fileRepo,
		folderRepo:             folderRepo,
		detector:               detector,
		workerPool:             make(chan struct{}, workerCount),
		scanTimeout:            time.Duration(timeoutMinutes) * time.Minute,
		scanningPaths:          make(map[string]bool),
		scanManager:            NewScanManager(),                           // 初始化扫描管理器
		activeTasks:            make(map[string]*domain_util.TaskProgress), // 新增初始化
		audioProcessingUsecase: audioProcessingUsecase,
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
		taskStatuses[id] = task.Status

		if task.Status == "preparing" {
			totalProgress += 0.01 // 准备中状态返回1%
			taskCount++
			continue
		}

		if !task.Initialized {
			totalProgress += 0.01
			taskCount++
			continue
		}

		walked := atomic.LoadInt32(&task.WalkedFiles)
		processed := atomic.LoadInt32(&task.ProcessedFiles)
		total := atomic.LoadInt32(&task.TotalFiles)

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
	taskProg := &domain_util.TaskProgress{
		ID:     taskID,
		Status: "preparing", // 初始状态
	}

	// 注册任务
	uc.activeTasks[taskID] = taskProg
	// 开始统计文件数
	taskProg.Status = "counting_files"
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

	if libraryFolders == nil {
		libraryFolders = make([]*domain_file_entity.LibraryFolderMetadata, 0)
	}

	// 4. 根据扫描模式执行处理
	switch ScanModel {
	case 0: // 扫描新的和有修改的文件
		if folderType == 1 {
			err := uc.audioProcessingUsecase.ProcessMusicDirectory(ctx, libraryFolders, taskProg, true, true, false)
			if err != nil {
				return err
			}
		}
	case 1: // 搜索缺失的元数据
		if folderType == 1 {
			err := uc.audioProcessingUsecase.ProcessMusicDirectory(ctx, libraryFolders, taskProg, true, false, false)
			if err != nil {
				return err
			}
		}
	case 2: // 覆盖所有元数据
		if folderType == 1 {
			err := uc.audioProcessingUsecase.ProcessMusicDirectory(ctx, libraryFolders, taskProg, true, true, true)
			if err != nil {
				return err
			}
		}
	case 3: // 重构媒体库
		if folderType == 1 {
			err := uc.audioProcessingUsecase.ProcessMusicDirectory(ctx, libraryFolders, taskProg, true, true, true)
			if err != nil {
				return err
			}
		}
	}

	log.Printf("媒体库扫描完成，共处理%d个目录", len(dirPaths))

	return nil
}

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
