package fs

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

const (
	DirPerm  = 0755
	InfoFile = ".info.json"
)

type FS interface {
	IsFolderExist(path string) (isExist bool, err error)
	CreateFolder(path string) error
	CreateAllFolder(path string) error
	CopyFolder(srcPath, dstPath string) error
	RemoveFolder(path string) error
	List(path string) (*entity.List, error)

	CreateEntry(path string, entry *entity.Entry, isIndent bool) error
	GetEntry(path string) (*entity.Entry, error)
	RemoveEntry(path string) error
	CreateBinary(path string, data []byte) error
	GetBinary(path string) ([]byte, error)
	RemoveBinary(path string) error
	IsFileExist(path string) (isExist bool, err error)
	MoveObject(oldPath, newPath string) error
	CreateInfo(path string, data *entity.FolderInfo, isIndent bool) error
	GetInfo(path string) (*entity.FolderInfo, error)
}

type fs struct {
}

func NewFS() FS {
	return fs{}
}

func (r fs) IsFolderExist(path string) (isExist bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// folder not exist
			return false, nil
		}
		// other error
		return false, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}

	// check if it is a folder
	if !stat.IsDir() {
		return false, fsentry_error.ErrorBadPath
	}

	// folder exists
	return true, nil
}
func (r fs) CreateFolder(path string) error {
	err := os.Mkdir(path, DirPerm)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r fs) CreateAllFolder(path string) error {
	err := os.MkdirAll(path, DirPerm)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r fs) CopyFolder(srcPath, dstPath string) error {
	err := copy.Copy(srcPath, dstPath)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r fs) RemoveFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

func (r fs) List(path string) (*entity.List, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer f.Close()

	files, err := f.Readdir(0)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}

	res := &entity.List{}
	for _, file := range files {
		name := file.Name()

		if name[0] == '.' {
			// skip hidden files
			continue
		}

		if file.IsDir() {
			res.Folders = append(res.Folders, file.Name())
		} else if filepath.Ext(name) == ".json" {
			res.Entries = append(res.Entries, name[0:len(name)-5])
		}
	}

	return res, nil
}

func (r fs) CreateEntry(path string, entry *entity.Entry, isIndent bool) error {
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
func (r fs) GetEntry(path string) (*entity.Entry, error) {
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
func (r fs) RemoveEntry(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r fs) CreateBinary(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer func() {
		if err = file.Sync(); err != nil {
			log.Printf("CreateBinary(): error sync file %q: %s", path, err.Error())
		}
		if err = file.Close(); err != nil {
			log.Printf("CreateBinary(): error close file %q: %s", path, err.Error())
		}
	}()

	_, err = file.Write(data)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r fs) GetBinary(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return data, nil
}
func (r fs) RemoveBinary(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r fs) IsFileExist(path string) (isExist bool, err error) {
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
func (r fs) MoveObject(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r fs) CreateInfo(path string, data *entity.FolderInfo, isIndent bool) error {
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
func (r fs) GetInfo(path string) (*entity.FolderInfo, error) {
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
