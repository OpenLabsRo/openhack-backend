package test

import (
	"backend/internal/db"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Initialize cache
	db.InitCache("test")

	// Initialize the event counter in Redis
	db.RDB.Set(db.Ctx, "event_counter", 0, 0)

	// Run all tests
	code := m.Run()

	os.Exit(code)
}
