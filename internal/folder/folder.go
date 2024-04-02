package folder

import (
	"github.com/HardDie/fsentry/pkg/fsentry"
)

type Service interface {
	Create(path, name string, data interface{}) (*fsentry.FolderInfo, error)
	Get(path, name string) (*fsentry.FolderInfo, error)
	Move(path, oldName, newName string) (*fsentry.FolderInfo, error)
	Update(path, name string, data interface{}) (*fsentry.FolderInfo, error)
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) (*fsentry.FolderInfo, error)
	MoveWithoutTimestamp(path, oldName, newName string) (*fsentry.FolderInfo, error)
}
