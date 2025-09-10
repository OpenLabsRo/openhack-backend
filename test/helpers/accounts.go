package helpers

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_AccountsCheck(
	t *testing.T,
	app *fiber.App,
	email string,
) (bodyBytes []byte, statusCode int) {
	// building the payload
	payload := struct {
		Email string `json:"email"`
	}{
		Email: email,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/accounts/check",
		sendBytes,
		nil,
	)
}

func API_AccountsRegister(
	t *testing.T,
	app *fiber.App,
	email string,
	password string,
) (bodyBytes []byte, statusCode int) {
	// building the payload
	payload := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    email,
		Password: password,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/accounts/register",
		sendBytes,
		nil,
	)
}

func API_AccountsLogin(
	t *testing.T,
	app *fiber.App,
	email string,
	password string,
) (bodyBytes []byte, statusCode int) {
	// payload for login request
	payload := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    email,
		Password: password,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/accounts/login",
		sendBytes,
		nil,
	)
}

func API_AccountsEdit(
	t *testing.T,
	app *fiber.App,
	name string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// payload for edit request
	payload := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PATCH",
		"/accounts",
		sendBytes,
		&token,
	)
}
