package fsutils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/internal/entry_error"
)

const (
	DirPerm  = 0755
	InfoFile = ".info.json"
)

func IsFolderExist(path string) (isExist bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// folder not exist
			return false, nil
		}
		// other error
		return false, entry_error.Wrap(err, entry_error.ErrorInternal)
	}

	// check if it is a folder
	if !stat.IsDir() {
		return false, entry_error.ErrorBadPath
	}

	// folder exists
	return true, nil
}
func CreateFolder(path string) error {
	err := os.Mkdir(path, DirPerm)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return nil
}
func CreateAllFolder(path string) error {
	err := os.MkdirAll(path, DirPerm)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return nil
}
func MoveFolder(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return nil
}
func CopyFolder(srcPath, dstPath string) error {
	err := copy.Copy(srcPath, dstPath)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return nil
}
func RemoveFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return nil
}

func IsEntryExist(path string) (isExist bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// entry not exist
			return false, nil
		}
		// other error
		return false, entry_error.Wrap(err, entry_error.ErrorInternal)
	}

	// check if it is not a folder
	if stat.IsDir() {
		return false, entry_error.ErrorBadPath
	}

	// entry exists
	return true, nil
}
func CreateEntry(path string, entry *entity.Entry) error {
	file, err := os.Create(path)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}

	err = json.NewEncoder(file).Encode(entry)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return nil
}
func GetEntry(path string) (*entity.Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, entry_error.Wrap(err, entry_error.ErrorInternal)
	}

	info := &entity.Entry{}
	err = json.NewDecoder(file).Decode(info)
	if err != nil {
		return nil, entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return info, nil
}
func RemoveEntry(path string) error {
	err := os.Remove(path)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return nil
}

func CreateInfo(path string, data *entity.FolderInfo) error {
	file, err := os.Create(filepath.Join(path, InfoFile))
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return nil
}
func GetInfo(path string) (*entity.FolderInfo, error) {
	file, err := os.Open(filepath.Join(path, InfoFile))
	if err != nil {
		return nil, entry_error.Wrap(err, entry_error.ErrorInternal)
	}

	info := &entity.FolderInfo{}
	err = json.NewDecoder(file).Decode(info)
	if err != nil {
		return nil, entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	return info, nil
}
