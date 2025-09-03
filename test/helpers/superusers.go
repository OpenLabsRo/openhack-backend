package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_SuperUsersLogin(
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

	req, err := http.NewRequest(
		"POST",
		"/superusers/login",
		bytes.NewBuffer(sendBytes),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// send request to the shared app
	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()

	statusCode = res.StatusCode

	bodyBytes, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err) // or handle error normally
	}

	return
}

func API_SuperUsersWhoAmI(
	app *fiber.App,
	t *testing.T, token string) (bodyBytes []byte, statusCode int) {

	req, err := http.NewRequest(
		"GET",
		"/superusers/whoami",
		bytes.NewBuffer([]byte("")),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// send request to the shared app
	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()

	statusCode = res.StatusCode

	bodyBytes, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err) // or handle error normally
	}

	return
}

func API_SuperUsersAccountsInitialize(
	t *testing.T,
	app *fiber.App,
	email string,
	name string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// payload
	payload := struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}{
		Email: email,
		Name:  name,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(
		"POST",
		"/superusers/accounts/initialize",
		bytes.NewBuffer(sendBytes),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// send request to the shared app
	res, err := app.Test(req)
	require.NoError(t, err)

	statusCode = res.StatusCode

	bodyBytes, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err) // or handle error normally
	}
	defer res.Body.Close()

	return
}
