package accounts

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
	app              *fiber.App
	testAccount      models.Account
	testAccountEmail string = "accountstesting@example.com"
	testPassword     string = "testingpassword"
	testToken        string
)

func TestAccountsPing(t *testing.T) {
	app = internal.SetupApp("dev")

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

func TestAccountsCheckNotInitialized(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsCheck(
		t,
		app,
		testAccountEmail,
	)

	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, errmsg.AccountNotInitialized.StatusCode, statusCode)
	require.Equal(t, errmsg.AccountNotInitialized.Message, body.Message)
}

func TestAccountsSetup(t *testing.T) {

	bodyBytes, statusCode := helpers.API_SuperUsersLogin(
		t,
		app,
		env.SUPERUSER_USERNAME,
		env.SUPERUSER_PASSWORD,
	)
	require.Equal(t, http.StatusOK, statusCode)

	var body struct {
		SuperUser models.SuperUser `json:"superuser"`
		Token     string           `json:"token"`
	}
	json.Unmarshal(bodyBytes, &body)

	bodyBytes, statusCode = helpers.API_SuperUsersAccountsInitialize(
		t,
		app,
		"accountstesting@example.com",
		"Accounts Testing",
		body.Token,
	)
	require.Equal(t, http.StatusOK, statusCode)

	// unmarshaling the body
	json.Unmarshal(bodyBytes, &testAccount)
}

func TestAccountsCheckNotRegistered(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsCheck(
		t,
		app,
		testAccountEmail,
	)

	var body struct {
		Registered bool `json:"registered"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, body.Registered, false)
}

func TestAccountsRegister(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsRegister(
		t,
		app,
		testAccount.Email,
		testPassword,
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

func TestAccountsCheckRegistered(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsCheck(
		t,
		app,
		testAccountEmail,
	)

	var body struct {
		Registered bool `json:"registered"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, body.Registered, true)
}

func TestAccountsRegisterAlreadyRegistered(t *testing.T) {
	// request payload
	payload := struct {
		Password string `json:"password"`
	}{
		Password: "testingpassword",
	}

	bodyBytes, statusCode := helpers.API_AccountsRegister(
		t,
		app,
		testAccount.Email,
		payload.Password,
	)

	// status code
	require.Equal(t, errmsg.AccountAlreadyRegistered.StatusCode, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountAlreadyRegistered.Message, body.Message)
}

func TestAccountsRegisterNotInitialized(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsRegister(
		t,
		app,
		"notinitialized@example.com",
		"testingpassword",
	)

	// status code
	require.Equal(t, errmsg.AccountNotInitialized.StatusCode, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountNotInitialized.Message, body.Message)
}

func TestAccountsLogin(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsLogin(
		t,
		app,
		testAccount.Email,
		testPassword,
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

func TestAccountsLoginWrongPassword(t *testing.T) {

	bodyBytes, statusCode := helpers.API_AccountsLogin(
		t,
		app,
		testAccount.Email,
		"wrongpassword",
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

func TestAccountsLoginWrongEmail(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsLogin(
		t,
		app,
		"wrongemail@example.com",
		testAccount.Password,
	)

	// status code
	require.Equal(t, errmsg.AccountNotInitialized.StatusCode, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountNotInitialized.Message, body.Message)
}

func TestAccountsEdit(t *testing.T) {
	updatedName := "Updated Name"

	bodyBytes, statusCode := helpers.API_AccountsEdit(
		t,
		app,
		"Updated Name",
		testToken,
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
	require.Equal(t, body.Account.Name, updatedName, "expected name to be equal to payload")

	// updating the account and token
	testToken = body.Token
	testAccount = body.Account
}

func TestAccountsEditNoToken(t *testing.T) {
	updatedName := "Updated Name"

	bodyBytes, statusCode := helpers.API_AccountsEdit(
		t,
		app,
		updatedName,
		"",
	)
	// status code
	require.Equal(t, errmsg.AccountNoToken.StatusCode, statusCode)

	// decode response
	var body struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountNoToken.Message, body.Message)
}

func TestAccountsCleanup(t *testing.T) {
	err := testAccount.Delete()
	if err != nil {
		fmt.Printf("failed to delete account: %v", err)
	}

	testAccount = models.Account{}
	testPassword = ""
	testToken = ""
}
