package main

/*
-communicate over cmd and REST: GET, SET
signature:
GET(key)
SET(key, Value)

- Values are strings for now
- Save values in a File
- Keep them in memory for retrival
- write external script for testing failures and load testing
	- create values
	- get values
	- fail the server
	- restart the server
	- benchmarking
*/

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Database interface {
	Get(key string) (string, bool)
	Set(key string, value string) bool
}

type StorageEngine struct {
	kvStore map[string]string
}

func NewDatabase() *StorageEngine {
	storage := &StorageEngine{
		kvStore: make(map[string]string),
	}
	return storage
}

func (s *StorageEngine) Set(key string, value string) bool {
	s.kvStore[key] = value
	return true
}

func (s *StorageEngine) Get(key string) (string, bool) {
	value, ok := s.kvStore[key]
	if !ok {
		return "", ok
	}
	return value, ok
}


func main() {
	scanner := bufio.NewScanner(os.Stdin)
	db := NewDatabase()
	fmt.Print("> ")
	for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					parts := strings.SplitN(line, " ", 3)

					switch strings.ToUpper(parts[0]) {
					case "GET":
									if len(parts) < 2 {
													fmt.Println("usage: GET key")
													break
									}
									if value, ok := db.Get(parts[1]); ok {
										fmt.Println(value)
									} else {
										fmt.Println("not found")
									}
									
					case "SET":
									if len(parts) < 3 {
													fmt.Println("usage: SET key value")
													break
									}
									if ok := db.Set(parts[1], parts[2]); ok {
										fmt.Println("value stored") 
									} else {
										fmt.Println("error in storage");
									}
					case "EXIT", "QUIT":
									return
					case "":
									// ignore blank lines
					default:
									fmt.Println("unknown command:", parts[0])
					}
					fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
					fmt.Fprintln(os.Stderr, "read error:", err)
	}
}