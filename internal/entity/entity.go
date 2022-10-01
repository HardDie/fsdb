package entity

import (
	"encoding/json"
	"time"

	"github.com/HardDie/fsentry/internal/fsdberror"
	"github.com/HardDie/fsentry/internal/utils"
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

func (i *FolderInfo) SetName(name string) *FolderInfo {
	i.Name = name
	i.Id = utils.NameToID(name)
	return i
}
func (i *FolderInfo) UpdatedNow() *FolderInfo {
	i.UpdatedAt = utils.Allocate(time.Now())
	return i
}
func (i *FolderInfo) FlushTime() *FolderInfo {
	i.CreatedAt = utils.Allocate(time.Now())
	i.UpdatedAt = nil
	return i
}
func (i *FolderInfo) UpdateData(data interface{}) error {
	dataJson, err := json.Marshal(data)
	if err != nil {
		return fsdberror.Wrap(err, fsdberror.ErrorInternal)
	}
	i.Data = dataJson
	i.UpdatedAt = utils.Allocate(time.Now())
	return nil
}
