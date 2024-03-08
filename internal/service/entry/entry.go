package entry

import (
	"sync"

	"github.com/HardDie/fsentry/internal/entity"
	repositoryEntry "github.com/HardDie/fsentry/internal/repository/entry"
	serviceCommon "github.com/HardDie/fsentry/internal/service/common"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

type Entry interface {
	CreateEntry(name string, data interface{}, path ...string) error
	GetEntry(name string, path ...string) (*entity.Entry, error)
	MoveEntry(oldName, newName string, path ...string) error
	UpdateEntry(name string, data interface{}, path ...string) error
	RemoveEntry(name string, path ...string) error
	DuplicateEntry(srcName, dstName string, path ...string) error
}

type entry struct {
	root string
	rwm  *sync.RWMutex

	isPretty bool

	repEntry repositoryEntry.Entry
	common   serviceCommon.Common
}

func NewEntry(
	root string,
	rwm *sync.RWMutex,
	isPretty bool,
	repEntry repositoryEntry.Entry,
	common serviceCommon.Common,
) Entry {
	return &entry{
		root:     root,
		rwm:      rwm,
		isPretty: isPretty,
		repEntry: repEntry,
		common:   common,
	}
}

func (s *entry) CreateEntry(name string, data interface{}, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsEntryNotExist(name, path...)
	if err != nil {
		return err
	}

	ent := entity.NewEntry(name, data, s.isPretty)
	err = s.repEntry.CreateEntry(fullPath, ent, s.isPretty)
	if err != nil {
		return err
	}

	return nil
}
func (s *entry) GetEntry(name string, path ...string) (*entity.Entry, error) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsEntryExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	ent, err := s.repEntry.GetEntry(fullPath)
	if err != nil {
		return nil, err
	}

	return ent, nil
}
func (s *entry) MoveEntry(oldName, newName string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(oldName) == "" || utils.NameToID(newName) == "" {
		return fsentry_error.ErrorBadName
	}

	// Check if source entry exist
	fullOldPath, err := s.common.IsEntryExist(oldName, path...)
	if err != nil {
		return err
	}

	// Read old entry
	ent, err := s.repEntry.GetEntry(fullOldPath)
	if err != nil {
		return err
	}

	ent.SetName(newName).UpdatedNow()

	var fullNewPath string
	// If entries have same ID
	if utils.NameToID(oldName) != utils.NameToID(newName) {
		// Check if destination entry not exist
		fullNewPath, err = s.common.IsEntryNotExist(newName, path...)
		if err != nil {
			return err
		}
	} else {
		fullNewPath = fullOldPath
	}

	// Remove old entry
	err = s.repEntry.RemoveEntry(fullOldPath)
	if err != nil {
		return err
	}

	// Create new entry
	err = s.repEntry.CreateEntry(fullNewPath, ent, s.isPretty)
	if err != nil {
		return err
	}

	return nil
}
func (s *entry) UpdateEntry(name string, data interface{}, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsEntryExist(name, path...)
	if err != nil {
		return err
	}

	// Get entry from file
	ent, err := s.repEntry.GetEntry(fullPath)
	if err != nil {
		return err
	}

	err = ent.UpdateData(data, s.isPretty)
	if err != nil {
		return err
	}

	// Update entry file
	err = s.repEntry.CreateEntry(fullPath, ent, s.isPretty)
	if err != nil {
		return err
	}

	return nil
}
func (s *entry) RemoveEntry(name string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsEntryExist(name, path...)
	if err != nil {
		return err
	}

	err = s.repEntry.RemoveEntry(fullPath)
	if err != nil {
		return err
	}

	return nil
}
func (s *entry) DuplicateEntry(srcName, dstName string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(srcName) == "" || utils.NameToID(dstName) == "" {
		return fsentry_error.ErrorBadName
	}

	// Check if source entry exist
	fullSrcPath, err := s.common.IsEntryExist(srcName, path...)
	if err != nil {
		return err
	}

	// Check if destination entry not exist
	fullDstPath, err := s.common.IsEntryNotExist(dstName, path...)
	if err != nil {
		return err
	}

	// Get entry from file
	ent, err := s.repEntry.GetEntry(fullSrcPath)
	if err != nil {
		return err
	}

	ent.SetName(dstName).FlushTime()

	// Create entry file
	err = s.repEntry.CreateEntry(fullDstPath, ent, s.isPretty)
	if err != nil {
		return err
	}

	return nil
}
