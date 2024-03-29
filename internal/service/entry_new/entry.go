package entry_new

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
	"github.com/HardDie/fsentry/pkg/fsentry_types"
)

const (
	entryFileSuffix = ".json"
)

type Entry interface {
	Create(path, name string, data interface{}) (*entity.Entry, error)
	Get(path, name string) (*entity.Entry, error)
	Move(path, oldName, newName string) (*entity.Entry, error)
	Update(path, name string, data interface{}) (*entity.Entry, error)
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) (*entity.Entry, error)
}

type entry struct {
	fs       fs.FS
	isPretty bool
	now      func() time.Time
}

func NewEntry(
	fs fs.FS,
	isPretty bool,
) Entry {
	return entry{
		fs:       fs,
		isPretty: isPretty,
		now:      time.Now,
	}
}

func (s entry) Create(path, name string, data interface{}) (*entity.Entry, error) {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Prepare a custom payload and convert it to a json byte slice.
	dataJSON, err := utils.StructToJSON(data, s.isPretty)
	if err != nil {
		return nil, err
	}

	// fullPath is the path where a new entry with this name will be created.
	fullPath := filepath.Join(path, id+entryFileSuffix)

	return s.createRaw(fullPath, name, id, dataJSON)
}
func (s entry) Get(path, name string) (*entity.Entry, error) {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+entryFileSuffix)

	data, err := s.fs.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	ent, err := utils.JSONToStruct[entity.Entry](data)
	if err != nil {
		return nil, err
	}

	return ent, nil
}
func (s entry) Move(path, oldName, newName string) (*entity.Entry, error) {
	// Check if the old entry name is a valid entry name.
	oldID := utils.NameToID(oldName)
	if oldID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Check if the new entry name is a valid entry name.
	newID := utils.NameToID(newName)
	if newID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// newFullPath - path to the new entry to which the old one will be moved.
	newFullPath := filepath.Join(path, newID+entryFileSuffix)

	// Check if the name of the new entry is not occupied by an existing entry.
	isExist, err := s.fs.IsFileExist(newFullPath)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, fsentry_error.ErrorExist
	}

	oldFullPath := filepath.Join(path, oldID+entryFileSuffix)

	// Read meta info from the current folder to update it.
	data, err := s.fs.ReadFile(oldFullPath)
	if err != nil {
		return nil, err
	}
	oldEnt, err := utils.JSONToStruct[entity.Entry](data)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	newEnt := entity.Entry{
		Id:        newID,
		Name:      fsentry_types.QS(newName),
		CreatedAt: oldEnt.CreatedAt,
		UpdatedAt: &now,
		Data:      oldEnt.Data,
	}

	newEntJSON, err := utils.StructToJSON(newEnt, s.isPretty)
	if err != nil {
		return nil, err
	}

	// The operation of renaming a entry is cheaper and faster than updating file data,
	// so we will first try moving the old entry to the new name.
	err = s.fs.Rename(oldFullPath, newFullPath)
	if err != nil {
		return nil, err
	}

	// If the entry has been successfully renamed, we attempt to update the data.
	err = s.fs.UpdateFile(newFullPath, newEntJSON)
	if err == nil {
		// Good. Returns information about the renamed entry.
		return &newEnt, nil
	}

	// If our attempt to update the data file fails, we assume the data file has an old value,
	// in which case we must rename it to the old name to keep the entry valid.
	err = s.fs.Rename(newFullPath, oldFullPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (s entry) Update(path, name string, data interface{}) (*entity.Entry, error) {
	ent, err := s.Get(path, name)
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(path, ent.Id+entryFileSuffix)

	// Prepare a custom payload and convert it to a json byte slice.
	dataJSON, err := utils.StructToJSON(data, s.isPretty)
	if err != nil {
		return nil, err
	}

	ent.Data = dataJSON
	ent.UpdatedAt = utils.Allocate(s.now().UTC())

	entJSON, err := utils.StructToJSON(ent, s.isPretty)
	if err != nil {
		return nil, err
	}

	err = s.fs.UpdateFile(fullPath, entJSON)
	if err != nil {
		return nil, err
	}

	return ent, nil
}
func (s entry) Remove(path, name string) error {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+entryFileSuffix)

	return s.fs.RemoveFile(fullPath)
}
func (s entry) Duplicate(path, oldName, newName string) (*entity.Entry, error) {
	oldEnt, err := s.Get(path, oldName)
	if err != nil {
		return nil, err
	}

	// Check if the new entry name is a valid entry name.
	newID := utils.NameToID(newName)
	if newID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	newFullPath := filepath.Join(path, newID+entryFileSuffix)

	return s.createRaw(newFullPath, newName, newID, oldEnt.Data)
}

func (s entry) createRaw(fullPath, name, id string, dataJSON json.RawMessage) (*entity.Entry, error) {
	// Creating and filling in information about a new entry.
	now := s.now().UTC()
	ent := entity.Entry{
		Id:        id,
		Name:      fsentry_types.QS(name),
		CreatedAt: &now,
		UpdatedAt: &now,
		Data:      dataJSON,
	}

	// Prepare the new entry and convert it into a json byte slice.
	entJSON, err := utils.StructToJSON(ent, s.isPretty)
	if err != nil {
		return nil, err
	}

	err = s.fs.CreateFile(fullPath, entJSON)
	if err != nil {
		return nil, err
	}

	return &ent, nil
}
