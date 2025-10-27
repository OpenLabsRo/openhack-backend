package db

import (
	"backend/internal/env"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Ctx = context.Background()
var RDB *redis.Client
var Client *mongo.Client
var DB_DEPLOYMENT string

var Accounts *mongo.Collection
var Teams *mongo.Collection
var Flags *mongo.Collection
var SuperUsers *mongo.Collection
var FlagStages *mongo.Collection
var Events *mongo.Collection
var Tags *mongo.Collection
var Settings *mongo.Collection
var Judges *mongo.Collection
var Judgments *mongo.Collection

func InitDB(deployment string) error {
	DB_DEPLOYMENT = deployment
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
	FlagStages = GetCollection(deployment, "flagstages", Client)
	Events = GetCollection(deployment, "events", Client)
	Tags = GetCollection(deployment, "tags", Client)
	Settings = GetCollection(deployment, "settings", Client)
	Judges = GetCollection(deployment, "judges", Client)
	Judgments = GetCollection(deployment, "judgments", Client)

	return nil
}

func GetCollection(database string, collectionName string, client *mongo.Client) *mongo.Collection {
	return client.Database(database).Collection(collectionName)
}

func InitCache(deployment string) error {
	var err error

	redisDatabase := 17
	switch deployment {
	case "prod":
		redisDatabase = 0
	case "dev":
		redisDatabase = 1
	case "test":
		redisDatabase = 2
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       redisDatabase,
	})

	err = RDB.Ping(Ctx).Err()
	if err != nil {
		log.Fatal("COULD NOT CONNECT TO REDIS")
		return err
	}

	return nil
}

func CacheSet(key string, value string) error {
	if DB_DEPLOYMENT != "prod" {
		return redis.Nil
	}

	return RDB.Set(Ctx, key, value, 0).Err()
}

func CacheSetBytes(key string, value []byte) error {
	if DB_DEPLOYMENT != "prod" {
		return redis.Nil
	}

	return RDB.Set(Ctx, key, value, 0).Err()
}

func CacheGet(key string) (string, error) {
	if DB_DEPLOYMENT != "prod" {
		return "", redis.Nil
	}

	return RDB.Get(Ctx, key).Result()
}

func CacheGetBytes(key string) ([]byte, error) {
	if DB_DEPLOYMENT != "prod" {
		return []byte{}, redis.Nil
	}

	return RDB.Get(Ctx, key).Bytes()

}

func CacheDel(key string) error {
	if DB_DEPLOYMENT != "prod" {
		return redis.Nil
	}
	_, err := RDB.Del(Ctx, key).Result()

	return err
}
