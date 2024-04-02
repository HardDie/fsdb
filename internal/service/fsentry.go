package service

import (
	"path/filepath"
	"sync"

	"github.com/HardDie/fsentry/internal/binary"
	"github.com/HardDie/fsentry/internal/entry"
	"github.com/HardDie/fsentry/internal/folder"
	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/pkg/fsentry"
)

var (
	// validate interface.
	_ fsentry.IFSEntry = &Service{}
)

type Service struct {
	log      fsentry.Logger
	root     string
	rwm      sync.RWMutex
	isPretty bool

	fs     fs.FS
	binary binary.Service
	entry  entry.Service
	folder folder.Service
}

func New(
	log fsentry.Logger,
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
func (s *Service) List(path ...string) (*fsentry.List, error) {
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
