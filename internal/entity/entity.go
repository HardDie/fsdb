package entity

import (
	"encoding/json"
	"time"

	"github.com/HardDie/fsdb/internal/utils"
)

type FolderInfo struct {
	Id        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt *time.Time      `json:"createdAt"`
	UpdatedAt *time.Time      `json:"updatedAt"`
	Data      json.RawMessage `json:"data"`
}

func NewFolderInfo(name string, data interface{}) *FolderInfo {
	dataJson, _ := json.Marshal(data)
	return &FolderInfo{
		Id:        utils.NameToID(name),
		Name:      name,
		CreatedAt: utils.Allocate(time.Now()),
		Data:      dataJson,
	}
}
