package folder

import (
	"encoding/json"
	"time"
)

type Info struct {
	// ID is a name, but it has all special characters removed, all spaces replaced with underscores,
	// and is shortened to 200 characters because some file systems prohibit files from having long names.
	ID string `json:"id"`
	// Name is the original name that was set by the user without any modification.
	Name string `json:"name"`
	// CreatedAt metadata for each Entry to track the original creation date.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt metadata to keep track of when the Entry was last updated.
	UpdatedAt time.Time `json:"updatedAt"`
	// Data is a custom json payload for custom data.
	Data json.RawMessage `json:"data"`
}

type Folder interface {
	Create(path, name string, data interface{}) (*Info, error)
	Get(path, name string) (*Info, error)
	Move(path, oldName, newName string) (*Info, error)
	Update(path, name string, data interface{}) (*Info, error)
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) (*Info, error)
	MoveWithoutTimestamp(path, oldName, newName string) (*Info, error)
}
