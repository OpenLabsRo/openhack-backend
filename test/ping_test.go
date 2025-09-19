package test

import (
	"backend/internal"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	app := internal.SetupApp("test")

	req, _ := http.NewRequest("GET", "/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "PONG")
}
