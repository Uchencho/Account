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
	}()

	router := mux.NewRouter()

	router.Handle("/api/register", server.BasicToken(http.HandlerFunc(server.Register)))

	if err := http.ListenAndServe(defaultServerAddress, router); err != http.ErrServerClosed {
		log.Println(err)
	}
}
