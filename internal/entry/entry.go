package entry

import (
	"encoding/json"
	"time"
)

type Entry struct {
	// ID is a name, but it has all special characters removed, all spaces replaced with underscores,
	// and is shortened to 200 characters because some file systems prohibit files from having long names.
	// When a file is created, it has the same name as the ID string. File extension is not saved in the ID.
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

type Service interface {
	Create(path, name string, data interface{}) (*Entry, error)
	Get(path, name string) (*Entry, error)
	Move(path, oldName, newName string) (*Entry, error)
	Update(path, name string, data interface{}) (*Entry, error)
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) (*Entry, error)
}
