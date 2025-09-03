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

	req, err := http.NewRequest(
		"POST",
		"/accounts/check",
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

	req, err := http.NewRequest(
		"POST",
		"/accounts/register",
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

	req, err := http.NewRequest(
		"POST",
		"/accounts/login",
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

	req, err := http.NewRequest(
		"PATCH",
		"/accounts",
		bytes.NewBuffer(sendBytes),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

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
