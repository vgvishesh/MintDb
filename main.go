package main

import (
	pb "MintDb/proto"
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", ":8900", "address the gRPC server listens on")
	dataDir := flag.String("data-dir", ".", "directory holding mint.aof")
	repl := flag.Bool("repl", true, "read commands from stdin; when false, run until SIGINT/SIGTERM")
	flag.Parse()

	db := NewDatabase(Disk, *dataDir)
	server := grpc.NewServer()
	pb.RegisterMintDbServer(server, NewGrpcServer(&db))

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	if *repl {
		runREPL(db)
	} else {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals
	}
	server.GracefulStop()
}

// runREPL reads commands from stdin until EXIT/QUIT or EOF.
func runREPL(db Database) {
	scanner := bufio.NewScanner(os.Stdin)
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
			fmt.Println(db.Delete(parts[1]))
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
