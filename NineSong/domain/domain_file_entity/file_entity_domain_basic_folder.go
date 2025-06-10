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

type FolderType int

const (
	MusicLibrary FolderType = iota + 1
	VideoLibrary
	DocumentLibrary
)

type FolderEntry struct {
	Name   string `json:"name"`    // 文件夹名称
	Path   string `json:"path"`    // 完整路径
	IsRoot bool   `json:"is_root"` // 是否为根目录
}

type LibraryFolderMetadata struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name" validate:"required"`
	FolderPath  string             `bson:"folder_path" validate:"required,filepath"`
	FolderType  int                `bson:"folder_type"`
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
	ListFolders(path string) ([]FolderEntry, error)

	Insert(ctx context.Context, folder *LibraryFolderMetadata) error
	FindLibrary(ctx context.Context, folderPath string, folderType int) (*LibraryFolderMetadata, error)
	GetAllByType(ctx context.Context, folderType int) ([]*LibraryFolderMetadata, error)
	UpdateStats(ctx context.Context, folderID primitive.ObjectID, fileCount int, status LibraryStatus) error

	Create(ctx context.Context, folder *LibraryFolderMetadata) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Update(ctx context.Context, id primitive.ObjectID, newName string, newFolderPath string) error
	IsLibraryInUse(ctx context.Context, id primitive.ObjectID) (bool, error)

	GetAll(ctx context.Context) ([]*LibraryFolderMetadata, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*LibraryFolderMetadata, error)
	DetectingDuplicates(ctx context.Context, folderPath string, folderType int) (*LibraryFolderMetadata, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status LibraryStatus) error
}
