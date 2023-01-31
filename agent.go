package main

import (
	"context"
	"regexp"
	"syscall"
	"time"

	"github.com/gentlemanautomaton/volmgmt/fileapi"
	"github.com/gentlemanautomaton/volmgmt/fileattr"
	"github.com/gentlemanautomaton/volmgmt/usn"
	"github.com/gentlemanautomaton/volmgmt/volume"
	"golang.org/x/sys/windows"
)

type Agent struct {
	VolHandle syscall.Handle
	Records   []usn.Record
}

func NewNTFSAgent(ctx context.Context, volumePath string) (*Agent, error) {
	vol, err := volume.New(volumePath)
	if err != nil {
		return nil, err
	}
	defer vol.Close()
	volHandle := vol.Handle()

	mft := vol.MFT()
	defer mft.Close()

	iter, err := mft.Enumerate(nil, usn.Min, usn.Max)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	cache := usn.NewCache()
	err = cache.ReadFrom(ctx, iter)
	if err != nil {
		return nil, err
	}

	records := cache.Records()
	return &Agent{VolHandle: volHandle, Records: records}, nil
}

func (a *Agent) Query(query string) ([]*FileInfo, error) {
	items := []*FileInfo{}

	q, err := regexp.Compile("(?i)" + query)
	if err != nil {
		return nil, err
	}

	//type token struct{}
	//sem := make(chan token, runtime.NumCPU())
	for _, record := range a.Records {
		// sem <- token{}
		// if ctx.Err() != nil {
		// 	return err
		// }
		// go func(i int, record usn.Record) {
		if record.FileAttributes.Match(fileattr.ReparsePoint) {
			continue
		}
		if !q.Match([]byte(record.Path)) {
			continue
		}

		size, modTime := a.GetFileInfo(&record)
		items = append(items, &FileInfo{Name: record.FileName, Path: record.Path, Size: size, Modified: modTime})
		// 	<-sem
		// }(i, record)
	}

	return items, nil
}

func (a *Agent) GetFileInfo(record *usn.Record) (int64, time.Time) {
	const access = uint32(windows.READ_CONTROL)
	const shareMode = uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE | syscall.FILE_SHARE_DELETE)
	fileHandle, err := fileapi.OpenFileByID(a.VolHandle, record.FileReferenceNumber, access, shareMode, syscall.FILE_FLAG_BACKUP_SEMANTICS)
	if err != nil {
		return 0, time.Time{}
	}
	defer syscall.CloseHandle(fileHandle)

	fileInfo := fileapi.FileInfoForHandle{FileName: record.FileName}
	fileInfo.ByHandleFileInformation, err = fileapi.GetFileInformationByHandle(fileHandle)
	if err != nil {
		return 0, time.Time{}
	}
	return fileInfo.Size(), fileInfo.ModTime()
}
