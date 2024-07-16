package binary

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

type Service struct {
	fs       fs.FS
	isPretty bool
	now      func() time.Time
}

func New(
	fs fs.FS,
	isPretty bool,
) Service {
	return Service{
		fs:       fs,
		isPretty: isPretty,
		now:      time.Now,
	}
}

func (s Service) Create(path, name string, data []byte) error {
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	return s.fs.CreateFile(fullPath, data)
}
func (s Service) Get(path, name string) ([]byte, error) {
	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	return s.fs.ReadFile(fullPath)
}
func (s Service) Move(path, oldName, newName string) error {
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
func (s Service) Update(path, name string, data []byte) error {
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	return s.fs.UpdateFile(fullPath, data)
}
func (s Service) Remove(path, name string) error {
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	return s.fs.RemoveFile(fullPath)
}
func (s Service) Duplicate(path, oldName, newName string) ([]byte, error) {
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
