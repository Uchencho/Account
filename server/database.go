package server

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func connectDB() *mongo.Client {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Error in connecting to db")
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalln("Could not ping db with error", err)
	}
	log.Println("Connected to MongoDB")
	return client
}

var Client = connectDB()
