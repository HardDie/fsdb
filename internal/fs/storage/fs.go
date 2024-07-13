package storage

import (
	"errors"
	"io"
	iofs "io/fs"
	"log"
	"os"
	"runtime"
	"syscall"

	"github.com/otiai10/copy"

	"github.com/HardDie/fsentry/internal/fs"
)

const (
	CreateDirPerm   = 0755
	CreateFileFlags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	UpdateFileFlags = os.O_WRONLY | os.O_TRUNC
	CreateFilePerm  = 0666
)

type FS struct{}

func New() FS {
	return FS{}
}

// CreateFile allows you to create a file and fill it with some binary data.
func (r FS) CreateFile(path string, data []byte) error {
	file, err := os.OpenFile(path, CreateFileFlags, CreateFilePerm)
	if err != nil {
		var pathErr *iofs.PathError
		if errors.As(err, &pathErr) {
			switch {
			case os.IsExist(pathErr):
				return fs.ErrorFileExist
			case os.IsNotExist(pathErr):
				return fs.ErrorBadPath
			case os.IsPermission(pathErr):
				return fs.ErrorPermission
			}
		}
		var syscallErr syscall.Errno
		if errors.As(err, &syscallErr) {
			if runtime.GOOS == "windows" {
				switch uintptr(syscallErr) {
				case 536870954:
					return fs.ErrorFileExist
				}
			}
		}
		log.Printf("fs.CreateFile() os.OpenFile: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	defer func() {
		if err = file.Sync(); err != nil {
			log.Printf("fs.CreateFile(): error sync file %q: %s", path, err.Error())
		}
		if err = file.Close(); err != nil {
			log.Printf("fs.CreateFile(): error close file %q: %s", path, err.Error())
		}
	}()

	n, err := file.Write(data)
	if err != nil {
		// TODO: process different types of errors
		log.Printf("fs.CreateFile() file.Write: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	if n != len(data) {
		log.Printf("fs.CreateFile(): the size of input and written data is different. Received: %d, written: %d", len(data), n)
	}
	return nil
}

// UpdateFile allows you to update a file.
func (r FS) UpdateFile(path string, data []byte) error {
	file, err := os.OpenFile(path, UpdateFileFlags, CreateFilePerm)
	if err != nil {
		var pathErr *iofs.PathError
		if errors.As(err, &pathErr) {
			switch {
			case os.IsNotExist(pathErr):
				return fs.ErrorFileNotExist
			case os.IsPermission(pathErr):
				return fs.ErrorPermission
			}
		}
		var syscallErr syscall.Errno
		if errors.As(err, &syscallErr) {
			switch uintptr(syscallErr) {
			case 21:
				return fs.ErrorFileNotExist
			}
			if runtime.GOOS == "windows" {
				switch uintptr(syscallErr) {
				case 536870954:
					return fs.ErrorFileNotExist
				}
			}
		}
		log.Printf("fs.UpdateFile() os.OpenFile: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	defer func() {
		if err = file.Sync(); err != nil {
			log.Printf("fs.UpdateFile(): error sync file %q: %s", path, err.Error())
		}
		if err = file.Close(); err != nil {
			log.Printf("fs.UpdateFile(): error close file %q: %s", path, err.Error())
		}
	}()

	n, err := file.Write(data)
	if err != nil {
		// TODO: process different types of errors
		log.Printf("fs.UpdateFile() file.Write: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	if n != len(data) {
		log.Printf("fs.UpdateFile(): the size of input and written data is different. Received: %d, written: %d", len(data), n)
	}
	return nil
}

// ReadFile attempts to open and read all binary data from the desired file.
func (r FS) ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		var pathErr *iofs.PathError
		if errors.As(err, &pathErr) {
			switch {
			case os.IsNotExist(pathErr):
				return nil, fs.ErrorFileNotExist
			case os.IsPermission(pathErr):
				return nil, fs.ErrorPermission
			}
		}
		log.Printf("fs.ReadFile() os.OpenFile: %T %s", err, err.Error())
		return nil, fs.ErrorInternal
	}
	defer func() {
		var err error
		if err = file.Close(); err != nil {
			log.Printf("fs.ReadFile(): error close file %q: %s", path, err.Error())
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		var syscallErr syscall.Errno
		if errors.As(err, &syscallErr) {
			switch uintptr(syscallErr) {
			case 21:
				return nil, fs.ErrorFileNotExist
			}
			if runtime.GOOS == "windows" {
				switch uintptr(syscallErr) {
				case 1:
					return nil, fs.ErrorFileNotExist
				}
			}
		}
		log.Printf("fs.ReadFile() io.ReadAll: %T %s", err, err.Error())
		return nil, fs.ErrorInternal
	}
	return data, nil
}

// RemoveFile allows you to delete a file or an empty folder.
func (r FS) RemoveFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		var pathErr *iofs.PathError
		if errors.As(err, &pathErr) {
			switch {
			case os.IsExist(pathErr):
				return fs.ErrorFolderNotEmpty
			case os.IsNotExist(pathErr):
				return fs.ErrorNotExist
			case os.IsPermission(pathErr):
				return fs.ErrorPermission
			}
		}
		log.Printf("fs.RemoveFile() os.Remove: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	return nil
}

// CreateFolder allows you to create a folder in the file system.
func (r FS) CreateFolder(path string) error {
	err := os.Mkdir(path, CreateDirPerm)
	if err != nil {
		var pathErr *iofs.PathError
		if errors.As(err, &pathErr) {
			switch {
			case os.IsExist(pathErr):
				return fs.ErrorFolderExist
			case os.IsNotExist(pathErr):
				return fs.ErrorBadPath
			case os.IsPermission(pathErr):
				return fs.ErrorPermission
			}
		}
		log.Printf("fs.CreateFolder() os.Mkdir: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	return nil
}

// CreateAllFolder allows you to create a folder in the file system,
// and if some intermediate folders in the desired path do not exist, they will also be created.
// If the specified folder already exists, there will be no error.
func (r FS) CreateAllFolder(path string) error {
	err := os.MkdirAll(path, CreateDirPerm)
	if err != nil {
		var pathErr *iofs.PathError
		if errors.As(err, &pathErr) {
			switch {
			case os.IsNotExist(pathErr):
				return fs.ErrorFolderExist
			case os.IsPermission(pathErr):
				return fs.ErrorPermission
			}
		}
		var syscallErr syscall.Errno
		if errors.As(err, &syscallErr) {
			switch uintptr(syscallErr) {
			case 20:
				return fs.ErrorFolderExist
			}
		}
		log.Printf("fs.CreateAllFolder() os.MkdirAll: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	return nil
}

// RemoveFolder will delete the desired folder even if it is not empty with all the data it contains.
func (r FS) RemoveFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		var pathErr *iofs.PathError
		if errors.As(err, &pathErr) {
			switch {
			case os.IsPermission(pathErr):
				return fs.ErrorPermission
			}
		}
		log.Printf("fs.RemoveFolder() os.RemoveAll: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	return nil
}

// Rename allows you to rename a file/directory or move it to another path.
func (r FS) Rename(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		// TODO: process different types of errors
		log.Printf("fs.Rename() os.Rename: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	return nil
}

// CopyFolder will recursively copy the source folder to the desired destination path.
func (r FS) CopyFolder(srcPath, dstPath string) error {
	err := copy.Copy(srcPath, dstPath)
	if err != nil {
		// TODO: process different types of errors
		log.Printf("fs.CopyFolder() copy.Copy: %T %s", err, err.Error())
		return fs.ErrorInternal
	}
	return nil
}

// List will read the complete list of objects on the specified path and return them.
func (r FS) List(path string) ([]os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		// TODO: process different types of errors
		log.Printf("fs.List() os.Open: %T %s", err, err.Error())
		return nil, fs.ErrorInternal
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Printf("fs.List(): error close file %q: %s", path, err.Error())
		}
	}()

	files, err := f.Readdir(0)
	if err != nil {
		// TODO: process different types of errors
		log.Printf("fs.List() f.Readdir: %T %s", err, err.Error())
		return nil, fs.ErrorInternal
	}
	return files, nil
}

// IsFileExist checks if an object that is a file, not a folder, exists at the specified path.
func (r FS) IsFileExist(path string) (isExist bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// entry not exist
			return false, nil
		}
		// other error
		log.Printf("fs.IsFileExist() os.Stat: %T %s", err, err.Error())
		return false, fs.ErrorInternal
	}

	// check if it is not a folder
	if stat.IsDir() {
		return false, fs.ErrorIsDir
	}

	// entry exists
	return true, nil
}

// IsFolderExist checks if an object that is a folder, exists at the specified path.
func (r FS) IsFolderExist(path string) (isExist bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// folder not exist
			return false, nil
		}
		// other error
		log.Printf("fs.IsFolderExist() os.Stat: %T %s", err, err.Error())
		return false, fs.ErrorInternal
	}

	// check if it is a folder
	if !stat.IsDir() {
		return false, fs.ErrorIsFile
	}

	// folder exists
	return true, nil
}
