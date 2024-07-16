package binary

import (
	"errors"
	"log"
	"path/filepath"
	"time"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/utils"
)

const (
	binaryFileSuffix = ".bin"
)

var (
	ErrorExist              = errors.New("binary already exist")
	ErrorNotExist           = errors.New("binary not exist")
	ErrorPermission         = errors.New("not enough permissions")
	ErrorBadName            = errors.New("bad name")
	ErrorBadSourceName      = errors.New("bad source name")
	ErrorBadDestinationName = errors.New("bad destination name")
	ErrorBadPath            = errors.New("bad path")
	ErrorInternal           = errors.New("internal")
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
		return ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	err := s.fs.CreateFile(fullPath, data)
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, fs.ErrorFileExist):
		return ErrorExist
	case errors.Is(err, fs.ErrorBadPath):
		return ErrorBadPath
	case errors.Is(err, fs.ErrorPermission):
		return ErrorPermission
	default:
		log.Printf("binary.Create() fs.CreateFile: %T %s", err, err.Error())
		return ErrorInternal
	}
}
func (s Service) Get(path, name string) ([]byte, error) {
	id := utils.NameToID(name)
	if id == "" {
		return nil, ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	data, err := s.fs.ReadFile(fullPath)
	if err == nil {
		return data, nil
	}
	switch {
	case errors.Is(err, fs.ErrorFileNotExist):
		return nil, ErrorNotExist
	case errors.Is(err, fs.ErrorPermission):
		return nil, ErrorPermission
	default:
		log.Printf("binary.Get() fs.ReadFile: %T %s", err, err.Error())
		return nil, ErrorInternal
	}
}
func (s Service) Move(path, oldName, newName string) error {
	oldID := utils.NameToID(oldName)
	if oldID == "" {
		return ErrorBadSourceName
	}

	newID := utils.NameToID(newName)
	if newID == "" {
		return ErrorBadDestinationName
	}

	oldFullPath := filepath.Join(path, oldID+binaryFileSuffix)
	newFullPath := filepath.Join(path, newID+binaryFileSuffix)

	return s.fs.Rename(oldFullPath, newFullPath)
}
func (s Service) Update(path, name string, data []byte) error {
	id := utils.NameToID(name)
	if id == "" {
		return ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	err := s.fs.UpdateFile(fullPath, data)
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, fs.ErrorFileNotExist):
		return ErrorNotExist
	case errors.Is(err, fs.ErrorPermission):
		return ErrorPermission
	default:
		log.Printf("binary.Update() fs.UpdateFile: %T %s", err, err.Error())
		return ErrorInternal
	}
}
func (s Service) Remove(path, name string) error {
	id := utils.NameToID(name)
	if id == "" {
		return ErrorBadName
	}

	fullPath := filepath.Join(path, id+binaryFileSuffix)

	err := s.fs.RemoveFile(fullPath)
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, fs.ErrorNotExist):
		return ErrorNotExist
	case errors.Is(err, fs.ErrorPermission):
		return ErrorPermission
	default:
		log.Printf("binary.Remove() fs.RemoveFile: %T %s", err, err.Error())
		return ErrorInternal
	}
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
