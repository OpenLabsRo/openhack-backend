package helpers

import (
	"log"

	"backend/internal/db"

	"go.mongodb.org/mongo-driver/bson"
)

func ResetTestCache() {
	if db.RDB == nil {
		return
	}
	_ = db.RDB.FlushDB(db.Ctx).Err()
}

func ResetTestEvents() {
	if db.Events == nil {
		log.Fatal("events collection not initialized")
	}

	if _, err := db.Events.DeleteMany(db.Ctx, bson.M{}); err != nil {
		log.Fatalf("failed to clear events collection: %v", err)
	}
}
