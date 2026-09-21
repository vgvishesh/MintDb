package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Database interface {
	Init()
	Get(key string) (string, bool)
	Set(key string, value string) bool
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
	kvStore    map[string]string
	fileName   string
	fileHandle *os.File
}

func NewDatabase(engineType StorageEngineType) Database {
	var storage Database
	switch engineType {
	case InMemory:
		storage = &StorageEngineInMemory{
			kvStore: make(map[string]string),
		}
		storage.Init()
	case Disk:
		const fileName = "mint.aof"
		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println("error initializing database with disk storage", err)
		}

		storage = &DiskStorageEngine{
			kvStore:    make(map[string]string),
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

func (d *DiskStorageEngine) Set(key string, value string) bool {
	d.kvStore[key] = value
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
		data := strings.Split((scanner.Text()), ",")
		d.kvStore[data[0]] = data[1]
	}
}

func (s *DiskStorageEngine) Get(key string) (string, bool) {
	value, ok := s.kvStore[key]
	if !ok {
		return "", ok
	}
	return value, ok
}
