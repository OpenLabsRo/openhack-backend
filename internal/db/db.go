package db

import (
	"backend/internal/env"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Ctx = context.Background()
var Client *mongo.Client

var Accounts *mongo.Collection
var Teams *mongo.Collection
var Flags *mongo.Collection
var SuperUsers *mongo.Collection

func InitDB(deployment string) error {
	var err error

	Client, err = mongo.Connect(
		Ctx,
		options.Client().ApplyURI(env.MONGO_URI),
	)
	if err != nil {
		return err
	}

	err = Client.Ping(Ctx, nil)
	if err != nil {
		log.Fatal("COULD NOT CONNECT TO MONGODB")
		return err
	}

	// loading collections
	Accounts = GetCollection(deployment, "accounts", Client)
	Teams = GetCollection(deployment, "teams", Client)
	Flags = GetCollection(deployment, "flags", Client)
	SuperUsers = GetCollection(deployment, "superusers", Client)

	return nil
}

func GetCollection(database string, collectionName string, client *mongo.Client) *mongo.Collection {
	return client.Database(database).Collection(collectionName)
}
