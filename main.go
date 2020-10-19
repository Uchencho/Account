package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Uchencho/Account/server"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const (
	// default server address
	defaultServerAddress = "127.0.0.1:8000"
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

	router := mux.NewRouter()

	router.HandleFunc("/api/register", server.Register)

	if err := http.ListenAndServe(defaultServerAddress, router); err != http.ErrServerClosed {
		log.Println(err)
	}
}
