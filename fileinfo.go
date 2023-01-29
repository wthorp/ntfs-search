package main

import (
	"context"
	"path"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/gentlemanautomaton/volmgmt/fileapi"
	"github.com/gentlemanautomaton/volmgmt/fileattr"
	"github.com/gentlemanautomaton/volmgmt/usn"
	"github.com/gentlemanautomaton/volmgmt/volume"
	"github.com/lxn/walk"
	"golang.org/x/sys/windows"
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
}

var _ walk.ReflectTableModel = new(FileInfoModel)

func NewFileInfoModel() *FileInfoModel {
	return new(FileInfoModel)
}

func (m *FileInfoModel) Items() interface{} {
	return m.items
}

func (m *FileInfoModel) SetQuery(query string) error {
	m.items = nil

	ctx := context.TODO()
	q, err := regexp.Compile("(?i)" + query)
	if err != nil {
		return err
	}
	vol, err := volume.New("c:\\")
	if err != nil {
		return err
	}
	defer vol.Close()
	volHandle := vol.Handle()

	mft := vol.MFT()
	defer mft.Close()

	iter, err := mft.Enumerate(nil, usn.Min, usn.Max)
	if err != nil {
		return err
	}
	defer iter.Close()

	cache := usn.NewCache()
	err = cache.ReadFrom(ctx, iter)
	if err != nil {
		return err
	}

	records := cache.Records()

	//type token struct{}
	//sem := make(chan token, runtime.NumCPU())
	for _, record := range records {
		// sem <- token{}
		// if ctx.Err() != nil {
		// 	return err
		// }
		// go func(i int, record usn.Record) {
		if record.FileAttributes.Match(fileattr.ReparsePoint) {
			continue
		}
		if !q.Match([]byte(path.Join(record.Path, record.FileName))) {
			continue
		}

		const access = uint32(windows.READ_CONTROL)
		const shareMode = uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE | syscall.FILE_SHARE_DELETE)
		fileHandle, err := fileapi.OpenFileByID(volHandle, record.FileReferenceNumber, access, shareMode, syscall.FILE_FLAG_BACKUP_SEMANTICS)
		if err != nil {
			continue
		}
		defer syscall.CloseHandle(fileHandle)

		fileInfo := fileapi.FileInfoForHandle{FileName: record.FileName}
		fileInfo.ByHandleFileInformation, err = fileapi.GetFileInformationByHandle(fileHandle)
		if err != nil {
			continue
		}
		m.items = append(m.items, &FileInfo{Name: record.FileName, Path: record.Path, Size: fileInfo.Size(), Modified: fileInfo.ModTime()})
		// 	<-sem
		// }(i, record)
	}

	m.PublishRowsReset()

	return nil
}

func (m *FileInfoModel) Image(row int) interface{} {
	return filepath.Join(m.items[row].Path, m.items[row].Name)
}
