package fs

import (
	"errors"
	"io"
	iofs "io/fs"
	"log"
	"os"

	"github.com/otiai10/copy"

	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

const (
	CreateDirPerm   = 0755
	CreateFileFlags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	CreateFilePerm  = 0666
)

type FS interface {
	CreateFile(path string, data []byte) error
	ReadFile(path string) ([]byte, error)
	RemoveFile(path string) error
	CreateFolder(path string) error
	CreateAllFolder(path string) error
	RemoveFolder(path string) error
	Rename(oldPath, newPath string) error
	CopyFolder(srcPath, dstPath string) error
	List(path string) ([]os.FileInfo, error)
	IsFileExist(path string) (isExist bool, err error)
	IsFolderExist(path string) (isExist bool, err error)
}

type fs struct{}

func NewFS() FS {
	return fs{}
}

// CreateFile allows you to create a file and fill it with some binary data.
func (r fs) CreateFile(path string, data []byte) error {
	file, err := os.OpenFile(path, CreateFileFlags, CreateFilePerm)
	if err != nil {
		if e := isKnownError(err); e != nil {
			return e
		}
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer func() {
		if err = file.Sync(); err != nil {
			log.Printf("CreateFile(): error sync file %q: %s", path, err.Error())
		}
		if err = file.Close(); err != nil {
			log.Printf("CreateFile(): error close file %q: %s", path, err.Error())
		}
	}()

	n, err := file.Write(data)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	if n != len(data) {
		log.Printf("CreateFile(): the size of input and written data is different. Received: %d, written: %d", len(data), n)
	}
	return nil
}

// ReadFile attempts to open and read all binary data from the desired file.
func (r fs) ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		if e := isKnownError(err); e != nil {
			return nil, e
		}
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer func() {
		var err error
		if err = file.Close(); err != nil {
			log.Printf("ReadFile(): error close file %q: %s", path, err.Error())
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		// TODO: process different types of errors
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return data, nil
}

// RemoveFile allows you to delete a file or an empty folder.
func (r fs) RemoveFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// CreateFolder allows you to create a folder in the file system.
func (r fs) CreateFolder(path string) error {
	err := os.Mkdir(path, CreateDirPerm)
	if err != nil {
		if e := isKnownError(err); e != nil {
			return e
		}
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// CreateAllFolder allows you to create a folder in the file system,
// and if some intermediate folders in the desired path do not exist, they will also be created.
func (r fs) CreateAllFolder(path string) error {
	err := os.MkdirAll(path, CreateDirPerm)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// RemoveFolder will delete the desired folder even if it is not empty with all the data it contains.
func (r fs) RemoveFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// Rename allows you to rename a file/directory or move it to another path.
func (r fs) Rename(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// CopyFolder will recursively copy the source folder to the desired destination path.
func (r fs) CopyFolder(srcPath, dstPath string) error {
	err := copy.Copy(srcPath, dstPath)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// List will read the complete list of objects on the specified path and return them.
func (r fs) List(path string) ([]os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		// TODO: process different types of errors
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer f.Close()

	files, err := f.Readdir(0)
	if err != nil {
		// TODO: process different types of errors
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return files, nil
}

// IsFileExist checks if an object that is a file, not a folder, exists at the specified path.
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

// IsFolderExist checks if an object that is a folder, exists at the specified path.
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

func isKnownError(err error) error {
	var pathErr *iofs.PathError
	if errors.As(err, &pathErr) {
		switch {
		case os.IsExist(pathErr):
			return fsentry_error.Wrap(err, fsentry_error.ErrorExist)
		case os.IsNotExist(pathErr):
			return fsentry_error.Wrap(err, fsentry_error.ErrorNotExist)
		case os.IsPermission(pathErr):
			return fsentry_error.Wrap(err, fsentry_error.ErrorPermissions)
		}
	}
	return nil
}
