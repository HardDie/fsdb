package entity

import (
	"encoding/json"
	"time"

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
