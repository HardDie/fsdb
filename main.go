// fsentry
//
// Allows storing hierarchical data in files and folders on the file system with json descriptions and creation/update timestamps.
package fsentry

import (
	"sync"

	"github.com/HardDie/fsentry/internal/entity"
	repFS "github.com/HardDie/fsentry/internal/repository/fs"
	serviceCommon "github.com/HardDie/fsentry/internal/service/common"
	serviceFolder "github.com/HardDie/fsentry/internal/service/folder"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

type IFSEntry interface {
	Init() error
	Drop() error
	List(path ...string) (*entity.List, error)

	serviceFolder.Folder

	CreateEntry(name string, data interface{}, path ...string) error
	GetEntry(name string, path ...string) (*entity.Entry, error)
	MoveEntry(oldName, newName string, path ...string) error
	UpdateEntry(name string, data interface{}, path ...string) error
	RemoveEntry(name string, path ...string) error
	DuplicateEntry(srcName, dstName string, path ...string) error

	CreateBinary(name string, data []byte, path ...string) error
	GetBinary(name string, path ...string) ([]byte, error)
	MoveBinary(oldName, newName string, path ...string) error
	UpdateBinary(name string, data []byte, path ...string) error
	RemoveBinary(name string, path ...string) error
}
type FSEntry struct {
	root string
	rwm  sync.RWMutex

	isPretty bool

	fs            repFS.FS
	serviceCommon serviceCommon.Common
	serviceFolder.Folder
}

var (
	// validate interface
	_ IFSEntry = &FSEntry{}
)

func WithPretty() func(fs *FSEntry) {
	return func(fs *FSEntry) {
		fs.isPretty = true
	}
}

func NewFSEntry(root string, ops ...func(fs *FSEntry)) IFSEntry {
	res := &FSEntry{
		root: root,
		fs:   repFS.NewFS(),
	}
	for _, op := range ops {
		op(res)
	}
	res.serviceCommon = serviceCommon.NewCommon(res.root, res.fs)
	res.Folder = serviceFolder.NewFolder(res.root, &res.rwm, res.isPretty, res.fs, res.serviceCommon)
	return res
}

// Basic

// Init check if a repository folder has been created and if not, create one.
func (db *FSEntry) Init() error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	// Check if db folder exist
	isExist, err := db.fs.IsFolderExist(db.root)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}
	err = db.fs.CreateAllFolder(db.root)
	if err != nil {
		return err
	}
	return nil
}

// Drop if you want to delete the fsentry repository you can use this method.
func (db *FSEntry) Drop() error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	// Check if db folder exist
	isExist, err := db.fs.IsFolderExist(db.root)
	if err != nil {
		return err
	}
	if !isExist {
		return nil
	}

	// Remove db folder
	err = db.fs.RemoveFolder(db.root)
	if err != nil {
		return err
	}

	return nil
}

// List allows you to get a list of objects (folders and entries) on the selected path.
func (db *FSEntry) List(path ...string) (*entity.List, error) {
	db.rwm.RLock()
	defer db.rwm.RUnlock()

	fullPath := db.serviceCommon.BuildPath("", path...)
	return db.fs.List(fullPath)
}

// Entry

func (db *FSEntry) CreateEntry(name string, data interface{}, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.serviceCommon.IsEntryNotExist(name, path...)
	if err != nil {
		return err
	}

	entry := entity.NewEntry(name, data, db.isPretty)
	err = db.fs.CreateEntry(fullPath, entry, db.isPretty)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) GetEntry(name string, path ...string) (*entity.Entry, error) {
	db.rwm.RLock()
	defer db.rwm.RUnlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := db.serviceCommon.IsEntryExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	entry, err := db.fs.GetEntry(fullPath)
	if err != nil {
		return nil, err
	}

	return entry, nil
}
func (db *FSEntry) MoveEntry(oldName, newName string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(oldName) == "" || utils.NameToID(newName) == "" {
		return fsentry_error.ErrorBadName
	}

	// Check if source entry exist
	fullOldPath, err := db.serviceCommon.IsEntryExist(oldName, path...)
	if err != nil {
		return err
	}

	// Read old entry
	entry, err := db.fs.GetEntry(fullOldPath)
	if err != nil {
		return err
	}

	entry.SetName(newName).UpdatedNow()

	var fullNewPath string
	// If entries have same ID
	if utils.NameToID(oldName) != utils.NameToID(newName) {
		// Check if destination entry not exist
		fullNewPath, err = db.serviceCommon.IsEntryNotExist(newName, path...)
		if err != nil {
			return err
		}
	} else {
		fullNewPath = fullOldPath
	}

	// Remove old entry
	err = db.fs.RemoveEntry(fullOldPath)
	if err != nil {
		return err
	}

	// Create new entry
	err = db.fs.CreateEntry(fullNewPath, entry, db.isPretty)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) UpdateEntry(name string, data interface{}, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.serviceCommon.IsEntryExist(name, path...)
	if err != nil {
		return err
	}

	// Get entry from file
	entry, err := db.fs.GetEntry(fullPath)
	if err != nil {
		return err
	}

	err = entry.UpdateData(data, db.isPretty)
	if err != nil {
		return err
	}

	// Update entry file
	err = db.fs.CreateEntry(fullPath, entry, db.isPretty)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) RemoveEntry(name string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.serviceCommon.IsEntryExist(name, path...)
	if err != nil {
		return err
	}

	err = db.fs.RemoveEntry(fullPath)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) DuplicateEntry(srcName, dstName string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(srcName) == "" || utils.NameToID(dstName) == "" {
		return fsentry_error.ErrorBadName
	}

	// Check if source entry exist
	fullSrcPath, err := db.serviceCommon.IsEntryExist(srcName, path...)
	if err != nil {
		return err
	}

	// Check if destination entry not exist
	fullDstPath, err := db.serviceCommon.IsEntryNotExist(dstName, path...)
	if err != nil {
		return err
	}

	// Get entry from file
	entry, err := db.fs.GetEntry(fullSrcPath)
	if err != nil {
		return err
	}

	entry.SetName(dstName).FlushTime()

	// Create entry file
	err = db.fs.CreateEntry(fullDstPath, entry, db.isPretty)
	if err != nil {
		return err
	}

	return nil
}

// Binary

func (db *FSEntry) CreateBinary(name string, data []byte, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.serviceCommon.IsBinaryNotExist(name, path...)
	if err != nil {
		return err
	}

	err = db.fs.CreateBinary(fullPath, data)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) GetBinary(name string, path ...string) ([]byte, error) {
	db.rwm.RLock()
	defer db.rwm.RUnlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := db.serviceCommon.IsBinaryExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get data from file
	data, err := db.fs.GetBinary(fullPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
func (db *FSEntry) MoveBinary(oldName, newName string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(oldName) == "" || utils.NameToID(newName) == "" {
		return fsentry_error.ErrorBadName
	}

	// Check if source binary exist
	fullOldPath, err := db.serviceCommon.IsBinaryExist(oldName, path...)
	if err != nil {
		return err
	}

	// Check if destination binary not exist
	fullNewPath, err := db.serviceCommon.IsBinaryNotExist(newName, path...)
	if err != nil {
		return err
	}

	// Rename binary
	err = db.fs.MoveObject(fullOldPath, fullNewPath)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) UpdateBinary(name string, data []byte, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.serviceCommon.IsBinaryExist(name, path...)
	if err != nil {
		return err
	}

	// Update binary file
	err = db.fs.CreateBinary(fullPath, data)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) RemoveBinary(name string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.serviceCommon.IsBinaryExist(name, path...)
	if err != nil {
		return err
	}

	err = db.fs.RemoveBinary(fullPath)
	if err != nil {
		return err
	}

	return nil
}
