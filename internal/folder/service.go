package folder

import (
	"encoding/json"
	"log"
	"path/filepath"
	"time"

	"github.com/HardDie/fsentry/dto"
	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
	"github.com/HardDie/fsentry/pkg/fsentry_types"
)

const (
	infoFileSuffix = ".info.json"
)

type InternalInfo struct {
	ID        string                     `json:"id"`
	Name      fsentry_types.QuotedString `json:"name"`
	CreatedAt *time.Time                 `json:"createdAt"`
	UpdatedAt *time.Time                 `json:"updatedAt"`
	Data      json.RawMessage            `json:"data"`
}
type UpdateInfoRequest struct {
	ID        *string          `json:"id"`
	Name      *string          `json:"name"`
	CreatedAt *time.Time       `json:"createdAt"`
	UpdatedAt *time.Time       `json:"updatedAt"`
	Data      *json.RawMessage `json:"data"`
}

func toInternalInfo(ext dto.FolderInfo) InternalInfo {
	return InternalInfo{
		ID:        ext.ID,
		Name:      fsentry_types.QS(ext.Name),
		CreatedAt: &ext.CreatedAt,
		UpdatedAt: &ext.UpdatedAt,
		Data:      ext.Data,
	}
}
func toExternalInfo(in InternalInfo) dto.FolderInfo {
	ext := dto.FolderInfo{
		ID:   in.ID,
		Name: in.Name.String(),
		Data: in.Data,
	}
	if in.CreatedAt == nil {
		now := time.Now().UTC()
		in.CreatedAt = &now
	}
	ext.CreatedAt = *in.CreatedAt
	if in.UpdatedAt != nil {
		in.UpdatedAt = in.CreatedAt
	}
	ext.UpdatedAt = *in.UpdatedAt
	return ext
}

type Service struct {
	fs       fs.FS
	isPretty bool
	now      func() time.Time
}

func New(
	fs fs.FS,
	isPretty bool,
) Service {
	return Service{
		fs:       fs,
		isPretty: isPretty,
		now:      time.Now,
	}
}

func (s Service) Create(path, name string, data interface{}) (*dto.FolderInfo, error) {
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
	inInfo := InternalInfo{
		ID:        id,
		Name:      fsentry_types.QS(name),
		CreatedAt: &now,
		UpdatedAt: &now,
		Data:      dataJSON,
	}

	// Prepare the new folder information and convert it into a json byte slice.
	infoJSON, err := utils.StructToJSON(inInfo, s.isPretty)
	if err != nil {
		return nil, err
	}

	// fullPath is the path where a new folder with this name will be created.
	fullPath := filepath.Join(path, id)
	// infoFilePath - path to the .info.json file with meta information that will be created in the new folder.
	infoFilePath := filepath.Join(fullPath, infoFileSuffix)

	// Try creating an empty folder. It must be created inside an existing folder.
	err = s.fs.CreateFolder(fullPath)
	if err != nil {
		return nil, err
	}

	// Try creating the .info.json file in a new folder.
	err = s.fs.CreateFile(infoFilePath, infoJSON)
	if err == nil {
		// Good. Returns information about the created folder.
		extInfo := toExternalInfo(inInfo)
		return &extInfo, nil
	}

	// If something went wrong and the .info.json file was not created,
	// it means that the folder you just created is corrupted, try deleting it.
	e := s.fs.RemoveFolder(fullPath)
	if e != nil {
		log.Printf("error remove folder %q after error create info: %q", e.Error(), err.Error())
	}
	return nil, err
}
func (s Service) Get(path, name string) (*dto.FolderInfo, error) {
	// Check if it is possible to translate a name into a valid ID.
	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath := filepath.Join(path, id)

	return s.getInfo(fullPath)
}
func (s Service) Move(path, oldName, newName string) (*dto.FolderInfo, error) {
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
	oldExtInfo, err := s.getInfo(oldFullPath)
	if err != nil {
		return nil, err
	}

	newInfoFilePath := filepath.Join(newFullPath, infoFileSuffix)

	now := s.now().UTC()
	newInInfo := InternalInfo{
		ID:        newID,
		Name:      fsentry_types.QS(newName),
		CreatedAt: &oldExtInfo.CreatedAt,
		UpdatedAt: &now,
		Data:      oldExtInfo.Data,
	}

	newInfoJSON, err := utils.StructToJSON(newInInfo, s.isPretty)
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
		newExtInfo := toExternalInfo(newInInfo)
		return &newExtInfo, nil
	}

	// If our attempt to update the meta info file fails, we assume the meta info file has an old value,
	// in which case we must rename it to the old name to keep the folder valid.
	err = s.fs.Rename(newFullPath, oldFullPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (s Service) Update(path, name string, data interface{}) (*dto.FolderInfo, error) {
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

	extInfo, err := s.updateInfo(fullPath, UpdateInfoRequest{
		Data:      utils.Allocate[json.RawMessage](dataJSON),
		UpdatedAt: utils.Allocate(s.now().UTC()),
	})
	if err != nil {
		return nil, err
	}

	return extInfo, nil
}
func (s Service) Remove(path, name string) error {
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
func (s Service) Duplicate(path, oldName, newName string) (*dto.FolderInfo, error) {
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
	oldExtInfo, err := s.getInfo(oldFullPath)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	newInInfo := InternalInfo{
		ID:        newID,
		Name:      fsentry_types.QS(newName),
		CreatedAt: &now,
		UpdatedAt: &now,
		Data:      oldExtInfo.Data,
	}

	newInfoJSON, err := utils.StructToJSON(newInInfo, s.isPretty)
	if err != nil {
		return nil, err
	}

	newInfoFilePath := filepath.Join(newFullPath, infoFileSuffix)

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
	newExtInfo := toExternalInfo(newInInfo)
	return &newExtInfo, nil
}
func (s Service) MoveWithoutTimestamp(path, oldName, newName string) (*dto.FolderInfo, error) {
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
	oldExtInfo, err := s.getInfo(oldFullPath)
	if err != nil {
		return nil, err
	}

	newInfoFilePath := filepath.Join(newFullPath, infoFileSuffix)

	newInInfo := InternalInfo{
		ID:        newID,
		Name:      fsentry_types.QS(newName),
		CreatedAt: &oldExtInfo.CreatedAt,
		UpdatedAt: &oldExtInfo.UpdatedAt,
		Data:      oldExtInfo.Data,
	}

	newInfoJSON, err := utils.StructToJSON(newInInfo, s.isPretty)
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
		newExtInfo := toExternalInfo(newInInfo)
		return &newExtInfo, nil
	}

	// If our attempt to update the meta info file fails, we assume the meta info file has an old value,
	// in which case we must rename it to the old name to keep the folder valid.
	err = s.fs.Rename(newFullPath, oldFullPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s Service) getInfo(fullPath string) (*dto.FolderInfo, error) {
	infoFilePath := filepath.Join(fullPath, infoFileSuffix)

	// If the folder exists, we will try to read information about the folder.
	data, err := s.fs.ReadFile(infoFilePath)
	if err != nil {
		return nil, err
	}
	inInfo, err := utils.JSONToStruct[InternalInfo](data)
	if err != nil {
		return nil, err
	}
	if inInfo == nil {
		log.Println("inInfo is nil")
		return nil, fsentry_error.ErrorInternal
	}

	extInfo := toExternalInfo(*inInfo)
	return &extInfo, nil
}
func (s Service) updateInfo(fullPath string, req UpdateInfoRequest) (*dto.FolderInfo, error) {
	oldExtInfo, err := s.getInfo(fullPath)
	if err != nil {
		return nil, err
	}
	if oldExtInfo == nil {
		log.Println("extInfo is nil")
		return nil, fsentry_error.ErrorInternal
	}
	inInfo := toInternalInfo(*oldExtInfo)

	if req.ID != nil {
		inInfo.ID = *req.ID
	}
	if req.Name != nil {
		inInfo.Name = fsentry_types.QS(*req.Name)
	}
	if req.CreatedAt != nil {
		inInfo.CreatedAt = req.CreatedAt
	}
	if req.UpdatedAt != nil {
		inInfo.UpdatedAt = req.UpdatedAt
	}
	if req.Data != nil {
		inInfo.Data = *req.Data
	}

	infoJSON, err := utils.StructToJSON(inInfo, s.isPretty)
	if err != nil {
		return nil, err
	}

	infoFilePath := filepath.Join(fullPath, infoFileSuffix)

	err = s.fs.UpdateFile(infoFilePath, infoJSON)
	if err != nil {
		return nil, err
	}

	newExtInfo := toExternalInfo(inInfo)
	return &newExtInfo, nil
}
func (s Service) isInfoExist(fullPath string) (bool, error) {
	infoFilePath := filepath.Join(fullPath, infoFileSuffix)
	return s.fs.IsFileExist(infoFilePath)
}
