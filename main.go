package main

import (
	"context"
	"log"

	"github.com/Uchencho/Account/server"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found, with error: %s", err)
	}
}

func main() {

	defer func() {
		ctx := context.Background()
		if err := server.Client.Disconnect(ctx); err != nil {
			log.Fatalln("Error in disconnecting", err)
		}
		log.Println("Disconnected successfully")
	}()
}
