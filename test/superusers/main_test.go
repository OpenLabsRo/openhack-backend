package superusers

import (
	"backend/internal"
	"backend/test/helpers"
	"flag"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
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
	helpers.ResetTestEvents()

	os.Exit(m.Run())
}
