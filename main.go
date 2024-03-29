// fsentry
//
// Allows storing hierarchical data in files and folders on the file system
// with json descriptions and creation/update timestamps.
package fsentry

import (
	"sync"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/internal/fs"
	fsStorage "github.com/HardDie/fsentry/internal/fs/storage"
	repositoryBinary "github.com/HardDie/fsentry/internal/repository/binary"
	repositoryEntry "github.com/HardDie/fsentry/internal/repository/entry"
	repositoryFolder "github.com/HardDie/fsentry/internal/repository/folder"
	serviceBinary "github.com/HardDie/fsentry/internal/service/binary"
	serviceCommon "github.com/HardDie/fsentry/internal/service/common"
	serviceEntry "github.com/HardDie/fsentry/internal/service/entry"
	serviceFolder "github.com/HardDie/fsentry/internal/service/folder"
)

var (
	// validate interface.
	_ IFSEntry = &FSEntry{}
)

type IFSEntry interface {
	Init() error
	Drop() error
	List(path ...string) (*entity.List, error)

	serviceFolder.Folder
	serviceEntry.Entry
	serviceBinary.Binary
}
type FSEntry struct {
	root string
	rwm  sync.RWMutex

	isPretty bool

	repFS         fs.FS
	repFolder     repositoryFolder.Folder
	repEntry      repositoryEntry.Entry
	repBinary     repositoryBinary.Binary
	serviceCommon serviceCommon.Common
	serviceFolder.Folder
	serviceEntry.Entry
	serviceBinary.Binary
}

func WithPretty() func(fs *FSEntry) {
	return func(fs *FSEntry) {
		fs.isPretty = true
	}
}

func NewFSEntry(root string, ops ...func(fs *FSEntry)) IFSEntry {
	res := &FSEntry{
		root:  root,
		repFS: fsStorage.NewFS(),
	}
	for _, op := range ops {
		op(res)
	}
	res.repFolder = repositoryFolder.NewFolder(res.repFS)
	res.repEntry = repositoryEntry.NewEntry(res.repFS)
	res.repBinary = repositoryBinary.NewBinary(res.repFS)

	res.serviceCommon = serviceCommon.NewCommon(res.root, res.repFolder, res.repEntry)
	res.Folder = serviceFolder.NewFolder(res.root, &res.rwm, res.isPretty, res.repFolder, res.repEntry, res.serviceCommon)
	res.Entry = serviceEntry.NewEntry(res.root, &res.rwm, res.isPretty, res.repEntry, res.serviceCommon)
	res.Binary = serviceBinary.NewBinary(res.root, &res.rwm, res.isPretty, res.repEntry, res.repBinary, res.serviceCommon)
	return res
}

// Basic

// Init check if a repository folder has been created and if not, create one.
func (db *FSEntry) Init() error {
	db.rwm.Lock()
	defer db.rwm.Unlock()

	// Check if db folder exist
	isExist, err := db.repFolder.IsFolderExist(db.root)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}
	err = db.repFolder.CreateAllFolder(db.root)
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
	isExist, err := db.repFolder.IsFolderExist(db.root)
	if err != nil {
		return err
	}
	if !isExist {
		return nil
	}

	// Remove db folder
	err = db.repFolder.RemoveFolder(db.root)
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
	return db.repFolder.List(fullPath)
}
