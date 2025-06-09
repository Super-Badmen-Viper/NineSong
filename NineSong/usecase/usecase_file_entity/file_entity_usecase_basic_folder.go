package usecase_file_entity

import (
	"context"
	"errors"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_file_entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"path/filepath"
)

type LibraryUsecase interface {
	CreateLibrary(ctx context.Context, name, path string, fileTypes []domain_file_entity.FileTypeNo) (*domain_file_entity.LibraryFolderMetadata, error)
	DeleteLibrary(ctx context.Context, id string) error
	GetLibraries(ctx context.Context) ([]*domain_file_entity.LibraryFolderMetadata, error)
}

type libraryUsecase struct {
	folderRepo domain_file_entity.FolderRepository
}

func NewLibraryUsecase(repo domain_file_entity.FolderRepository) LibraryUsecase {
	return &libraryUsecase{folderRepo: repo}
}

func (uc *libraryUsecase) CreateLibrary(ctx context.Context, name, path string, fileTypes []domain_file_entity.FileTypeNo) (*domain_file_entity.LibraryFolderMetadata, error) {
	// 路径验证
	if !filepath.IsAbs(path) {
		return nil, errors.New("path must be absolute")
	}

	library := &domain_file_entity.LibraryFolderMetadata{
		Name:       name,
		FolderPath: filepath.Clean(path),
		FileTypes:  fileTypes,
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
	return uc.folderRepo.Delete(ctx, objID)
}

func (uc *libraryUsecase) GetLibraries(ctx context.Context) ([]*domain_file_entity.LibraryFolderMetadata, error) {
	return uc.folderRepo.GetAll(ctx)
}
