package test

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

var usersToCreate = 5

var (
	app *fiber.App

	testSuperUserToken string

	testAccounts         []models.Account
	testAccountEmails    []string
	testAccountPasswords []string
	testAccountTokens    []string

	testTeamID string
)

func TestTeamsPing(t *testing.T) {
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

func TestTeamsSetup(t *testing.T) {
	// superuser stuff
	bodyBytes, statusCode := helpers.API_SuperUsersLogin(
		t,
		app,
		env.SUPERUSER_USERNAME,
		env.SUPERUSER_PASSWORD,
	)
	require.Equal(t, http.StatusOK, statusCode)

	var body struct {
		Token string `json:"token"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	testSuperUserToken = body.Token

	// creating all the necessary accounts
	for i := range usersToCreate {
		testAccountEmails = append(
			testAccountEmails,
			fmt.Sprintf("teamstesting%v@example.com", i),
		)
		testAccountPasswords = append(
			testAccountEmails,
			fmt.Sprintf("teamstesting%v", i),
		)

		bodyBytes, statusCode = helpers.API_SuperUsersAccountsInitialize(
			t,
			app,
			fmt.Sprintf("teamstesting%v@example.com", i),
			fmt.Sprintf("Accounts Testing %v", i),
			body.Token,
		)
		require.Equal(t, http.StatusOK, statusCode)

		bodyBytes, statusCode = helpers.API_AccountsRegister(
			t,
			app,
			fmt.Sprintf("teamstesting%v@example.com", i),
			fmt.Sprintf("teamspassword%v", i),
		)
		require.Equal(t, http.StatusOK, statusCode)

		var body struct {
			Token   string         `json:"token"`
			Account models.Account `json:"account"`
		}
		err := json.Unmarshal(bodyBytes, &body)
		require.NoError(t, err)

		testAccounts = append(testAccounts, body.Account)
		testAccountTokens = append(testAccountTokens, body.Token)

		require.Equal(t, body.Account.Email, testAccountEmails[i])
	}
}

func TestTeamsChangeHasNoTeam(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsChange(
		t,
		app,
		"whatever",
		testAccountTokens[0],
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountHasNoTeam,
		bodyBytes,
		statusCode,
	)
}

func TestTeamsGetHasNoTeam(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsGet(
		t,
		app,
		testAccountTokens[0],
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountHasNoTeam,
		bodyBytes,
		statusCode,
	)
}

func TestTeamsCreate(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsCreate(
		t,
		app,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	testAccounts[0] = body.Account
	testAccountTokens[0] = body.Token

	require.NotEqual(t, "", body.Account.TeamID)

	testTeamID = body.Account.TeamID
}

func TestTeamsCreateAlreadyHasTeam(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsCreate(
		t,
		app,
		testAccountTokens[0],
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountAlreadyHasTeam,
		bodyBytes,
		statusCode,
	)
}

func TestTeamsChange(t *testing.T) {
	newName := "Changed Team Name"
	bodyBytes, statusCode := helpers.API_TeamsChange(
		t,
		app,
		newName,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	tempTeam := models.Team{}
	err := json.Unmarshal(bodyBytes, &tempTeam)
	require.NoError(t, err)

	require.Equal(t, tempTeam.Name, newName)
}

func TestTeamsJoinNotFound(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsJoin(
		t,
		app,
		"123",
		testAccountTokens[1],
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.TeamNotFound,
		bodyBytes,
		statusCode,
	)
}

// 1, 2, and 3 join 0's team
func TestTeamsJoin(t *testing.T) {
	for i := 1; i <= 3; i++ {
		bodyBytes, statusCode := helpers.API_TeamsJoin(
			t,
			app,
			testTeamID,
			testAccountTokens[i],
		)

		require.Equal(t, http.StatusOK, statusCode)

		var body struct {
			Account models.Account `json:"account"`
			Token   string         `json:"token"`
		}
		err := json.Unmarshal(bodyBytes, &body)
		require.NoError(t, err)

		require.Equal(t, body.Account.TeamID, testTeamID)

		testAccounts[i] = body.Account
		testAccountTokens[i] = body.Token
	}
}

func TestTeamsJoinTeamFull(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsJoin(
		t,
		app,
		testTeamID,
		testAccountTokens[4],
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.TeamFull,
		bodyBytes,
		statusCode,
	)
}

func TestTeamsJoinAlreadyHasTeam(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsJoin(
		t,
		app,
		testTeamID,
		testAccountTokens[1],
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountAlreadyHasTeam,
		bodyBytes,
		statusCode,
	)
}

func TestTeamsLeave(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsLeave(
		t,
		app,
		testAccountTokens[1],
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, body.Account.TeamID, "")

	testAccounts[1] = body.Account
	testAccountTokens[1] = body.Token
}

func TestTeamLeaveHasNoTeam(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsLeave(
		t,
		app,
		testAccountTokens[1],
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountHasNoTeam,
		bodyBytes,
		statusCode,
	)
}

func TestTeamKick(t *testing.T) {
	for i := 2; i <= 3; i++ {
		_, statusCode := helpers.API_TeamsKick(
			t,
			app,
			testAccounts[i].ID,
			testAccountTokens[0],
		)

		require.Equal(t, http.StatusOK, statusCode)
	}
}

func TestTeamKickAccountNotFound(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsKick(
		t,
		app,
		"wrongiaccount",
		testAccountTokens[0],
	)

	helpers.ResponseErrorCheck(t, app,
		errmsg.AccountNotFound,
		bodyBytes,
		statusCode,
	)
}

func TestTeamsDelete(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsDelete(
		t,
		app,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Equal(t, body.Account.TeamID, "")

	testAccounts[0] = body.Account
	testAccountTokens[0] = body.Token
}

func TestTeamsCleanup(t *testing.T) {
	for i := range usersToCreate {
		err := testAccounts[i].Delete()
		if err != nil {
			fmt.Printf("failed to delete account: %v", err)
		}

		testAccounts[i] = models.Account{}
		testAccountEmails[i] = ""
		testAccountPasswords[i] = ""
		testAccountTokens[i] = ""
	}
}
