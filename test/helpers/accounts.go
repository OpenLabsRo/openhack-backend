package helpers

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_AccountsAuthCheck(
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
		"/accounts/auth/check",
		sendBytes,
		nil,
	)
}

func API_AccountsAuthRegister(
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
		"/accounts/auth/register",
		sendBytes,
		nil,
	)
}

func API_AccountsAuthLogin(
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
		"/accounts/auth/login",
		sendBytes,
		nil,
	)
}

func API_AccountsProfileUpdate(
	t *testing.T,
	app *fiber.App,
	firstName string,
	lastName string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// payload for edit request
	payload := struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}{
		FirstName: firstName,
		LastName:  lastName,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PATCH",
		"/accounts/me",
		sendBytes,
		&token,
	)
}

func API_AccountsGetFlags(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/accounts/flags",
		nil,
		&token,
	)
}

func API_AccountsVotingStatus(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/accounts/voting/status",
		nil,
		&token,
	)
}

func API_AccountsVotingFinalists(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/accounts/voting/finalists",
		nil,
		&token,
	)
}

func API_AccountsVotingCastVote(
	t *testing.T,
	app *fiber.App,
	teamID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	payload := struct {
		TeamID string `json:"teamID"`
	}{
		TeamID: teamID,
	}

	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/accounts/voting/vote",
		sendBytes,
		&token,
	)
}
