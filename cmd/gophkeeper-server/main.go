package main

import (
	"github.com/romanyakovlev/gophkeeper/internal/server"
	"log"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("An error occurred: %v", err)
	}
}
