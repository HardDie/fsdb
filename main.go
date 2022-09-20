package fsdb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HardDie/fsdb/internal/entity"
	"github.com/HardDie/fsdb/internal/fsdberror"
	"github.com/HardDie/fsdb/internal/fsutils"
	"github.com/HardDie/fsdb/internal/utils"
)

type IFsDB interface {
	Init() error
	Drop() error

	CreateFolder(name string, data interface{}, path ...string) error
	GetFolder(name string, path ...string) (*entity.FolderInfo, error)
	MoveFolder(oldName, newName string, path ...string) error
	UpdateFolder(name string, data interface{}, path ...string) error
	DeleteFolder(name string, path ...string) error
	DuplicateFolder(srcName, dstName string, path ...string) error
}
type FsDB struct {
	root string
	rwm  sync.RWMutex
}

func NewFsDB(root string) IFsDB {
	return &FsDB{
		root: root,
	}
}

func (db *FsDB) buildPath(id string, path ...string) string {
	pathSlice := append([]string{db.root}, path...)
	return filepath.Join(append(pathSlice, id)...)
}
func (db *FsDB) isFolderExist(name string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsdberror.ErrorBadFolderName
	}

	fullPath := db.buildPath(id, path...)

	// Check if root folder exist
	isExist, err := fsutils.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsdberror.ErrorBadPath
	}

	// Check if destination folder exist
	isExist, err = fsutils.IsFolderExist(fullPath)
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsdberror.ErrorNotExist
	}

	return fullPath, nil
}
func (db *FsDB) isFolderNotExist(name string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsdberror.ErrorBadFolderName
	}

	fullPath := db.buildPath(id, path...)

	// Check if destination folder exist
	isExist, err := fsutils.IsFolderExist(fullPath)
	if err != nil {
		return "", err
	}
	if isExist {
		return "", fsdberror.ErrorExist
	}

	return fullPath, nil
}

func (db *FsDB) Init() error {
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
func (db *FsDB) Drop() error {
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
	err = os.RemoveAll(db.root)
	if err != nil {
		return fsdberror.Wrap(err, fsdberror.ErrorInternal)
	}

	return nil
}

func (db *FsDB) CreateFolder(name string, data interface{}, path ...string) error {
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
func (db *FsDB) GetFolder(name string, path ...string) (*entity.FolderInfo, error) {
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
func (db *FsDB) MoveFolder(oldName, newName string, path ...string) error {
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

	info.Id = utils.NameToID(newName)
	info.Name = newName
	info.UpdatedAt = utils.Allocate(time.Now())

	// Update info file
	err = fsutils.CreateInfo(fullOldPath, info)
	if err != nil {
		return err
	}

	return fsutils.MoveFolder(fullOldPath, fullNewPath)
}
func (db *FsDB) UpdateFolder(name string, data interface{}, path ...string) error {
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

	dataJson, err := json.Marshal(data)
	if err != nil {
		return fsdberror.Wrap(err, fsdberror.ErrorInternal)
	}
	info.Data = dataJson
	info.UpdatedAt = utils.Allocate(time.Now())

	// Update info file
	err = fsutils.CreateInfo(fullPath, info)
	if err != nil {
		return err
	}

	return nil
}
func (db *FsDB) DeleteFolder(name string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	fullPath, err := db.isFolderExist(name, path...)
	if err != nil {
		return err
	}

	err = os.RemoveAll(fullPath)
	if err != nil {
		return fsdberror.Wrap(err, fsdberror.ErrorInternal)
	}

	return nil
}
func (db *FsDB) DuplicateFolder(srcName, dstName string, path ...string) error {
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

	info.Id = utils.NameToID(dstName)
	info.Name = dstName
	info.CreatedAt = utils.Allocate(time.Now())
	info.UpdatedAt = nil

	// Update info file
	err = fsutils.CreateInfo(fullDstPath, info)
	if err != nil {
		return err
	}

	return nil
}
