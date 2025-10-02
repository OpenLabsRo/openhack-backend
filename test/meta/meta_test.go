package meta

import (
	"backend/internal"
	"backend/internal/env"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

var (
	metaApp  *fiber.App
	metaOnce sync.Once
)

func metaTestApp(t *testing.T) *fiber.App {
	t.Helper()

	metaOnce.Do(func() {
		metaApp = internal.SetupApp("test", "", "")
	})

	return metaApp
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
