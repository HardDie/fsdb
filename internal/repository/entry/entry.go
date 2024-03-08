package entry

import (
	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/internal/repository/common"
	repositoryFS "github.com/HardDie/fsentry/internal/repository/fs"
)

type Entry interface {
	CreateEntry(path string, entry *entity.Entry, isIndent bool) error
	UpdateEntry(path string, entry *entity.Entry, isIndent bool) error
	GetEntry(path string) (*entity.Entry, error)
	RemoveEntry(path string) error
	IsFileExist(path string) (isExist bool, err error)
	MoveObject(oldPath, newPath string) error
}

type entry struct {
	fs repositoryFS.FS
}

func NewEntry(fs repositoryFS.FS) Entry {
	return entry{
		fs: fs,
	}
}

// CreateEntry allows you to create a json file with default metadata such as created_at and updated_at,
// as well as a custom payload. Entry is the main object for storing information in the fsentry library.
func (r entry) CreateEntry(path string, entry *entity.Entry, isIndent bool) error {
	data, err := common.DataToJSON(entry, isIndent)
	if err != nil {
		return err
	}
	err = r.fs.CreateFile(path, data)
	if err != nil {
		return err
	}
	return nil
}

// UpdateEntry allows you to update an existing Entry json file.
func (r entry) UpdateEntry(path string, entry *entity.Entry, isIndent bool) error {
	data, err := common.DataToJSON(entry, isIndent)
	if err != nil {
		return err
	}
	err = r.fs.UpdateFile(path, data)
	if err != nil {
		return err
	}
	return nil
}

// GetEntry attempts to read the specified path from the file system and parse it as an Entry object.
func (r entry) GetEntry(path string) (*entity.Entry, error) {
	data, err := r.fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	info, err := common.JSONToData[entity.Entry](data)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// RemoveEntry allows you to remove an Entry object from a folder.
// If you pass the path to another object, it should return an error.
func (r entry) RemoveEntry(path string) error {
	return r.fs.RemoveFile(path)
}

// IsFileExist checks if an object that is a file, not a folder, exists at the specified path.
func (r entry) IsFileExist(path string) (isExist bool, err error) {
	return r.fs.IsFileExist(path)
}

// MoveObject allows you to rename an Entry object or move it to a different path.
func (r entry) MoveObject(oldPath, newPath string) error {
	return r.fs.Rename(oldPath, newPath)
}
