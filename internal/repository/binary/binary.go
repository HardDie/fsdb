package binary

import (
	"io"
	"log"
	"os"

	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

type Binary interface {
	CreateBinary(path string, data []byte) error
	GetBinary(path string) ([]byte, error)
	RemoveBinary(path string) error
}

type binary struct {
}

func NewBinary() Binary {
	return binary{}
}

func (r binary) CreateBinary(path string, data []byte) error {
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
func (r binary) GetBinary(path string) ([]byte, error) {
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
func (r binary) RemoveBinary(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}
