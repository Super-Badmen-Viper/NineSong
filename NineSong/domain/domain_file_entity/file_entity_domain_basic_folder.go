package domain_file_entity

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type FolderMeta struct {
	FileCount   int       `bson:"file_count"`
	LastScanned time.Time `bson:"last_scanned"`
}

type LibraryFolderMetadata struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	FolderPath string             `bson:"folder_path"`
	FileTypes  []FileTypeNo       `bson:"file_types"`
	FolderMeta FolderMeta         `bson:"folder_meta"`
}

type FolderRepository interface {
	Insert(ctx context.Context, folder *LibraryFolderMetadata) error
	FindLibrary(ctx context.Context, folderPath string, fileTypes []FileTypeNo) (*LibraryFolderMetadata, error)
	GetAllLibrary(ctx context.Context, fileType FileTypeNo) ([]*LibraryFolderMetadata, error)
	UpdateStats(ctx context.Context, folderID primitive.ObjectID, fileCount int) error
}

var LibraryMusicType = []FileTypeNo{
	1, // Music
}
