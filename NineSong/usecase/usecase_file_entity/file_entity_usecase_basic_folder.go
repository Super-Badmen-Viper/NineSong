package usecase_file_entity

import (
	"context"
	"errors"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"path/filepath"
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

type libraryUsecase struct {
	folderRepo domain_file_entity.FolderRepository
}

func NewLibraryUsecase(repo domain_file_entity.FolderRepository) LibraryUsecase {
	return &libraryUsecase{folderRepo: repo}
}

func (uc *libraryUsecase) BrowseFolders(ctx context.Context, path string) ([]domain_file_entity.FolderEntry, error) {
	return uc.folderRepo.ListFolders(path)
}

func (uc *libraryUsecase) CreateLibrary(ctx context.Context, name, path string, folderType int) (*domain_file_entity.LibraryFolderMetadata, error) {
	// 路径验证
	if !filepath.IsAbs(path) {
		return nil, errors.New("path must be absolute")
	}

	library := &domain_file_entity.LibraryFolderMetadata{
		Name:       name,
		FolderPath: filepath.Clean(path),
		FolderType: folderType,
	}

	if err := uc.folderRepo.Create(ctx, library); err != nil {
		return nil, err
	}
	return library, nil
}

func (uc *libraryUsecase) DeleteLibrary(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid library ID")
	}
	return uc.folderRepo.DeleteByID(ctx, objID)
}

func (uc *libraryUsecase) UpdateLibrary(
	ctx context.Context,
	id primitive.ObjectID,
	newName string,
	newFolderPath string,
) error {
	if inUse, err := uc.folderRepo.IsLibraryInUse(ctx, id); err != nil {
		return fmt.Errorf("状态检查失败: %w", err)
	} else if inUse {
		return domain_file_entity.ErrLibraryInUse
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
