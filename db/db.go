package db

import (
	"backend/env"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Ctx = context.Background()
var Client *mongo.Client

var Accounts *mongo.Collection
var Teams *mongo.Collection

func InitDB(deployment string) error {
	var err error

	Client, err = mongo.Connect(
		Ctx,
		options.Client().ApplyURI(env.MONGO_URI),
	)

	if err != nil {
		return err
	}

	// loading collections
	Accounts = GetCollection(deployment, "accounts", Client)
	Teams = GetCollection(deployment, "teams", Client)

	fmt.Println("Connected to MongoDB")
	return nil
}

func GetCollection(database string, collectionName string, client *mongo.Client) *mongo.Collection {
	return client.Database(database).Collection(collectionName)
}
