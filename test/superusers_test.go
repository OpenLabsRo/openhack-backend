package test

import (
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuperUsersPing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/superusers/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "PONG")
}

func TestSuperUsersLogin(t *testing.T) {
	// request payload
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: env.SUPERUSER_USERNAME,
		Password: env.SUPERUSER_PASSWORD,
	}

	bodyBytes, statusCode := API_SuperUsersLogin(
		t, payload,
	)

	// status code
	require.Equal(t, http.StatusOK, statusCode)

	// decode response
	var body struct {
		SuperUser models.SuperUser `json:"superuser"`
		Token     string           `json:"token"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.NotEmpty(t, body.SuperUser.Username, "expected ID to be set")
	require.NotEmpty(t, body.Token, "expected token to be set")

	testSuperUserToken = body.Token
	testSuperUser = body.SuperUser
}

func TestSuperUsersLoginWrongPassword(t *testing.T) {
	// request payload
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: env.SUPERUSER_USERNAME,
		Password: "wrongpassword",
	}

	bodyBytes, statusCode := API_SuperUsersLogin(
		t, payload,
	)

	// status code
	require.Equal(t, errmsg.AccountLoginWrongPassword.StatusCode, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountLoginWrongPassword.Message, body.Message)
}

func TestSuperUsersLoginWrongEmail(t *testing.T) {
	// request payload
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: "wrongusername",
		Password: env.SUPERUSER_PASSWORD,
	}

	bodyBytes, statusCode := API_SuperUsersLogin(
		t, payload,
	)

	// status code
	require.Equal(t, errmsg.AccountNotExists.StatusCode, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountNotExists.Message, body.Message)
}

func TestSuperUsersWhoAmI(t *testing.T) {
	bodyBytes, statusCode := API_SuperUsersWhoAmI(
		t, testSuperUserToken,
	)

	// status code
	require.Equal(t, http.StatusOK, statusCode)

	// decode response
	var body struct {
		SuperUser models.SuperUser `json:"superuser"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, body.SuperUser.Username, env.SUPERUSER_USERNAME)

}
