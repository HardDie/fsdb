package folder_new

import (
	"encoding/json"
	"log"
	"path/filepath"
	"time"

	"github.com/HardDie/fsentry/internal/entity"
	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
	"github.com/HardDie/fsentry/pkg/fsentry_types"
)

const (
	infoFile = ".info.json"
)

type Folder interface {
	Create(path, name string, data interface{}) (*entity.FolderInfo, error)
	Get(path, name string) (*entity.FolderInfo, error)
	Move(path, oldName, newName string) (*entity.FolderInfo, error)
	Update(path, name string, data interface{}) (*entity.FolderInfo, error)
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) (*entity.FolderInfo, error)
	MoveWithoutTimestamp(path, oldName, newName string) (*entity.FolderInfo, error)
}

type folder struct {
	fs       fs.FS
	isPretty bool
	now      func() time.Time
}

func NewFolder(
	fs fs.FS,
	isPretty bool,
) Folder {
	return folder{
		fs:       fs,
		isPretty: isPretty,
		now:      time.Now,
	}
}

func (s folder) Create(path, name string, data interface{}) (*entity.FolderInfo, error) {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Prepare a custom payload and convert it to a json byte slice.
	dataJSON, err := utils.StructToJSON(data, s.isPretty)
	if err != nil {
		return nil, err
	}

	// Creating and filling in information about a new folder.
	now := s.now().UTC()
	info := entity.FolderInfo{
		Id:        id,
		Name:      fsentry_types.QS(name),
		CreatedAt: &now,
		UpdatedAt: &now,
		Data:      dataJSON,
	}

	// Prepare the new folder information and convert it into a json byte slice.
	infoJSON, err := utils.StructToJSON(info, s.isPretty)
	if err != nil {
		return nil, err
	}

	// fullPath is the path where a new folder with this name will be created.
	fullPath := filepath.Join(path, id)
	// infoFilePath - path to the .info.json file with meta information that will be created in the new folder.
	infoFilePath := filepath.Join(fullPath, infoFile)

	// Try creating an empty folder. It must be created inside an existing folder.
	err = s.fs.CreateFolder(fullPath)
	if err != nil {
		return nil, err
	}

	// Try creating the .info.json file in a new folder.
	err = s.fs.CreateFile(infoFilePath, infoJSON)
	if err == nil {
		// Good. Returns information about the created folder.
		return &info, nil
	}

	// If something went wrong and the .info.json file was not created,
	// it means that the folder you just created is corrupted, try deleting it.
	e := s.fs.RemoveFolder(fullPath)
	if e != nil {
		log.Printf("error remove folder %q after error create info: %q", e.Error(), err.Error())
	}
	return nil, err
}
func (s folder) Get(path, name string) (*entity.FolderInfo, error) {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id)

	return s.getInfo(fullPath)
}
func (s folder) Move(path, oldName, newName string) (*entity.FolderInfo, error) {
	// Check if the old folder name is a valid folder name.
	oldID := utils.NameToID(oldName)
	if oldID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Check if the new folder name is a valid folder name.
	newID := utils.NameToID(newName)
	if newID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// newFullPath - path to the new folder to which the old one will be moved.
	newFullPath := filepath.Join(path, newID)

	// Check if the name of the new folder is not occupied by an existing folder.
	isExist, err := s.fs.IsFolderExist(newFullPath)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, fsentry_error.ErrorExist
	}

	oldFullPath := filepath.Join(path, oldID)

	// Read meta info from the current folder to update it.
	oldInfo, err := s.getInfo(oldFullPath)
	if err != nil {
		return nil, err
	}

	newInfoFilePath := filepath.Join(newFullPath, infoFile)

	now := s.now().UTC()
	newInfo := entity.FolderInfo{
		Id:        newID,
		Name:      fsentry_types.QS(newName),
		CreatedAt: oldInfo.CreatedAt,
		UpdatedAt: &now,
		Data:      oldInfo.Data,
	}

	newInfoJSON, err := utils.StructToJSON(newInfo, s.isPretty)
	if err != nil {
		return nil, err
	}

	// The operation of renaming a folder is cheaper and faster than updating file data,
	// so we will first try moving the old folder to the new name.
	err = s.fs.Rename(oldFullPath, newFullPath)
	if err != nil {
		return nil, err
	}

	// If the folder has been successfully renamed, we attempt to update the meta info about the folder.
	err = s.fs.UpdateFile(newInfoFilePath, newInfoJSON)
	if err == nil {
		// Good. Returns information about the renamed folder.
		return &newInfo, nil
	}

	// If our attempt to update the meta info file fails, we assume the meta info file has an old value,
	// in which case we must rename it to the old name to keep the folder valid.
	err = s.fs.Rename(newFullPath, oldFullPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (s folder) Update(path, name string, data interface{}) (*entity.FolderInfo, error) {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Prepare a custom payload and convert it to a json byte slice.
	dataJSON, err := utils.StructToJSON(data, s.isPretty)
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(path, id)

	info, err := s.updateInfo(fullPath, entity.UpdateFolderInfo{
		Data:      utils.Allocate[json.RawMessage](dataJSON),
		UpdatedAt: utils.Allocate(s.now().UTC()),
	})
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (s folder) Remove(path, name string) error {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id)

	isExist, err := s.fs.IsFolderExist(fullPath)
	if err != nil {
		return err
	}
	if !isExist {
		return fsentry_error.ErrorNotExist
	}

	isExist, err = s.isInfoExist(fullPath)
	if err != nil {
		return err
	}
	if !isExist {
		return fsentry_error.ErrorFolderCorrupted
	}

	return s.fs.RemoveFolder(fullPath)
}
func (s folder) Duplicate(path, oldName, newName string) (*entity.FolderInfo, error) {
	// Check if the old folder name is a valid folder name.
	oldID := utils.NameToID(oldName)
	if oldID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Check if the new folder name is a valid folder name.
	newID := utils.NameToID(newName)
	if newID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// newFullPath - path to the new folder to which the old one will be moved.
	newFullPath := filepath.Join(path, newID)

	// Check if the name of the new folder is not occupied by an existing folder.
	isExist, err := s.fs.IsFolderExist(newFullPath)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, fsentry_error.ErrorExist
	}

	oldFullPath := filepath.Join(path, oldID)

	// Read meta info from the current folder to update it.
	oldInfo, err := s.getInfo(oldFullPath)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	newInfo := entity.FolderInfo{
		Id:        newID,
		Name:      fsentry_types.QS(newName),
		CreatedAt: &now,
		UpdatedAt: &now,
		Data:      oldInfo.Data,
	}

	newInfoJSON, err := utils.StructToJSON(newInfo, s.isPretty)
	if err != nil {
		return nil, err
	}

	newInfoFilePath := filepath.Join(newFullPath, infoFile)

	err = s.fs.CopyFolder(oldFullPath, newFullPath)
	if err != nil {
		// Clean up if attempt was unsuccessful
		if e := s.fs.RemoveFolder(newFullPath); e != nil {
			log.Printf("error remove invalid folder %q after error copy: %q", newFullPath, e.Error())
		}
		return nil, err
	}

	err = s.fs.UpdateFile(newInfoFilePath, newInfoJSON)
	if err != nil {
		// Clean up if attempt was unsuccessful
		if e := s.fs.RemoveFolder(newFullPath); e != nil {
			log.Printf("error remove invalid folder %q after error info update: %q", newFullPath, e.Error())
		}
		return nil, err
	}
	return &newInfo, nil
}
func (s folder) MoveWithoutTimestamp(path, oldName, newName string) (*entity.FolderInfo, error) {
	// Check if the old folder name is a valid folder name.
	oldID := utils.NameToID(oldName)
	if oldID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Check if the new folder name is a valid folder name.
	newID := utils.NameToID(newName)
	if newID == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// newFullPath - path to the new folder to which the old one will be moved.
	newFullPath := filepath.Join(path, newID)

	// Check if the name of the new folder is not occupied by an existing folder.
	isExist, err := s.fs.IsFolderExist(newFullPath)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, fsentry_error.ErrorExist
	}

	oldFullPath := filepath.Join(path, oldID)

	// Read meta info from the current folder to update it.
	oldInfo, err := s.getInfo(oldFullPath)
	if err != nil {
		return nil, err
	}

	newInfoFilePath := filepath.Join(newFullPath, infoFile)

	newInfo := entity.FolderInfo{
		Id:        newID,
		Name:      fsentry_types.QS(newName),
		CreatedAt: oldInfo.CreatedAt,
		UpdatedAt: oldInfo.UpdatedAt,
		Data:      oldInfo.Data,
	}

	newInfoJSON, err := utils.StructToJSON(newInfo, s.isPretty)
	if err != nil {
		return nil, err
	}

	// The operation of renaming a folder is cheaper and faster than updating file data,
	// so we will first try moving the old folder to the new name.
	err = s.fs.Rename(oldFullPath, newFullPath)
	if err != nil {
		return nil, err
	}

	// If the folder has been successfully renamed, we attempt to update the meta info about the folder.
	err = s.fs.UpdateFile(newInfoFilePath, newInfoJSON)
	if err == nil {
		// Good. Returns information about the renamed folder.
		return &newInfo, nil
	}

	// If our attempt to update the meta info file fails, we assume the meta info file has an old value,
	// in which case we must rename it to the old name to keep the folder valid.
	err = s.fs.Rename(newFullPath, oldFullPath)
	if err != nil {
		return nil, err
	}

	return oldInfo, nil
}

func (s folder) getInfo(fullPath string) (*entity.FolderInfo, error) {
	infoFilePath := filepath.Join(fullPath, infoFile)

	// If the folder exists, we will try to read information about the folder.
	data, err := s.fs.ReadFile(infoFilePath)
	if err != nil {
		return nil, err
	}
	info, err := utils.JSONToStruct[entity.FolderInfo](data)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (s folder) updateInfo(fullPath string, req entity.UpdateFolderInfo) (*entity.FolderInfo, error) {
	info, err := s.getInfo(fullPath)
	if err != nil {
		return nil, err
	}

	if req.ID != nil {
		info.Id = *req.ID
	}
	if req.Name != nil {
		info.Name = fsentry_types.QS(*req.Name)
	}
	if req.CreatedAt != nil {
		info.CreatedAt = req.CreatedAt
	}
	if req.UpdatedAt != nil {
		info.UpdatedAt = req.UpdatedAt
	}
	if req.Data != nil {
		info.Data = *req.Data
	}

	infoJSON, err := utils.StructToJSON(info, s.isPretty)
	if err != nil {
		return nil, err
	}

	infoFilePath := filepath.Join(fullPath, infoFile)

	err = s.fs.UpdateFile(infoFilePath, infoJSON)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (s folder) isInfoExist(fullPath string) (bool, error) {
	infoFilePath := filepath.Join(fullPath, infoFile)
	return s.fs.IsFileExist(infoFilePath)
}
