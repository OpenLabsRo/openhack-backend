package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// I think what would help me best is
// to flesh out the API calls here, to clear anything i might need

func API_AccountInitialize(t *testing.T, payload struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}) (bodyBytes []byte, statusCode int) {
	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(
		"POST",
		"/accounts/initialize",
		bytes.NewBuffer(sendBytes),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

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

func API_AccountRegister(
	t *testing.T, accountID string, payload struct {
		Password string `json:"password"`
	}) (bodyBytes []byte, statusCode int) {
	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/accounts/register?id=%v", accountID),
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

func API_AccountLogin(
	t *testing.T, payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}) (bodyBytes []byte, statusCode int) {

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

func API_AccountEdit(
	t *testing.T, token string, payload struct {
		Name string `json:"name"`
	}) (bodyBytes []byte, statusCode int) {

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
