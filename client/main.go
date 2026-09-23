package main

import (
	"context"
	"fmt"
	"log"

	pb "MintDb/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	address := "localhost:8900"
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewMintDbClient(conn)
	ctx := context.Background()
	value, err := client.Get(ctx, &pb.GetRequest{
		Key: "vgApiKey",
	})

	if err!= nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("value: ", value)
	}

	client.Set(ctx, &pb.SetRequest{Key: "vgApiKey", Value: "querty"})

	value, err = client.Get(ctx, &pb.GetRequest{
		Key: "vgApiKey",
	})
	fmt.Println("value: ", value)
}
