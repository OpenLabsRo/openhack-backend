package accounts

import (
	"backend/internal"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"

	"backend/test/helpers"

	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	app                 *fiber.App
	testAccount         models.Account
	testAccountEmail    string = "accountstesting@example.com"
	testAccountPassword string = "testingpassword"
	testAccountToken    string
)

// package-level test flags (registered during init)
var (
	envRootFlag    = flag.String("env-root", "", "directory containing environment files")
	appVersionFlag = flag.String("app-version", "", "application version override")
)

func TestAccountsPing(t *testing.T) {
	// app is initialized in TestMain
	if app == nil {
		t.Fatalf("test app not initialized; ensure TestMain is present and registers flags")
	}

	req, _ := http.NewRequest("GET", "/accounts/meta/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "PONG")
}

func TestMain(m *testing.M) {
	// parse flags registered at package scope
	flag.Parse()

	app = internal.SetupApp("test", *envRootFlag, *appVersionFlag)
	helpers.ResetTestCache()
	helpers.ResetTestEvents()

	// run tests
	os.Exit(m.Run())
}

func TestAccountsCheckNotInitialized(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsAuthCheck(
		t,
		app,
		testAccountEmail,
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountNotInitialized,
		bodyBytes,
		statusCode,
	)
}

func TestAccountsSetup(t *testing.T) {

	bodyBytes, statusCode := helpers.API_SuperUsersAuthLogin(
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

	bodyBytes, statusCode = helpers.API_SuperUsersParticipantsInitialize(
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
	bodyBytes, statusCode := helpers.API_AccountsAuthCheck(
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
	bodyBytes, statusCode := helpers.API_AccountsAuthRegister(
		t,
		app,
		testAccount.Email,
		testAccountPassword,
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
	bodyBytes, statusCode := helpers.API_AccountsAuthCheck(
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

	bodyBytes, statusCode := helpers.API_AccountsAuthRegister(
		t,
		app,
		testAccount.Email,
		payload.Password,
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountAlreadyRegistered,
		bodyBytes,
		statusCode,
	)
}

func TestAccountsRegisterNotInitialized(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsAuthRegister(
		t,
		app,
		"notinitialized@example.com",
		"testingpassword",
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountNotInitialized,
		bodyBytes,
		statusCode,
	)
}

func TestAccountsLogin(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsAuthLogin(
		t,
		app,
		testAccount.Email,
		testAccountPassword,
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

	testAccountToken = body.Token
	testAccount = body.Account
}

func TestAccountsLoginWrongPassword(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsAuthLogin(
		t,
		app,
		testAccount.Email,
		"wrongpassword",
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountLoginWrongPassword,
		bodyBytes,
		statusCode,
	)
}

func TestAccountsLoginWrongEmail(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsAuthLogin(
		t,
		app,
		"wrongemail@example.com",
		testAccount.Password,
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountNotInitialized,
		bodyBytes,
		statusCode,
	)
}

func TestAccountsEdit(t *testing.T) {
	updatedName := "Updated Name"

	bodyBytes, statusCode := helpers.API_AccountsProfileUpdate(
		t,
		app,
		"Updated Name",
		testAccountToken,
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
	testAccountToken = body.Token
	testAccount = body.Account
}

func TestAccountsEditNoToken(t *testing.T) {
	updatedName := "Updated Name"

	bodyBytes, statusCode := helpers.API_AccountsProfileUpdate(
		t,
		app,
		updatedName,
		"",
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountNoToken,
		bodyBytes,
		statusCode,
	)
}

func TestAccountsGetFlags(t *testing.T) {
	bodyBytes, statusCode := helpers.API_AccountsGetFlags(
		t,
		app,
		testAccountToken,
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body models.Flags
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)
}

func TestAccountsCleanup(t *testing.T) {
	err := testAccount.Delete()
	if err != nil {
		fmt.Printf("failed to delete account: %v", err)
	}

	testAccount = models.Account{}
	testAccountPassword = ""
	testAccountToken = ""
}
