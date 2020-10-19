package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Uchencho/Account/server"
)

func main() {

	fmt.Println("It starts here")

	defer func() {
		ctx := context.Background()
		if err := server.Client.Disconnect(ctx); err != nil {
			log.Fatalln("Error in disconnecting", err)
		}
		log.Println("Disconnected successfully")
	}()
	fmt.Println("It connected")

	fmt.Println("When will this be called?")
}
