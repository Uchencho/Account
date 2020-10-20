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

	// Using Gorilla Mux as a router
	router := mux.NewRouter()
	router.NotFoundHandler = server.BasicToken(http.HandlerFunc(server.NotAvailable))

	router.Handle("/api/register", server.BasicToken(http.HandlerFunc(server.Register)))
	router.Handle("/api/login", server.BasicToken(http.HandlerFunc(server.Login)))
	router.Handle("/api/profile", server.TheUser(http.HandlerFunc(server.UserProfile)))
	router.HandleFunc("/api/refresh-token", server.RefreshTokenAPI)

	// // Using HttpRouter as a router(https://github.com/julienschmidt/httprouter)
	// /*
	// 	Current challenge with this router is the fact that you must specify the method coming in
	//  This means I am unable to handle different methods in the handler, rather would have to repeat
	//  myself
	// */
	// router := httprouter.New()
	// router.Handler("POST", "/api/register", server.BasicToken(http.HandlerFunc(server.Register)))
	// router.Handler("POST", "/api/login", server.BasicToken(http.HandlerFunc(server.Login)))
	// router.Handler("GET", "/api/profile", server.TheUser(http.HandlerFunc(server.UserProfile)))
	// router.NotFound = server.BasicToken(http.HandlerFunc(server.NotAvailable))

	if err := http.ListenAndServe(defaultServerAddress, router); err != http.ErrServerClosed {
		log.Println(err)
	}
}
