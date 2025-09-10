package helpers

import (
	"backend/internal/models"
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_SuperUsersFlagsGet(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	// marshalling the payload into JSON

	return RequestRunner(t, app,
		"GET",
		"/superusers/flags",
		[]byte{},
		&token,
	)
}

func API_SuperUsersFlagsSet(
	t *testing.T,
	app *fiber.App,
	flag string,
	value bool,
	token string,
) (bodyBytes []byte, statusCode int) {
	// marshalling the payload into JSON
	payload := struct {
		Flag  string `json:"flag"`
		Value bool   `json:"value"`
	}{
		Flag:  flag,
		Value: value,
	}

	// marshalling the paylaod into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/superusers/flags",
		sendBytes,
		&token,
	)
}

func API_SuperUsersFlagsSetBulk(
	t *testing.T,
	app *fiber.App,
	rawFlags map[string]bool,
	token string,
) (bodyBytes []byte, statusCode int) {
	// marshalling the paylaod into JSON
	sendBytes, err := json.Marshal(rawFlags)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PUT",
		"/superusers/flags",
		sendBytes,
		&token,
	)
}

func API_SuperUsersFlagsReset(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"PUT",
		"/superusers/flags/reset",
		[]byte{},
		&token,
	)
}

func API_SuperUsersFlagsUnset(
	t *testing.T,
	app *fiber.App,
	flag string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// marshalling the payload into JSON
	payload := struct {
		Flag  string `json:"flag"`
		Value bool   `json:"value"`
	}{
		Flag: flag,
	}

	// marshalling the paylaod into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"DELETE",
		"/superusers/flags",
		sendBytes,
		&token,
	)
}

func API_SuperUsersFlagsMiddleware(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/superusers/flags/test",
		[]byte{},
		&token,
	)
}

func API_SuperUsersFlagStagesGet(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/superusers/flags/stages",
		[]byte{},
		&token,
	)
}

func API_SuperUsersFlagStagesCreate(
	t *testing.T,
	app *fiber.App,
	fstage models.FlagStage,
	token string,
) (bodyBytes []byte, statusCode int) {
	// marshaling the flagstage into JSON
	sendBytes, err := json.Marshal(fstage)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/superusers/flags/stages",
		sendBytes,
		&token,
	)
}

func API_SuperUsersFlagStagesDelete(
	t *testing.T,
	app *fiber.App,
	id string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"DELETE",
		"/superusers/flags/stages?id="+id,
		[]byte{},
		&token,
	)
}

func API_SuperUsersFlagStagesExecute(
	t *testing.T,
	app *fiber.App,
	id string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"POST",
		"/superusers/flags/stages/execute?id="+id,
		[]byte{},
		&token,
	)
}
