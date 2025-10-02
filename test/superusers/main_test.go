package superusers

import (
	"backend/internal"
	"backend/internal/db"
	"backend/test/helpers"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	app *fiber.App
)

func TestMain(m *testing.M) {
	envRoot := flag.String("env-root", "", "directory containing environment files")
	appVersion := flag.String("app-version", "", "application version override")

	flag.Parse()

	app = internal.SetupApp("test", *envRoot, *appVersion)
	helpers.ResetTestCache()
	clearEvents()

	os.Exit(m.Run())
}

func clearEvents() {
	if db.Events == nil {
		log.Fatal("events collection not initialized")
	}

	if _, err := db.Events.DeleteMany(db.Ctx, bson.M{}); err != nil {
		log.Fatalf("failed to clear events collection: %v", err)
	}
}
