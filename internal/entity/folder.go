package entity

import (
	"encoding/json"
	"time"

	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
	"github.com/HardDie/fsentry/pkg/fsentry_types"
)

// FolderInfo is an .info.json file inside each folder created by the fsentry library.
// The main purpose of this file is to store the original name of the folder,
// to keep track of the timestamps of the create and update events.
// But FolderInfo, like Entry files, has the ability to store custom payloads.
type FolderInfo struct {
	// Id is a name, but it has all special characters removed, all spaces replaced with underscores,
	// and is shortened to 200 characters because some file systems prohibit files from having long names.
	Id string `json:"id"`
	// Name is the original name that was set by the user without any modification.
	Name fsentry_types.QuotedString `json:"name"`
	// CreatedAt metadata for each Entry to track the original creation date.
	CreatedAt *time.Time `json:"createdAt"`
	// UpdatedAt metadata to keep track of when the Entry was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`
	// Data is a custom json payload for custom data.
	Data json.RawMessage `json:"data"`
}

type UpdateFolderInfo struct {
	ID        *string                     `json:"id"`
	Name      *fsentry_types.QuotedString `json:"name"`
	CreatedAt *time.Time                  `json:"createdAt"`
	UpdatedAt *time.Time                  `json:"updatedAt"`
	Data      *json.RawMessage            `json:"data"`
}

func NewFolderInfo(id, name string, data interface{}, isIndent bool) *FolderInfo {
	var dataJson []byte
	if isIndent {
		dataJson, _ = json.MarshalIndent(data, "", "	")
	} else {
		dataJson, _ = json.Marshal(data)
	}
	now := time.Now().UTC()
	return &FolderInfo{
		Id:        id,
		Name:      fsentry_types.QuotedString(name),
		CreatedAt: &now,
		UpdatedAt: &now,
		Data:      dataJson,
	}
}

func (i *FolderInfo) SetName(name string) *FolderInfo {
	i.Name = fsentry_types.QuotedString(name)
	i.Id = utils.NameToID(name)
	return i
}
func (i *FolderInfo) UpdatedNow() *FolderInfo {
	i.UpdatedAt = utils.Allocate(time.Now().UTC())
	return i
}
func (i *FolderInfo) FlushTime() *FolderInfo {
	i.CreatedAt = utils.Allocate(time.Now().UTC())
	i.UpdatedAt = nil
	return i
}
func (i *FolderInfo) UpdateData(data interface{}, isIndent bool) error {
	var dataJson []byte
	var err error
	if isIndent {
		dataJson, err = json.MarshalIndent(data, "", "	")
	} else {
		dataJson, err = json.Marshal(data)
	}
	if err != nil {
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	i.Data = dataJson
	i.UpdatedAt = utils.Allocate(time.Now().UTC())
	return nil
}
