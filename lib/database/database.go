/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package database

import (
	"context"

	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// MongoDB exported mongodb connection
	MongoDB       *mongo.Database
	MongoDBClient *mongo.Client
)

// ConnectMongo creates new MongoDB connection.
func ConnectMongo() *mongo.Database {
	// mongoURI mongo db uri
	var mongoURI string
	if os.Getenv("MONGODB_URL") != "" {
		mongoURI = os.Getenv("MONGODB_URL")
	} else {
		log.Fatal("MONGODB_URL is empty or not set in environment variables")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalln(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	if os.Getenv("MONGODB_DATABASE") != "" {
		MongoDB = client.Database(os.Getenv("MONGODB_DATABASE"))
		MongoDBClient = client
	} else {
		log.Fatal("MONGODB_DATABASE is empty or not set in environment variables")
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Infoln("Connected to MongoDB Server")
	}
	return MongoDB
}
