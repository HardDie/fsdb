package folder

import (
	"path/filepath"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/repository/common"
)

const (
	InfoFile = ".info.json"
)

type Folder interface {
	CreateFolder(path string) error
	CreateAllFolder(path string) error
	CopyFolder(srcPath, dstPath string) error
	RemoveFolder(path string) error
	List(path string) (*entity.List, error)
	CreateInfo(path string, data *entity.FolderInfo, isIndent bool) error
	UpdateInfo(path string, info *entity.FolderInfo, isIndent bool) error
	GetInfo(path string) (*entity.FolderInfo, error)
	IsFolderExist(path string) (isExist bool, err error)
}

type folder struct {
	fs fs.FS
}

func NewFolder(fs fs.FS) Folder {
	return folder{
		fs: fs,
	}
}

// CreateFolder allows you to create a folder in the file system.
func (r folder) CreateFolder(path string) error {
	return r.fs.CreateFolder(path)
}

// CreateAllFolder allows you to create a folder in the file system,
// and if some intermediate folders in the desired path do not exist, they will also be created.
func (r folder) CreateAllFolder(path string) error {
	return r.fs.CreateAllFolder(path)
}

// CopyFolder will recursively copy the source folder to the desired destination path.
func (r folder) CopyFolder(srcPath, dstPath string) error {
	return r.fs.CopyFolder(srcPath, dstPath)
}

// RemoveFolder will delete the desired folder even if it is not empty with all the data it contains.
func (r folder) RemoveFolder(path string) error {
	return r.fs.RemoveFolder(path)
}

// List will return two separate slices with folders in the specified directory and entries.
func (r folder) List(path string) (*entity.List, error) {
	files, err := r.fs.List(path)
	if err != nil {
		return nil, err
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

// CreateInfo allows you to create an .info.json file with metadata for the Folder object.
func (r folder) CreateInfo(path string, info *entity.FolderInfo, isIndent bool) error {
	data, err := common.DataToJSON(info, isIndent)
	if err != nil {
		return err
	}
	err = r.fs.CreateFile(filepath.Join(path, InfoFile), data)
	if err != nil {
		return err
	}
	return nil
}

// UpdateInfo allows you to update an existing .info.json file with metadata for a Folder object.
func (r folder) UpdateInfo(path string, info *entity.FolderInfo, isIndent bool) error {
	data, err := common.DataToJSON(info, isIndent)
	if err != nil {
		return err
	}
	err = r.fs.UpdateFile(filepath.Join(path, InfoFile), data)
	if err != nil {
		return err
	}
	return nil
}

// GetInfo reads metadata about the Folder object. If .info.json does not exist inside a folder,
// that folder will not be considered a Folder object in fsentry terms.
func (r folder) GetInfo(path string) (*entity.FolderInfo, error) {
	data, err := r.fs.ReadFile(filepath.Join(path, InfoFile))
	if err != nil {
		return nil, err
	}
	info, err := common.JSONToData[entity.FolderInfo](data)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// IsFolderExist checks if an object that is a folder, exists at the specified path.
func (r folder) IsFolderExist(path string) (isExist bool, err error) {
	return r.fs.IsFolderExist(path)
}
