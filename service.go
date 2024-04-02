package fsentry

import (
	"path/filepath"
	"sync"

	"github.com/HardDie/fsentry/dto"
	"github.com/HardDie/fsentry/internal/binary"
	"github.com/HardDie/fsentry/internal/entry"
	"github.com/HardDie/fsentry/internal/folder"
	"github.com/HardDie/fsentry/internal/fs"
)

type Binary interface {
	Create(path, name string, data []byte) error
	Get(path, name string) ([]byte, error)
	Move(path, oldName, newName string) error
	Update(path, name string, data []byte) error
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) ([]byte, error)
}

type Entry interface {
	Create(path, name string, data interface{}) (*dto.Entry, error)
	Get(path, name string) (*dto.Entry, error)
	Move(path, oldName, newName string) (*dto.Entry, error)
	Update(path, name string, data interface{}) (*dto.Entry, error)
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) (*dto.Entry, error)
}

type Folder interface {
	Create(path, name string, data interface{}) (*dto.FolderInfo, error)
	Get(path, name string) (*dto.FolderInfo, error)
	Move(path, oldName, newName string) (*dto.FolderInfo, error)
	Update(path, name string, data interface{}) (*dto.FolderInfo, error)
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) (*dto.FolderInfo, error)
	MoveWithoutTimestamp(path, oldName, newName string) (*dto.FolderInfo, error)
}

type Service struct {
	log      Logger
	root     string
	rwm      sync.RWMutex
	isPretty bool

	fs     fs.FS
	binary Binary
	entry  Entry
	folder Folder
}

func New(
	log Logger,
	root string,
	isPretty bool,
	fs fs.FS,
	binary binary.Service,
	entry entry.Service,
	folder folder.Service,
) *Service {
	return &Service{
		log:      log,
		root:     root,
		isPretty: isPretty,
		fs:       fs,
		binary:   binary,
		entry:    entry,
		folder:   folder,
	}
}

// Init check if a repository folder has been created and if not, create one.
func (s *Service) Init() error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	// Check if db folder exist
	isExist, err := s.fs.IsFolderExist(s.root)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}
	err = s.fs.CreateAllFolder(s.root)
	if err != nil {
		return err
	}
	return nil
}

// Drop if you want to delete the fsentry repository you can use this method.
func (s *Service) Drop() error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	// Check if db folder exist
	isExist, err := s.fs.IsFolderExist(s.root)
	if err != nil {
		return err
	}
	if !isExist {
		return nil
	}

	// Remove db folder
	err = s.fs.RemoveFolder(s.root)
	if err != nil {
		return err
	}

	return nil
}

// List allows you to get a list of objects (folders and entries) on the selected path.
func (s *Service) List(path ...string) (*dto.List, error) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	fullPath := s.buildPath(path...)
	_, err := s.fs.List(fullPath)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Service) buildPath(path ...string) string {
	pathSlice := append([]string{s.root}, path...)
	return filepath.Join(pathSlice...)
}
