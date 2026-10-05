package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Database interface {
	Init()
	Get(key string) (string, bool)
	Set(key string, value string) bool
	Delete(key string) bool
}

type StorageEngineType int

const (
	Disk StorageEngineType = iota
	InMemory
)

type StorageEngineInMemory struct {
	kvStore map[string]string
}

type DiskStorageEngine struct {
	inMemKvStore    StorageEngineInMemory
	fileName   string
	fileHandle *os.File
}

// NewDatabase creates a storage engine. dataDir is where the Disk engine keeps
// mint.aof and is ignored by InMemory.
func NewDatabase(engineType StorageEngineType, dataDir string) Database {
	var storage Database
	switch engineType {
	case InMemory:
		storage = &StorageEngineInMemory{
			kvStore: make(map[string]string),
		}
		storage.Init()
	case Disk:
		fileName := filepath.Join(dataDir, "mint.aof")
		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println("error initializing database with disk storage", err)
		}

		storage = &DiskStorageEngine{
			inMemKvStore: StorageEngineInMemory{
				kvStore: make(map[string]string),
			},
			fileName:   fileName,
			fileHandle: file,
		}
		storage.Init()
	}

	return storage
}

func (s *StorageEngineInMemory) Set(key string, value string) bool {
	s.kvStore[key] = value
	return true
}

func (s *StorageEngineInMemory) Init() {}

func (s *StorageEngineInMemory) Get(key string) (string, bool) {
	value, ok := s.kvStore[key]
	if !ok {
		return "", ok
	}
	return value, ok
}

func (s *StorageEngineInMemory) Delete(key string) bool {
	delete(s.kvStore, key)
	return true
}

func (d *DiskStorageEngine) Set(key string, value string) bool {
	d.inMemKvStore.Set(key, value)
	d.fileHandle.Seek(0, 2)
	dataToWrite := key + "," + value + "\n"
	_, err := d.fileHandle.WriteString(dataToWrite)
	if err != nil {
		fmt.Println("failed to persist data on disk")
		return false
	}
	return true
}

func (d *DiskStorageEngine) Init() {
	d.fileHandle.Seek(0, 0)
	scanner := bufio.NewScanner(d.fileHandle)
	for scanner.Scan() {
		record := scanner.Text()
		//detected a tombstone record
		if strings.Contains(record, "Delete:") {
			key := strings.Split(record, ":")[1]
			d.inMemKvStore.Delete(key)
		} else {
			separatorIndex := strings.Index(record,",")
			key:= record[:separatorIndex]
			value:= record[separatorIndex+1:]
			d.inMemKvStore.Set(key, value)
		}
	}
}

func (s *DiskStorageEngine) Get(key string) (string, bool) {
	value, ok := s.inMemKvStore.Get(key)
	if !ok {
		return "", ok
	}
	return value, ok
}

func (d *DiskStorageEngine) Delete(key string) bool {
	d.inMemKvStore.Delete(key)
	d.fileHandle.Seek(0, 2)
	deleteRecord := "Delete:" + key + "\n" //tombstone record
	_, err := d.fileHandle.WriteString(deleteRecord)
	if err != nil {
		fmt.Println("failed to persist the delete record")
		return false
	}

	return true
}
