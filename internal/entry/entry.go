package entry

import (
	"github.com/HardDie/fsentry/pkg/fsentry"
)

type Service interface {
	Create(path, name string, data interface{}) (*fsentry.Entry, error)
	Get(path, name string) (*fsentry.Entry, error)
	Move(path, oldName, newName string) (*fsentry.Entry, error)
	Update(path, name string, data interface{}) (*fsentry.Entry, error)
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) (*fsentry.Entry, error)
}
