package folder

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

const (
	DirPerm = 0755
)

type Folder interface {
	IsFolderExist(path string) (isExist bool, err error)
	CreateFolder(path string) error
	CreateAllFolder(path string) error
	CopyFolder(srcPath, dstPath string) error
	RemoveFolder(path string) error
	List(path string) (*entity.List, error)
}

type folder struct{}

func NewFolder() Folder {
	return folder{}
}

func (r folder) IsFolderExist(path string) (isExist bool, err error) {
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
func (r folder) CreateFolder(path string) error {
	err := os.Mkdir(path, DirPerm)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r folder) CreateAllFolder(path string) error {
	err := os.MkdirAll(path, DirPerm)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r folder) CopyFolder(srcPath, dstPath string) error {
	err := copy.Copy(srcPath, dstPath)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
func (r folder) RemoveFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

func (r folder) List(path string) (*entity.List, error) {
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
