package entity

import (
	"encoding/json"
	"time"

	"github.com/HardDie/fsentry/internal/entry_error"
	"github.com/HardDie/fsentry/internal/utils"
)

type Entry struct {
	Id        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt *time.Time      `json:"createdAt"`
	UpdatedAt *time.Time      `json:"updatedAt"`
	Data      json.RawMessage `json:"data"`
}

func NewEntry(name string, data interface{}) *Entry {
	dataJson, _ := json.Marshal(data)
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
func (i *Entry) UpdateData(data interface{}) error {
	dataJson, err := json.Marshal(data)
	if err != nil {
		return entry_error.Wrap(err, entry_error.ErrorInternal)
	}
	i.Data = dataJson
	i.UpdatedAt = utils.Allocate(time.Now())
	return nil
}
