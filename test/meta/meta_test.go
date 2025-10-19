package meta

import (
	"backend/internal"
	"backend/internal/env"
	"flag"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"

	"backend/test/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

var (
	metaApp  *fiber.App
	metaOnce sync.Once
	// package-level flags
	envRootFlag    = flag.String("env-root", "", "directory containing environment files")
	appVersionFlag = flag.String("app-version", "", "application version override")
)

func metaTestApp(t *testing.T) *fiber.App {
	t.Helper()

	metaOnce.Do(func() {
		metaApp = internal.SetupApp("test", *envRootFlag, *appVersionFlag)
	})

	return metaApp
}

func TestMain(m *testing.M) {
	// ensure flags are parsed and test environment is initialized
	flag.Parse()

	// initialize app and reset test cache
	metaApp = internal.SetupApp("test", *envRootFlag, *appVersionFlag)
	helpers.ResetTestCache()
	helpers.ResetTestEvents()

	os.Exit(m.Run())
}

func TestMetaPing(t *testing.T) {
	app := metaTestApp(t)

	req, _ := http.NewRequest("GET", "/meta/ping", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "PONG", string(body))
}

func TestMetaVersion(t *testing.T) {
	app := metaTestApp(t)

	req, _ := http.NewRequest("GET", "/meta/version", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, env.VERSION, string(body))
}
