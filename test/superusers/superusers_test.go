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

	testFlagStageID string
)

func TestSupersUsersSetup(t *testing.T) {
	app = internal.SetupApp("test")
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
	bodyBytes, statusCode := helpers.API_SuperUsersLogin(
		t,
		app,
		env.SUPERUSER_PASSWORD,
		env.SUPERUSER_PASSWORD,
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
	bodyBytes, statusCode := helpers.API_SuperUsersLogin(
		t,
		app,
		env.SUPERUSER_USERNAME,
		"wrongpassword",
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountLoginWrongPassword,
		bodyBytes,
		statusCode,
	)
}

func TestSuperUsersLoginWrongEmail(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersLogin(
		t,
		app,
		"wrongusername",
		env.SUPERUSER_PASSWORD,
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.SuperUserNotExists,
		bodyBytes,
		statusCode,
	)
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
	testAccountEmail := "initializeaccounttest@example.com"
	testAccountName := "Test Initialize"

	bodyBytes, statusCode := helpers.API_SuperUsersAccountsInitialize(
		t,
		app,
		testAccountEmail,
		testAccountName,
		testSuperUserToken,
	)

	// status code
	require.Equal(t, http.StatusOK, statusCode)

	// unmarshaling the body
	err := json.Unmarshal(bodyBytes, &testAccount)
	assert.NoError(t, err)

	// assertions
	require.NotEmpty(t, testAccount.ID, "expected ID to be set")
	require.Equal(t, testAccountEmail, testAccount.Email, "email should match")
	require.Equal(t, testAccountName, testAccount.Name, "name should match")
}

func TestSuperUsersAccountsInitializeDuplicateEmail(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersAccountsInitialize(
		t,
		app,
		testAccount.Email,
		"Test Initialize",
		testSuperUserToken,
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountAlreadyInitialized,
		bodyBytes,
		statusCode,
	)
}

func TestSuperUsersAccountsCleanup(t *testing.T) {
	err := testAccount.Delete()
	if err != nil {
		fmt.Printf("failed to delete account: %v", err)
	}
	testAccount = models.Account{}
}

func TestSuperUsersFlagsSetup(t *testing.T) {
	_, statusCode := helpers.API_SuperUsersFlagsSetBulk(
		t,
		app,
		map[string]bool{
			"test":    true,
			"testing": true,
		},
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)
}

func TestSuperUsersFlagsGet(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagsGet(
		t,
		app,
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body models.Flags
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)
}

func TestSuperUsersFlagsSet(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagsSet(
		t,
		app,
		"test",
		false,
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body map[string]bool
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)
	require.Equal(t, body["test"], false)
}

func TestSuperUsersFlagsMiddlewareFlagRequired(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagsMiddleware(
		t,
		app,
		testSuperUserToken,
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.FlagRequired,
		bodyBytes,
		statusCode,
	)

	helpers.API_SuperUsersFlagsSet(
		t,
		app,
		"test",
		true,
		testSuperUserToken,
	)
}

func TestSuperUsersFlagsMiddleware(t *testing.T) {
	_, statusCode := helpers.API_SuperUsersFlagsMiddleware(
		t,
		app,
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)
}

func TestSuperUsersFlagsUnset(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagsUnset(
		t,
		app,
		"testingbulk1",
		testSuperUserToken,
	)

	bodyBytes, statusCode = helpers.API_SuperUsersFlagsUnset(
		t,
		app,
		"testingbulk2",
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body map[string]bool
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)
}

func TestSuperUsersFlagsReset(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagsReset(
		t,
		app,
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body map[string]bool
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	for _, v := range body {
		require.NotEqual(t, v, true)
	}
}

func TestSuperUsersFlagStagesGet(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagStagesGet(
		t,
		app,
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body []models.FlagStage
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)
}

func TestSuperUsersFlagStagesCreate(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagStagesCreate(
		t,
		app,
		models.FlagStage{
			Name:    "test",
			TurnOff: []string{"flagstagesoff"},
			TurnOn:  []string{"flagstageson"},
		},
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body models.FlagStage
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	testFlagStageID = body.ID
}

func TestSuperUsersFlagStagesExecuteNotFound(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagStagesExecute(
		t,
		app,
		"123",
		testSuperUserToken,
	)
	helpers.ResponseErrorCheck(t, app,
		errmsg.FlagStageNotFound,
		bodyBytes,
		statusCode,
	)
}

func TestSuperUsersFlagStagesExecute(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagStagesExecute(
		t,
		app,
		testFlagStageID,
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)
	var body models.Flags
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, body.Stage.ID, testFlagStageID)
}

func TestSuperUsersFlagStagesDelete(t *testing.T) {
	bodyBytes, statusCode := helpers.API_SuperUsersFlagStagesDelete(
		t,
		app,
		testFlagStageID,
		testSuperUserToken,
	)

	require.Equal(t, http.StatusOK, statusCode)
	var body models.FlagStage
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, body.ID, "")
}

func TestSuperUsersFlagStagesCleanup(t *testing.T) {
	_, statusCode := helpers.API_SuperUsersFlagsUnset(
		t,
		app,
		"flagstageson",
		testSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)

	_, statusCode = helpers.API_SuperUsersFlagsUnset(
		t,
		app,
		"flagstagesoff",
		testSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)
}
