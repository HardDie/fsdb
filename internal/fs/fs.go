package fs

import "os"

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
