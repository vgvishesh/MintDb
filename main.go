package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	db := NewDatabase(Disk)
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
				fmt.Println("error in storage")
			}
		case "DELETE":
			if len(parts) < 2 {
				fmt.Println("usage: DELETE key")
				break
			}
			if  ok := db.Delete(parts[1]); ok {
				fmt.Println(ok)
			} else {
				fmt.Println(ok)
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
