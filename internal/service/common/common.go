package common

import (
	"path/filepath"

	repositoryEntry "github.com/HardDie/fsentry/internal/repository/entry"
	repositoryFolder "github.com/HardDie/fsentry/internal/repository/folder"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

type Common interface {
	BuildPath(id string, path ...string) string
	IsFolderExist(name string, path ...string) (string, error)
	IsFolderNotExist(name string, path ...string) (string, error)
	IsEntryExist(name string, path ...string) (string, error)
	IsEntryNotExist(name string, path ...string) (string, error)
	IsBinaryExist(name string, path ...string) (string, error)
	IsBinaryNotExist(name string, path ...string) (string, error)
	IsFileExist(name, ext string, path ...string) (string, error)
	IsFileNotExist(name, ext string, path ...string) (string, error)
}

type common struct {
	root      string
	repFolder repositoryFolder.Folder
	repEntry  repositoryEntry.Entry
}

func NewCommon(
	root string,
	repFolder repositoryFolder.Folder,
	repEntry repositoryEntry.Entry,
) Common {
	return common{
		root:      root,
		repFolder: repFolder,
		repEntry:  repEntry,
	}
}

func (s common) BuildPath(id string, path ...string) string {
	pathSlice := append([]string{s.root}, path...)
	return filepath.Join(append(pathSlice, id)...)
}
func (s common) IsFolderExist(name string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsentry_error.ErrorBadName
	}

	fullPath := s.BuildPath(id, path...)

	// Check if root folder exist
	isExist, err := s.repFolder.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorBadPath
	}

	// Check if destination folder exist
	isExist, err = s.repFolder.IsFolderExist(fullPath)
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorNotExist
	}

	return fullPath, nil
}
func (s common) IsFolderNotExist(name string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsentry_error.ErrorBadName
	}

	fullPath := s.BuildPath(id, path...)

	// Check if root folder exist
	isExist, err := s.repFolder.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorBadPath
	}

	// Check if destination folder exist
	isExist, err = s.repFolder.IsFolderExist(fullPath)
	if err != nil {
		return "", err
	}
	if isExist {
		return "", fsentry_error.ErrorExist
	}

	return fullPath, nil
}
func (s common) IsEntryExist(name string, path ...string) (string, error) {
	return s.IsFileExist(name, ".json", path...)
}
func (s common) IsEntryNotExist(name string, path ...string) (string, error) {
	return s.IsFileNotExist(name, ".json", path...)
}
func (s common) IsBinaryExist(name string, path ...string) (string, error) {
	return s.IsFileExist(name, ".bin", path...)
}
func (s common) IsBinaryNotExist(name string, path ...string) (string, error) {
	return s.IsFileNotExist(name, ".bin", path...)
}
func (s common) IsFileExist(name, ext string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsentry_error.ErrorBadName
	}
	id += ext

	fullPath := s.BuildPath(id, path...)

	// Check if root folder exist
	isExist, err := s.repFolder.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorBadPath
	}

	// Check if destination entry exist
	isExist, err = s.repEntry.IsFileExist(fullPath)
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorNotExist
	}

	return fullPath, nil
}
func (s common) IsFileNotExist(name, ext string, path ...string) (string, error) {
	id := utils.NameToID(name)
	if id == "" {
		return "", fsentry_error.ErrorBadName
	}
	id += ext

	fullPath := s.BuildPath(id, path...)

	// Check if root folder exist
	isExist, err := s.repFolder.IsFolderExist(filepath.Dir(fullPath))
	if err != nil {
		return "", err
	}
	if !isExist {
		return "", fsentry_error.ErrorBadPath
	}

	// Check if destination entry exist
	isExist, err = s.repEntry.IsFileExist(fullPath)
	if err != nil {
		return "", err
	}
	if isExist {
		return "", fsentry_error.ErrorExist
	}

	return fullPath, nil
}
