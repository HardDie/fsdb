package entity

import (
	"encoding/json"
	"time"

	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

// Entry is the main object for storing data in the fsentry library. It stores data in *.json files,
// which are created on the hard disk in the format shown below.
type Entry struct {
	// Id is a name, but it has all special characters removed, all spaces replaced with underscores,
	// and is shortened to 200 characters because some file systems prohibit files from having long names.
	// When a file is created, it has the same name as the Id string. File extension is not saved in the Id.
	Id string `json:"id"` // TODO: rename to ID
	// Name is the original name that was set by the user without any modification.
	Name string `json:"name"`
	// CreatedAt metadata for each Entry to track the original creation date.
	CreatedAt *time.Time `json:"createdAt"` // TODO: remove pointer, CreatedAt must be set always
	// UpdatedAt metadata to keep track of when the Entry was last updated.
	UpdatedAt *time.Time `json:"updatedAt"` // TODO: remove pointer, UpdatedAt could be init with same value as CreatedAt
	// Data is a custom json payload for custom data.
	Data json.RawMessage `json:"data"`
}

func NewEntry(name string, data interface{}, isIndent bool) *Entry {
	var dataJson []byte
	if isIndent {
		dataJson, _ = json.MarshalIndent(data, "", "	")
	} else {
		dataJson, _ = json.Marshal(data)
	}
	return &Entry{
		Id:        utils.NameToID(name),
		Name:      name,
		CreatedAt: utils.Allocate(time.Now()),
		Data:      dataJson,
	}
}

func (i *Entry) SetName(name string) *Entry {
	i.Name = name
	i.Id = utils.NameToID(name)
	return i
}
func (i *Entry) UpdatedNow() *Entry {
	i.UpdatedAt = utils.Allocate(time.Now())
	return i
}
func (i *Entry) FlushTime() *Entry {
	i.CreatedAt = utils.Allocate(time.Now())
	i.UpdatedAt = nil
	return i
}
func (i *Entry) UpdateData(data interface{}, isIndent bool) error {
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
	i.UpdatedAt = utils.Allocate(time.Now())
	return nil
}
