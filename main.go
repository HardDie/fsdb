// fsentry
//
// Allows storing hierarchical data in files and folders on the file system with json descriptions and creation/update timestamps.
package fsentry

import (
	"path/filepath"
	"sync"

	"github.com/HardDie/fsentry/internal/entity"
	repFS "github.com/HardDie/fsentry/internal/repository/fs"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
	"github.com/HardDie/fsentry/pkg/fsentry_types"
)

type IFSEntry interface {
	Init() error
	Drop() error
	List(path ...string) (*entity.List, error)

	CreateFolder(name string, data interface{}, path ...string) (*entity.FolderInfo, error)
	GetFolder(name string, path ...string) (*entity.FolderInfo, error)
	MoveFolder(oldName, newName string, path ...string) (*entity.FolderInfo, error)
	UpdateFolder(name string, data interface{}, path ...string) (*entity.FolderInfo, error)
	RemoveFolder(name string, path ...string) error
	DuplicateFolder(srcName, dstName string, path ...string) (*entity.FolderInfo, error)
	UpdateFolderNameWithoutTimestamp(name, newName string, path ...string) error

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

	fs repFS.FS
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

	fullPath := db.buildPath("", path...)
	return db.fs.List(fullPath)
}

// CreateFolder you can use this method to create a folder within the repository.
// name - Name of the folder to be created.
// data - If you want to store some payload inside the json metadata you can pass it here.
// path - Optional value if you want to create a folder inside an existing folder. If you want to create a folder in the root of the storage, you can leave this value empty.
//
// As the result you receive JSON metadata file that was created in the created folder.
//
// Examples:
//
// Create a folder in the root of the storage:
//
//	resp, err := db.CreateFolder("f1", nil)
//	if err != nil {
//		panic(err)
//	}
//
// Create a folder inside an existing folder:
//
//	resp, err := db.CreateFolder("f2", nil, "f1")
//	if err != nil {
//		panic(err)
//	}
func (db *FSEntry) CreateFolder(name string, data interface{}, path ...string) (*entity.FolderInfo, error) {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := db.isFolderNotExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Create folder
	err = db.fs.CreateFolder(fullPath)
	if err != nil {
		return nil, err
	}

	// Create info file
	info := entity.NewFolderInfo(name, data, db.isPretty)
	err = db.fs.CreateInfo(fullPath, info, db.isPretty)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (db *FSEntry) GetFolder(name string, path ...string) (*entity.FolderInfo, error) {
	db.rwm.RLock()
	defer db.rwm.RUnlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := db.isFolderExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := db.fs.GetInfo(fullPath)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (db *FSEntry) MoveFolder(oldName, newName string, path ...string) (*entity.FolderInfo, error) {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(oldName) == "" || utils.NameToID(newName) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Check if source folder exist
	fullOldPath, err := db.isFolderExist(oldName, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := db.fs.GetInfo(fullOldPath)
	if err != nil {
		return nil, err
	}

	info.SetName(newName).UpdatedNow()

	// Update info file
	err = db.fs.CreateInfo(fullOldPath, info, db.isPretty)
	if err != nil {
		return nil, err
	}

	// If folders have same ID
	if utils.NameToID(oldName) == utils.NameToID(newName) {
		return info, nil
	}

	// Check if destination folder not exist
	fullNewPath, err := db.isFolderNotExist(newName, path...)
	if err != nil {
		return nil, err
	}

	// Rename folder
	err = db.fs.MoveObject(fullOldPath, fullNewPath)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (db *FSEntry) UpdateFolder(name string, data interface{}, path ...string) (*entity.FolderInfo, error) {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := db.isFolderExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := db.fs.GetInfo(fullPath)
	if err != nil {
		return nil, err
	}

	err = info.UpdateData(data, db.isPretty)
	if err != nil {
		return nil, err
	}

	// Update info file
	err = db.fs.CreateInfo(fullPath, info, db.isPretty)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (db *FSEntry) RemoveFolder(name string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.isFolderExist(name, path...)
	if err != nil {
		return err
	}

	err = db.fs.RemoveFolder(fullPath)
	if err != nil {
		return err
	}

	return nil
}
func (db *FSEntry) DuplicateFolder(srcName, dstName string, path ...string) (*entity.FolderInfo, error) {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(srcName) == "" || utils.NameToID(dstName) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Check if source folder exist
	fullSrcPath, err := db.isFolderExist(srcName, path...)
	if err != nil {
		return nil, err
	}

	// Check if destination folder not exist
	fullDstPath, err := db.isFolderNotExist(dstName, path...)
	if err != nil {
		return nil, err
	}

	// Copy folder
	err = db.fs.CopyFolder(fullSrcPath, fullDstPath)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := db.fs.GetInfo(fullDstPath)
	if err != nil {
		return nil, err
	}

	info.SetName(dstName).FlushTime()

	// Update info file
	err = db.fs.CreateInfo(fullDstPath, info, db.isPretty)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (db *FSEntry) UpdateFolderNameWithoutTimestamp(name, newName string, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" || utils.NameToID(newName) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.isFolderExist(name, path...)
	if err != nil {
		return err
	}

	// Get info from file
	info, err := db.fs.GetInfo(fullPath)
	if err != nil {
		return err
	}

	info.Id = utils.NameToID(newName)
	info.Name = fsentry_types.QuotedString(newName)

	// Update info file
	err = db.fs.CreateInfo(fullPath, info, db.isPretty)
	if err != nil {
		return err
	}

	return nil
}

// Entry

func (db *FSEntry) CreateEntry(name string, data interface{}, path ...string) error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := db.isEntryNotExist(name, path...)
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

	fullPath, err := db.isEntryExist(name, path...)
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
	fullOldPath, err := db.isEntryExist(oldName, path...)
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
		fullNewPath, err = db.isEntryNotExist(newName, path...)
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

	fullPath, err := db.isEntryExist(name, path...)
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

	fullPath, err := db.isEntryExist(name, path...)
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
	fullSrcPath, err := db.isEntryExist(srcName, path...)
	if err != nil {
		return err
	}

	// Check if destination entry not exist
	fullDstPath, err := db.isEntryNotExist(dstName, path...)
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

	fullPath, err := db.isBinaryNotExist(name, path...)
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

	fullPath, err := db.isBinaryExist(name, path...)
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
	fullOldPath, err := db.isBinaryExist(oldName, path...)
	if err != nil {
		return err
	}

	// Check if destination binary not exist
	fullNewPath, err := db.isBinaryNotExist(newName, path...)
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

	fullPath, err := db.isBinaryExist(name, path...)
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

	fullPath, err := db.isBinaryExist(name, path...)
	if err != nil {
		return err
	}

	err = db.fs.RemoveBinary(fullPath)
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
		return "", fsentry_error.ErrorBadName
	}

	fullPath := db.buildPath(id, path...)

	// Check if root folder exist
	isExist, err := db.fs.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorBadPath
	}

	// Check if destination folder exist
	isExist, err = db.fs.IsFolderExist(fullPath)
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorNotExist
	}

	return fullPath, nil
}
func (db *FSEntry) isFolderNotExist(name string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsentry_error.ErrorBadName
	}

	fullPath := db.buildPath(id, path...)

	// Check if root folder exist
	isExist, err := db.fs.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorBadPath
	}

	// Check if destination folder exist
	isExist, err = db.fs.IsFolderExist(fullPath)
	if err != nil {
		return "", err
	}
	if isExist {
		return "", fsentry_error.ErrorExist
	}

	return fullPath, nil
}
func (db *FSEntry) isEntryExist(name string, path ...string) (string, error) {
	return db.isFileExist(name, ".json", path...)
}
func (db *FSEntry) isEntryNotExist(name string, path ...string) (string, error) {
	return db.isFileNotExist(name, ".json", path...)
}
func (db *FSEntry) isBinaryExist(name string, path ...string) (string, error) {
	return db.isFileExist(name, ".bin", path...)
}
func (db *FSEntry) isBinaryNotExist(name string, path ...string) (string, error) {
	return db.isFileNotExist(name, ".bin", path...)
}

func (db *FSEntry) isFileExist(name, ext string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsentry_error.ErrorBadName
	}
	id += ext

	fullPath := db.buildPath(id, path...)

	// Check if root folder exist
	isExist, err := db.fs.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorBadPath
	}

	// Check if destination entry exist
	isExist, err = db.fs.IsFileExist(fullPath)
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorNotExist
	}

	return fullPath, nil
}
func (db *FSEntry) isFileNotExist(name, ext string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsentry_error.ErrorBadName
	}
	id += ext

	fullPath := db.buildPath(id, path...)

	// Check if root folder exist
	isExist, err := db.fs.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorBadPath
	}

	// Check if destination entry exist
	isExist, err = db.fs.IsFileExist(fullPath)
	if err != nil {
		return "", err
	}
	if isExist {
		return "", fsentry_error.ErrorExist
	}

	return fullPath, nil
}
