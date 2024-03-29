package binary_new

import (
	"path/filepath"
	"time"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

const (
	binaryFileSuffix = ".bin"
)

type Binary interface {
	Create(path, name string, data []byte) error
	Get(path, name string) ([]byte, error)
	Move(path, oldName, newName string) error
	Update(path, name string, data []byte) error
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) ([]byte, error)
}

type binary struct {
	fs       fs.FS
	isPretty bool
	now      func() time.Time
}

func NewBinary(
	fs fs.FS,
	isPretty bool,
) Binary {
	return binary{
		fs:       fs,
		isPretty: isPretty,
		now:      time.Now,
	}
}

func (s binary) Create(path, name string, data []byte) error {
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	return s.fs.CreateFile(fullPath, data)
}
func (s binary) Get(path, name string) ([]byte, error) {
	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	return s.fs.ReadFile(fullPath)
}
func (s binary) Move(path, oldName, newName string) error {
	oldID := utils.NameToID(oldName)
	if oldID == "" {
		return fsentry_error.ErrorBadName
	}

	newID := utils.NameToID(newName)
	if newID == "" {
		return fsentry_error.ErrorBadName
	}

	oldFullPath := filepath.Join(path, oldID+binaryFileSuffix)
	newFullPath := filepath.Join(path, newID+binaryFileSuffix)

	return s.fs.Rename(oldFullPath, newFullPath)
}
func (s binary) Update(path, name string, data []byte) error {
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	return s.fs.UpdateFile(fullPath, data)
}
func (s binary) Remove(path, name string) error {
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	return s.fs.RemoveFile(fullPath)
}
func (s binary) Duplicate(path, oldName, newName string) ([]byte, error) {
	data, err := s.Get(path, oldName)
	if err != nil {
		return nil, err
	}
	err = s.Create(path, newName, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
