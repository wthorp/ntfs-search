package main

import (
	"path/filepath"
	"time"

	"github.com/lxn/walk"
)

type FileInfo struct {
	Name     string
	Path     string
	Size     int64
	Modified time.Time
}

type FileInfoModel struct {
	walk.SortedReflectTableModelBase
	items []*FileInfo
	agent *Agent
}

var _ walk.ReflectTableModel = new(FileInfoModel)

func NewFileInfoModel(agent *Agent) *FileInfoModel {
	return &FileInfoModel{agent: agent}
}

func (m *FileInfoModel) Items() interface{} {
	return m.items
}

func (m *FileInfoModel) SetQuery(query string) error {
	m.items = nil
	m.items, _ = m.agent.Query(query)
	m.PublishRowsReset()
	return nil
}

func (m *FileInfoModel) Image(row int) interface{} {
	return filepath.Join(m.items[row].Path, m.items[row].Name)
}
