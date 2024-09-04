package main

import (
	"context"
	"github.com/romanyakovlev/gophkeeper/internal/client"
	"google.golang.org/grpc"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	keeper := client.NewKeeperServiceClient(conn)

	// Adjust path to the actual file you intend to upload
	err = keeper.SaveBytes(context.Background(), "./file.txt")
	err = keeper.SaveBytes(context.Background(), "./file.txt")
	_, err = keeper.SaveCredentials(context.Background(), "login", "password")
	//err = keeper.GetBytes(context.Background(), "df24daf0-eefa-4254-bc7f-d1fcb9d263a9")
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
}
