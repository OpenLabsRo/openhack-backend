package main

import (
	"backend/internal/db"
	"backend/internal/env"
	"context"
	"flag"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	envRoot := flag.String("env-root", "", "directory containing environment files")
	appVersion := flag.String("app-version", "", "application version override")
	flag.Parse()

	env.Init(*envRoot, *appVersion)

	if err := db.InitDB("test"); err != nil {
		log.Fatalf("failed to initialize test database: %v", err)
	}
	defer func() {
		if db.Client != nil {
			_ = db.Client.Disconnect(context.Background())
		}
	}()

	if db.Events == nil {
		log.Fatal("events collection is not initialized")
	}

	if _, err := db.Events.DeleteMany(db.Ctx, bson.M{}); err != nil {
		log.Fatalf("failed to clear events collection: %v", err)
	}
}
