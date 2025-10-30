package helpers

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_SuperUsersAuthLogin(
	t *testing.T,
	app *fiber.App,
	username string,
	password string,
) (bodyBytes []byte, statusCode int) {
	// payload
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/superusers/auth/login",
		sendBytes,
		nil,
	)
}

func API_SuperUsersMetaWhoAmI(
	app *fiber.App,
	t *testing.T, token string) (bodyBytes []byte, statusCode int) {

	return RequestRunner(t, app,
		"GET",
		"/superusers/meta/whoami",
		[]byte{},
		&token,
	)
}

func API_SuperUsersParticipantsInitialize(
	t *testing.T,
	app *fiber.App,
	email string,
	firstName string,
	lastName string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// payload
	payload := struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/superusers/participants",
		sendBytes,
		&token,
	)
}
