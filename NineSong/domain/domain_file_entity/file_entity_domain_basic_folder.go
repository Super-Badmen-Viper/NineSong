package domain_file_entity

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type LibraryStatus int

const (
	StatusActive LibraryStatus = iota + 1
	StatusDisabled
	StatusScanning
)

type LibraryFolderMetadata struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name" validate:"required"`
	FolderPath  string             `bson:"folder_path" validate:"required,filepath"`
	FileTypes   []FileTypeNo       `bson:"file_types"`
	Status      LibraryStatus      `bson:"status"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	LastScanned time.Time          `bson:"last_scanned"`
	FileCount   int                `bson:"file_count"`
}

var (
	ErrLibraryNotFound  = errors.New("media library not found")
	ErrLibraryDuplicate = errors.New("duplicate media library")
	ErrLibraryInUse     = errors.New("library is currently in use")
)

type FolderRepository interface {
	Insert(ctx context.Context, folder *LibraryFolderMetadata) error
	FindLibrary(ctx context.Context, folderPath string, fileTypes []FileTypeNo) (*LibraryFolderMetadata, error)
	GetAllByType(ctx context.Context, fileType FileTypeNo) ([]*LibraryFolderMetadata, error)
	UpdateStats(ctx context.Context, folderID primitive.ObjectID, fileCount int, status LibraryStatus) error

	Create(ctx context.Context, folder *LibraryFolderMetadata) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetAll(ctx context.Context) ([]*LibraryFolderMetadata, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*LibraryFolderMetadata, error)
	DetectingDuplicates(ctx context.Context, FolderPath string, FileTypes []FileTypeNo) (*LibraryFolderMetadata, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status LibraryStatus) error
}

var LibraryMusicType = []FileTypeNo{
	1, // Music
}
