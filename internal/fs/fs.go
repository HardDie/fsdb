package fs

import (
	"errors"
	"os"
)

type FS interface {
	CreateFile(path string, data []byte) error
	ReadFile(path string) ([]byte, error)
	UpdateFile(path string, data []byte) error
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

var (
	ErrorBadPath        = errors.New("bad path")
	ErrorExist          = errors.New("file or folder already exist")
	ErrorNotExist       = errors.New("file or folder not exist")
	ErrorPermission     = errors.New("not enough permissions")
	ErrorFileExist      = errors.New("file already exist")
	ErrorFileNotExist   = errors.New("file not exist")
	ErrorIsFile         = errors.New("it's a file")
	ErrorFolderNotEmpty = errors.New("folder not empty")
	ErrorFolderExist    = errors.New("folder already exist")
	ErrorIsDir          = errors.New("it's a directory")
	ErrorInternal       = errors.New("internal")
)
