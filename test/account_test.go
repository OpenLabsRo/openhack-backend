package test

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AccountCleanup() {
	err := testAccount.Delete()
	if err != nil {
		fmt.Printf("failed to delete account: %v", err)
	}

	testAccount = models.Account{}
	testPassword = ""
	testToken = ""
}

func TestAccountPing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/accounts/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "PONG")
}

func TestAccountInitialize(t *testing.T) {
	// request payload
	payload := struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}{
		Email: "testaccount1@openhack.ro",
		Name:  "Test Initialize",
	}

	bodyBytes, statusCode := API_AccountInitialize(t, payload)

	// status code
	require.Equal(t, http.StatusOK, statusCode)

	// unmarshaling the body
	err := json.Unmarshal(bodyBytes, &testAccount)
	assert.NoError(t, err)

	// assertions
	require.NotEmpty(t, testAccount.ID, "expected ID to be set")
	require.Equal(t, payload.Email, testAccount.Email, "email should match")
	require.Equal(t, payload.Name, testAccount.Name, "name should match")
}

func TestAccountInitializeDuplicateEmail(t *testing.T) {
	payload := struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}{
		Email: testAccount.Email,
		Name:  "Test Initialize",
	}

	bodyBytes, statusCode := API_AccountInitialize(t, payload)

	require.Equal(t, http.StatusConflict, statusCode)

	// unmarshaling the error message
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, errmsg.AccountInitializeAlreadyExists, body.Message)
}

func TestAccountRegister(t *testing.T) {
	// request payload
	payload := struct {
		Password string `json:"password"`
	}{
		Password: "testingpassword",
	}
	testPassword = payload.Password

	bodyBytes, statusCode := API_AccountRegister(
		t, testAccount.ID, payload,
	)

	// status code
	require.Equal(t, http.StatusOK, statusCode)

	// decode response
	var body struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.NotEmpty(t, body.Account.Password, "expected password to be set")
	require.NotEmpty(t, body.Token, "expected token to be set")
}

func TestAccountRegisterNotExist(t *testing.T) {
	// request payload
	payload := struct {
		Password string `json:"password"`
	}{
		Password: "testingpassword",
	}

	bodyBytes, statusCode := API_AccountRegister(
		t, "123", payload,
	)

	// status code
	require.Equal(t, http.StatusNotFound, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountRegisterNotExist, body.Message)
}

func TestAccountLogin(t *testing.T) {
	// request payload
	payload := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    testAccount.Email,
		Password: testPassword,
	}

	bodyBytes, statusCode := API_AccountLogin(
		t, payload,
	)

	// status code
	require.Equal(t, http.StatusOK, statusCode)

	// decode response
	var body struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.NotEmpty(t, body.Account.ID, "expected ID to be set")
	require.NotEmpty(t, body.Token, "expected token to be set")

	testToken = body.Token
	testAccount = body.Account
}

func TestAccountLoginWrongPassword(t *testing.T) {
	// request payload
	payload := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    testAccount.Email,
		Password: "wrongpassword",
	}

	bodyBytes, statusCode := API_AccountLogin(
		t, payload,
	)

	// status code
	require.Equal(t, http.StatusUnauthorized, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountLoginWrongPassword, body.Message)
}

func TestAccountEdit(t *testing.T) {
	// request payload
	payload := struct {
		Name string `json:"name"`
	}{
		Name: "Updated Name",
	}

	bodyBytes, statusCode := API_AccountEdit(
		t, testToken, payload,
	)

	// status code
	require.Equal(t, http.StatusOK, statusCode)

	// decode response
	var body struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.NotEmpty(t, body.Account.ID, "expected ID to be set")
	require.NotEmpty(t, body.Token, "expected token to be set")
	require.Equal(t, body.Account.Name, payload.Name, "expected name to be equal to payload")

	// updating the account and token
	testToken = body.Token
	testAccount = body.Account
}

func TestAccountEditNoToken(t *testing.T) {
	// request payload
	payload := struct {
		Name string `json:"name"`
	}{
		Name: "Updated Name",
	}

	bodyBytes, statusCode := API_AccountEdit(
		t, "", payload,
	)

	// status code
	require.Equal(t, http.StatusUnauthorized, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountNoToken, body.Message)
}

func TestAccountCleanup(t *testing.T) {
	t.Cleanup(func() {
		AccountCleanup()
	})
}
