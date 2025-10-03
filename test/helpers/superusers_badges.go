package helpers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_SuperUsersBadgesGet(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/superusers/badges",
		nil,
		&token,
	)
}

func API_SuperUsersBadgesCompute(
	t *testing.T,
	app *fiber.App,
	trials *int,
	token string,
) (bodyBytes []byte, statusCode int) {
	var payload []byte
	if trials != nil {
		body := struct {
			Trials int `json:"trials"`
		}{Trials: *trials}

		var err error
		payload, err = json.Marshal(body)
		require.NoError(t, err)
	}

	return RequestRunner(t, app,
		"POST",
		"/superusers/badges",
		payload,
		&token,
		fiber.TestConfig{
			Timeout: 5 * time.Second,
		},
	)
}
