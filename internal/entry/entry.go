package entry

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/HardDie/fsentry/dto"
	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
	"github.com/HardDie/fsentry/pkg/fsentry_types"
)

const (
	entryFileSuffix = ".json"
)

type InternalEntry struct {
	ID        string                     `json:"id"`
	Name      fsentry_types.QuotedString `json:"name"`
	CreatedAt *time.Time                 `json:"createdAt"`
	UpdatedAt *time.Time                 `json:"updatedAt"`
	Data      json.RawMessage            `json:"data"`
}

func toInternalEntry(ext dto.Entry) InternalEntry {
	return InternalEntry{
		ID:        ext.ID,
		Name:      fsentry_types.QS(ext.Name),
		CreatedAt: &ext.CreatedAt,
		UpdatedAt: &ext.UpdatedAt,
		Data:      ext.Data,
	}
}
func toExternalEntry(in InternalEntry) dto.Entry {
	ext := dto.Entry{
		ID:   in.ID,
		Name: in.Name.String(),
		Data: in.Data,
	}
	if in.CreatedAt == nil {
		now := time.Now().UTC()
		in.CreatedAt = &now
	}
	ext.CreatedAt = *in.CreatedAt
	if in.UpdatedAt != nil {
		in.UpdatedAt = in.CreatedAt
	}
	ext.UpdatedAt = *in.UpdatedAt
	return ext
}

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

func (s Service) Create(path, name string, data interface{}) (*dto.Entry, error) {
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
func (s Service) Get(path, name string) (*dto.Entry, error) {
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
	inEntry, err := utils.JSONToStruct[InternalEntry](data)
	if err != nil {
		return nil, err
	}
	if inEntry == nil {
		return nil, fsentry_error.ErrorInternal
	}

	extEntry := toExternalEntry(*inEntry)
	return &extEntry, nil
}
func (s Service) Move(path, oldName, newName string) (*dto.Entry, error) {
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
	oldInEnt, err := utils.JSONToStruct[InternalEntry](data)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	newInEnt := InternalEntry{
		ID:        newID,
		Name:      fsentry_types.QS(newName),
		CreatedAt: oldInEnt.CreatedAt,
		UpdatedAt: &now,
		Data:      oldInEnt.Data,
	}

	newEntJSON, err := utils.StructToJSON(newInEnt, s.isPretty)
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
		newExtEnt := toExternalEntry(newInEnt)
		return &newExtEnt, nil
	}

	// If our attempt to update the data file fails, we assume the data file has an old value,
	// in which case we must rename it to the old name to keep the entry valid.
	err = s.fs.Rename(newFullPath, oldFullPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (s Service) Update(path, name string, data interface{}) (*dto.Entry, error) {
	oldExtEnt, err := s.Get(path, name)
	if err != nil {
		return nil, err
	}
	if oldExtEnt == nil {
		return nil, fsentry_error.ErrorInternal
	}
	fullPath := filepath.Join(path, oldExtEnt.ID+entryFileSuffix)

	// Prepare a custom payload and convert it to a json byte slice.
	dataJSON, err := utils.StructToJSON(data, s.isPretty)
	if err != nil {
		return nil, err
	}

	inEnt := toInternalEntry(*oldExtEnt)
	inEnt.Data = dataJSON
	inEnt.UpdatedAt = utils.Allocate(s.now().UTC())

	entJSON, err := utils.StructToJSON(inEnt, s.isPretty)
	if err != nil {
		return nil, err
	}

	err = s.fs.UpdateFile(fullPath, entJSON)
	if err != nil {
		return nil, err
	}

	newExtEnt := toExternalEntry(inEnt)
	return &newExtEnt, nil
}
func (s Service) Remove(path, name string) error {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id+entryFileSuffix)

	return s.fs.RemoveFile(fullPath)
}
func (s Service) Duplicate(path, oldName, newName string) (*dto.Entry, error) {
	oldExtEnt, err := s.Get(path, oldName)
	if err != nil {
		return nil, err
	}

	// Check if the new entry name is a valid entry name.
	newID := utils.NameToID(newName)
	if newID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	newFullPath := filepath.Join(path, newID+entryFileSuffix)

	return s.createRaw(newFullPath, newName, newID, oldExtEnt.Data)
}

func (s Service) createRaw(fullPath, name, id string, dataJSON json.RawMessage) (*dto.Entry, error) {
	// Creating and filling in information about a new entry.
	now := s.now().UTC()
	inEntry := InternalEntry{
		ID:        id,
		Name:      fsentry_types.QS(name),
		CreatedAt: &now,
		UpdatedAt: &now,
		Data:      dataJSON,
	}

	// Prepare the new entry and convert it into a json byte slice.
	entJSON, err := utils.StructToJSON(inEntry, s.isPretty)
	if err != nil {
		return nil, err
	}

	err = s.fs.CreateFile(fullPath, entJSON)
	if err != nil {
		return nil, err
	}

	extEntry := toExternalEntry(inEntry)
	return &extEntry, nil
}
