package entry

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

const (
	InfoFile = ".info.json"
)

type Entry interface {
	CreateEntry(path string, entry *entity.Entry, isIndent bool) error
	GetEntry(path string) (*entity.Entry, error)
	RemoveEntry(path string) error
	IsFileExist(path string) (isExist bool, err error)
	MoveObject(oldPath, newPath string) error
	CreateInfo(path string, data *entity.FolderInfo, isIndent bool) error
	GetInfo(path string) (*entity.FolderInfo, error)
}

type entry struct{}

func NewEntry() Entry {
	return entry{}
}

func (r entry) CreateEntry(path string, entry *entity.Entry, isIndent bool) error {
	file, err := os.Create(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer func() {
		if err = file.Sync(); err != nil {
			log.Printf("CreateEntry(): error sync file %q: %s", path, err.Error())
		}
		if err = file.Close(); err != nil {
			log.Printf("CreateEntry(): error close file %q: %s", path, err.Error())
		}
	}()

	enc := json.NewEncoder(file)
	if isIndent {
		enc.SetIndent("", "	")
	}
	err = enc.Encode(entry)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r entry) GetEntry(path string) (*entity.Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer file.Close()

	info := &entity.Entry{}
	err = json.NewDecoder(file).Decode(info)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return info, nil
}
func (r entry) RemoveEntry(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r entry) IsFileExist(path string) (isExist bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// entry not exist
			return false, nil
		}
		// other error
		return false, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}

	// check if it is not a folder
	if stat.IsDir() {
		return false, fsentry_error.ErrorBadPath
	}

	// entry exists
	return true, nil
}
func (r entry) MoveObject(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r entry) CreateInfo(path string, data *entity.FolderInfo, isIndent bool) error {
	file, err := os.Create(filepath.Join(path, InfoFile))
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer func() {
		var err error
		if err = file.Sync(); err != nil {
			log.Printf("CreateInfo(): error sync file %q: %s", path, err.Error())
		}
		if err = file.Close(); err != nil {
			log.Printf("CreateInfo(): error close file %q: %s", path, err.Error())
		}
	}()

	enc := json.NewEncoder(file)
	if isIndent {
		enc.SetIndent("", "	")
	}
	err = enc.Encode(data)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r entry) GetInfo(path string) (*entity.FolderInfo, error) {
	file, err := os.Open(filepath.Join(path, InfoFile))
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer file.Close()

	info := &entity.FolderInfo{}
	err = json.NewDecoder(file).Decode(info)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return info, nil
}
