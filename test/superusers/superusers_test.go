package superusers

import (
	"backend/internal"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"

	"backend/test/helpers"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	app                *fiber.App
	testSuperUser      models.SuperUser
	testSuperUserToken string

	testAccount models.Account
)

func TestSupersUsersSetup(t *testing.T) {
	app = internal.SetupApp("dev")
	fmt.Println("SuperUsers Setup Complete!")
}

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

	bodyBytes, statusCode := helpers.API_SuperUsersLogin(
		app, t, payload,
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

	bodyBytes, statusCode := helpers.API_SuperUsersLogin(
		app, t, payload,
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

	bodyBytes, statusCode := helpers.API_SuperUsersLogin(
		app, t, payload,
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
	bodyBytes, statusCode := helpers.API_SuperUsersWhoAmI(
		app, t, testSuperUserToken,
	)

	// status code
	require.Equal(t, http.StatusOK, statusCode)

	// decode responsae
	var body struct {
		SuperUser models.SuperUser `json:"superuser"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, body.SuperUser.Username, env.SUPERUSER_USERNAME)
}

func TestSuperUsersAccountsInitialize(t *testing.T) {
	// request payload
	payload := struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}{
		Email: "initializeaccounttest@example.com",
		Name:  "Test Initialize",
	}

	bodyBytes, statusCode := helpers.API_SuperUsersAccountsInitialize(
		app, t, payload, testSuperUserToken,
	)

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

func TestSuperUsersAccountsInitializeDuplicateEmail(t *testing.T) {
	payload := struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}{
		Email: testAccount.Email,
		Name:  "Test Initialize",
	}

	bodyBytes, statusCode := helpers.API_SuperUsersAccountsInitialize(
		app, t, payload, testSuperUserToken,
	)

	require.Equal(t, errmsg.AccountExists.StatusCode, statusCode)

	// unmarshaling the error message
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, errmsg.AccountExists.Message, body.Message)
}

func TestSuperUsersAccountsCleanup(t *testing.T) {
	err := testAccount.Delete()
	if err != nil {
		fmt.Printf("failed to delete account: %v", err)
	}
	testAccount = models.Account{}
}
