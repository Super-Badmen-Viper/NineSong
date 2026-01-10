package usecase_file_entity

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LibraryUsecase interface {
	BrowseFolders(ctx context.Context, path string) ([]domain_file_entity.FolderEntry, error)

	CreateLibrary(ctx context.Context, name, path string, folderType int) (*domain_file_entity.LibraryFolderMetadata, error)
	DeleteLibrary(ctx context.Context, id string) error
	UpdateLibrary(ctx context.Context,
		id primitive.ObjectID,
		newName string,
		newFolderPath string) error
	GetLibraries(ctx context.Context) ([]*domain_file_entity.LibraryFolderMetadata, error)
}

// 新增：LibraryUsecaseWithScan 接口，包含扫描相关方法
type LibraryUsecaseWithScan interface {
	LibraryUsecase
	SetFileUsecase(uc *FileUsecase)
}

type libraryUsecase struct {
	folderRepo  domain_file_entity.FolderRepository
	fileUsecase *FileUsecase
}

func NewLibraryUsecase(repo domain_file_entity.FolderRepository) LibraryUsecaseWithScan {
	return &libraryUsecase{folderRepo: repo}
}

// 新增：SetFileUsecase 设置FileUsecase引用
func (uc *libraryUsecase) SetFileUsecase(fileUsecase *FileUsecase) {
	uc.fileUsecase = fileUsecase
}

func (uc *libraryUsecase) BrowseFolders(ctx context.Context, path string) ([]domain_file_entity.FolderEntry, error) {
	return uc.folderRepo.ListFolders(path)
}

func (uc *libraryUsecase) CreateLibrary(ctx context.Context, name, path string, folderType int) (*domain_file_entity.LibraryFolderMetadata, error) {
	// 路径验证
	if !filepath.IsAbs(path) {
		return nil, errors.New("path must be absolute")
	}

	// 检查是否有正在进行的扫描任务
	if uc.fileUsecase != nil {
		uc.fileUsecase.scanMutex.RLock()
		activeScan := uc.fileUsecase.activeScanCount > 0
		uc.fileUsecase.scanMutex.RUnlock()

		if activeScan {
			return nil, domain_file_entity.ErrLibraryInUse
		}
	}

	library := &domain_file_entity.LibraryFolderMetadata{
		Name:       name,
		FolderPath: filepath.Clean(path),
		FolderType: folderType,
	}

	if err := uc.folderRepo.Create(ctx, library); err != nil {
		return nil, err
	}

	if uc.fileUsecase != nil {
		log.Printf("媒体库创建成功，开始异步扫描: %s", library.FolderPath)

		go func() {
			scanCtx := context.Background()
			if err := uc.fileUsecase.ProcessDirectory(
				scanCtx,
				[]string{library.FolderPath},
				folderType,
				0,
			); err != nil {
				log.Printf("异步扫描失败: %v", err)
			} else {
				log.Printf("异步扫描完成: %s", library.FolderPath)
			}
		}()
	}

	return library, nil
}

func (uc *libraryUsecase) DeleteLibrary(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid library ID")
	}

	if lib, err := uc.folderRepo.GetByID(ctx, objID); err == nil && lib.Status == domain_file_entity.StatusScanning {
		return domain_file_entity.ErrLibraryInUse
	}

	var deletedLibraryPath string
	if lib, err := uc.folderRepo.GetByID(ctx, objID); err == nil {
		deletedLibraryPath = lib.FolderPath
	}

	err = uc.folderRepo.DeleteByID(ctx, objID)
	if err != nil {
		return err
	}

	libraries, err := uc.folderRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("检查媒体库数量失败: %w", err)
	}

	log.Printf("DeleteLibrary: libraries count=%d, fileUsecase=%v, deletedPath=%s", len(libraries), uc.fileUsecase != nil, deletedLibraryPath)

	if len(libraries) == 0 {
		log.Printf("已删除最后一个媒体库，异步清空所有媒体数据")
		if uc.fileUsecase != nil && uc.fileUsecase.libraryDeleteUsecase != nil {
			go func() {
				clearCtx := context.Background()
				if err := uc.fileUsecase.libraryDeleteUsecase.ClearAllData(clearCtx); err != nil {
					log.Printf("异步清空媒体数据失败: %v", err)
				}
			}()
		}
		return nil
	}

	if deletedLibraryPath != "" && uc.fileUsecase != nil {
		log.Printf("媒体库已删除，开始异步清理: %s", deletedLibraryPath)

		taskID := primitive.NewObjectID().Hex()
		taskProg := &domain_util.TaskProgress{
			ID:          taskID,
			Status:      "processing",
			TotalFiles:  5000,
			Initialized: true,
		}

		uc.fileUsecase.activeTasksMu.Lock()
		uc.fileUsecase.activeTasks[taskID] = taskProg
		uc.fileUsecase.activeTasksMu.Unlock()

		go func() {
			defer func() {
				uc.fileUsecase.activeTasksMu.Lock()
				delete(uc.fileUsecase.activeTasks, taskID)
				uc.fileUsecase.activeTasksMu.Unlock()
			}()

			deleteCtx := context.Background()

			remainingLibraries, _ := uc.folderRepo.GetAll(deleteCtx)
			var remainingPaths []string
			for _, lib := range remainingLibraries {
				remainingPaths = append(remainingPaths, lib.FolderPath)
			}

			if uc.fileUsecase.libraryDeleteUsecase != nil {
				log.Printf("异步清理被删除媒体库的元数据")
				if _, err := uc.fileUsecase.libraryDeleteUsecase.DeleteByLibraryPaths(deleteCtx, []string{deletedLibraryPath}); err != nil {
					log.Printf("异步清理媒体库元数据失败: %v", err)
				} else {
					log.Printf("异步清理被删除媒体库的元数据完成")
				}

				if len(remainingPaths) > 0 {
					log.Printf("异步清理孤儿数据")
					if deletedCount, err := uc.fileUsecase.libraryDeleteUsecase.CleanupOrphanedMedia(deleteCtx, remainingPaths, taskProg); err != nil {
						log.Printf("异步清理孤儿数据失败: %v", err)
					} else {
						log.Printf("已异步清理 %d 个孤儿媒体文件", deletedCount)
					}
				}
			}
		}()
	}

	return nil
}

func (uc *libraryUsecase) UpdateLibrary(
	ctx context.Context,
	id primitive.ObjectID,
	newName string,
	newFolderPath string,
) error {
	// 1. 检查媒体库状态
	if inUse, err := uc.folderRepo.IsLibraryInUse(ctx, id); err != nil {
		return fmt.Errorf("状态检查失败: %w", err)
	} else if inUse {
		return domain_file_entity.ErrLibraryInUse
	}

	// 2. 检查是否有活跃的扫描任务
	if uc.fileUsecase != nil {
		uc.fileUsecase.scanMutex.RLock()
		hasActiveScan := uc.fileUsecase.activeScanCount > 0
		uc.fileUsecase.scanMutex.RUnlock()
		if hasActiveScan {
			return domain_file_entity.ErrLibraryInUse
		}
	}

	dup, err := uc.folderRepo.DetectingDuplicates(ctx, newFolderPath, 1)
	if err != nil {
		return fmt.Errorf("duplicate detection failed: %w", err)
	}
	if dup != nil && dup.ID != id {
		return domain_file_entity.ErrLibraryDuplicate
	}

	return uc.folderRepo.Upsert(ctx, id, newName, newFolderPath)
}

func (uc *libraryUsecase) GetLibraries(ctx context.Context) ([]*domain_file_entity.LibraryFolderMetadata, error) {
	return uc.folderRepo.GetAll(ctx)
}
