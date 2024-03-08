package binary

import (
	repositoryFS "github.com/HardDie/fsentry/internal/repository/fs"
)

type Binary interface {
	CreateBinary(path string, data []byte) error
	GetBinary(path string) ([]byte, error)
	RemoveBinary(path string) error
}

type binary struct {
	fs repositoryFS.FS
}

func NewBinary(fs repositoryFS.FS) Binary {
	return binary{
		fs: fs,
	}
}

// CreateBinary allows you to create a *.bin file at the specified path.
func (r binary) CreateBinary(path string, data []byte) error {
	return r.fs.CreateFile(path, data)
}

// GetBinary checks if the file can be accessed, reads all the contents from it and returns it.
func (r binary) GetBinary(path string) ([]byte, error) {
	return r.fs.ReadFile(path)
}

// RemoveBinary allows you to delete a binary file.
func (r binary) RemoveBinary(path string) error {
	return r.fs.RemoveFile(path)
}
