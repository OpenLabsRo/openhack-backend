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
	// package-level flags
	envRootFlag    = flag.String("env-root", "", "directory containing environment files")
	appVersionFlag = flag.String("app-version", "", "application version override")
)

func TestMain(m *testing.M) {
	flag.Parse()

	app = internal.SetupApp("test", *envRootFlag, *appVersionFlag)
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
