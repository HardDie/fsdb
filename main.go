package fsentry

import (
	"path/filepath"
	"sync"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/internal/entry_error"
	"github.com/HardDie/fsentry/internal/fsutils"
	"github.com/HardDie/fsentry/internal/utils"
)

type IFSEntry interface {
	Init() error
	Drop() error

	CreateFolder(name string, data interface{}, path ...string) error
	GetFolder(name string, path ...string) (*entity.FolderInfo, error)
	MoveFolder(oldName, newName string, path ...string) error
	UpdateFolder(name string, data interface{}, path ...string) error
	RemoveFolder(name string, path ...string) error
	DuplicateFolder(srcName, dstName string, path ...string) error
}
type FSEntry struct {
	root string
	rwm  sync.RWMutex
}

var (
	// validate interface
	_ IFSEntry = &FSEntry{}
)

func NewFSEntry(root string) IFSEntry {
	return &FSEntry{
		root: root,
	}
}

// Basic

func (db *FSEntry) Init() error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	// Check if db folder exist
	isExist, err := fsutils.IsFolderExist(db.root)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}
	err = fsutils.CreateAllFolder(db.root)
	if err != nil {
		return err
	}
	return nil
}
func (db *FSEntry) Drop() error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	// Check if db folder exist
	isExist, err := fsutils.IsFolderExist(db.root)
	if err != nil {
		return err
	}
	if !isExist {
		return nil
	}

	// Remove db folder
	err = fsutils.RemoveFolder(db.root)
	if err != nil {
		return err
	}

	return nil
}

// Folder

func (db *FSEntry) CreateFolder(name string, data interface{}, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	fullPath, err := db.isFolderNotExist(name, path...)
	if err != nil {
		return err
	}

	// Create folder
	err = fsutils.CreateFolder(fullPath)
	if err != nil {
		return err
	}

	// Create info file
	info := entity.NewFolderInfo(name, data)
	err = fsutils.CreateInfo(fullPath, info)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) GetFolder(name string, path ...string) (*entity.FolderInfo, error) {
	db.rwm.RLock()
	defer db.rwm.RUnlock()

	fullPath, err := db.isFolderExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := fsutils.GetInfo(fullPath)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (db *FSEntry) MoveFolder(oldName, newName string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	// Check if source folder exist
	fullOldPath, err := db.isFolderExist(oldName, path...)
	if err != nil {
		return err
	}

	// Check if destination folder not exist
	fullNewPath, err := db.isFolderNotExist(newName, path...)
	if err != nil {
		return err
	}

	// Get info from file
	info, err := fsutils.GetInfo(fullOldPath)
	if err != nil {
		return err
	}

	info.SetName(newName).UpdatedNow()

	// Update info file
	err = fsutils.CreateInfo(fullOldPath, info)
	if err != nil {
		return err
	}

	return fsutils.MoveFolder(fullOldPath, fullNewPath)
}
func (db *FSEntry) UpdateFolder(name string, data interface{}, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	fullPath, err := db.isFolderExist(name, path...)
	if err != nil {
		return err
	}

	// Get info from file
	info, err := fsutils.GetInfo(fullPath)
	if err != nil {
		return err
	}

	err = info.UpdateData(data)
	if err != nil {
		return err
	}

	// Update info file
	err = fsutils.CreateInfo(fullPath, info)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) RemoveFolder(name string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	fullPath, err := db.isFolderExist(name, path...)
	if err != nil {
		return err
	}

	err = fsutils.RemoveFolder(fullPath)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) DuplicateFolder(srcName, dstName string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	// Check if source folder exist
	fullSrcPath, err := db.isFolderExist(srcName, path...)
	if err != nil {
		return err
	}

	// Check if destination folder not exist
	fullDstPath, err := db.isFolderNotExist(dstName, path...)
	if err != nil {
		return err
	}

	// Copy folder
	err = fsutils.CopyFolder(fullSrcPath, fullDstPath)
	if err != nil {
		return err
	}

	// Get info from file
	info, err := fsutils.GetInfo(fullDstPath)
	if err != nil {
		return err
	}

	info.SetName(dstName).FlushTime()

	// Update info file
	err = fsutils.CreateInfo(fullDstPath, info)
	if err != nil {
		return err
	}

	return nil
}

// util

func (db *FSEntry) buildPath(id string, path ...string) string {
	pathSlice := append([]string{db.root}, path...)
	return filepath.Join(append(pathSlice, id)...)
}
func (db *FSEntry) isFolderExist(name string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", entry_error.ErrorBadFolderName
	}

	fullPath := db.buildPath(id, path...)

	// Check if root folder exist
	isExist, err := fsutils.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", entry_error.ErrorBadPath
	}

	// Check if destination folder exist
	isExist, err = fsutils.IsFolderExist(fullPath)
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", entry_error.ErrorNotExist
	}

	return fullPath, nil
}
func (db *FSEntry) isFolderNotExist(name string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", entry_error.ErrorBadFolderName
	}

	fullPath := db.buildPath(id, path...)

	// Check if destination folder exist
	isExist, err := fsutils.IsFolderExist(fullPath)
	if err != nil {
		return "", err
	}
	if isExist {
		return "", entry_error.ErrorExist
	}

	return fullPath, nil
}
