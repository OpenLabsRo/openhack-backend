package test

import (
	"backend/internal"
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/test/helpers"
	"context"
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
	"go.mongodb.org/mongo-driver/bson"
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

	// package-level flags
	envRootFlag    = flag.String("env-root", "", "directory containing environment files")
	appVersionFlag = flag.String("app-version", "", "application version override")
)

func TestMain(m *testing.M) {
	flag.Parse()
	app = internal.SetupApp("test", *envRootFlag, *appVersionFlag)
	helpers.ResetTestCache()
	helpers.ResetTestEvents()
	os.Exit(m.Run())
}

func TestTeamsPing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/teams/meta/ping", nil)
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
	bodyBytes, statusCode := helpers.API_SuperUsersAuthLogin(
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

		_, statusCode = helpers.API_SuperUsersParticipantsInitialize(
			t,
			app,
			fmt.Sprintf("teamstesting%v@example.com", i),
			fmt.Sprintf("Accounts Testing %v", i),
			body.Token,
		)
		require.Equal(t, http.StatusOK, statusCode)

		bodyBytes, statusCode = helpers.API_AccountsAuthRegister(
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

func TestTeamsChangeNameHasNoTeam(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsChangeName(
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

func TestTeamsNameChange(t *testing.T) {
	newName := "Changed Team Name"
	bodyBytes, statusCode := helpers.API_TeamsChangeName(
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

func TestTeamsChangeTable(t *testing.T) {
	newTable := "Table 42"
	bodyBytes, statusCode := helpers.API_TeamsChangeTable(
		t,
		app,
		newTable,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	var team models.Team
	require.NoError(t, json.Unmarshal(bodyBytes, &team))
	require.Equal(t, testTeamID, team.ID)
	require.Equal(t, newTable, team.Table)

	var stored models.Team
	require.NoError(t, db.Teams.FindOne(db.Ctx, bson.M{"id": testTeamID}).Decode(&stored))
	require.Equal(t, newTable, stored.Table)
}

func TestTeamsSubmissionsChangeName(t *testing.T) {
	newName := "Changed Submission Name"
	bodyBytes, statusCode := helpers.API_TeamsSubmissionsChangeName(
		t,
		app,
		newName,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	tempTeam := models.Team{}
	err := json.Unmarshal(bodyBytes, &tempTeam)
	require.NoError(t, err)

	require.Equal(t, tempTeam.Submission.Name, newName)
}

func TestTeamsSubmissionChangeDesc(t *testing.T) {
	newDesc := "Changed Description"
	bodyBytes, statusCode := helpers.API_TeamsSubmissionsChangeDesc(
		t,
		app,
		newDesc,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	tempTeam := models.Team{}
	err := json.Unmarshal(bodyBytes, &tempTeam)
	require.NoError(t, err)

	require.Equal(t, tempTeam.Submission.Desc, newDesc)
}

func TestTeamsSubmissionChangeRepo(t *testing.T) {
	newRepo := "Changed Repo"
	bodyBytes, statusCode := helpers.API_TeamsSubmissionsChangeRepo(
		t,
		app,
		newRepo,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	tempTeam := models.Team{}
	err := json.Unmarshal(bodyBytes, &tempTeam)
	require.NoError(t, err)

	require.Equal(t, tempTeam.Submission.Repo, newRepo)
}

func TestTeamsSubmissionChangePres(t *testing.T) {
	newPres := "Changed Pres"
	bodyBytes, statusCode := helpers.API_TeamsSubmissionsChangePres(
		t,
		app,
		newPres,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	tempTeam := models.Team{}
	err := json.Unmarshal(bodyBytes, &tempTeam)
	require.NoError(t, err)

	require.Equal(t, tempTeam.Submission.Pres, newPres)
}

func TestTeamsJoinNotFound(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsMembersJoin(
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
		bodyBytes, statusCode := helpers.API_TeamsMembersJoin(
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

func TestTeamsGetMembers(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsGetMembers(
		t,
		app,
		testAccountTokens[0],
	)

	require.Equal(t, http.StatusOK, statusCode)

	var body []models.Account
	err := json.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	require.Len(t, body, 4)
}

func TestTeamsJoinTeamFull(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsMembersJoin(
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
	bodyBytes, statusCode := helpers.API_TeamsMembersJoin(
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
	bodyBytes, statusCode := helpers.API_TeamsMembersLeave(
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
	bodyBytes, statusCode := helpers.API_TeamsMembersLeave(
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
		_, statusCode := helpers.API_TeamsMembersKick(
			t,
			app,
			testAccounts[i].ID,
			testAccountTokens[0],
		)

		require.Equal(t, http.StatusOK, statusCode)
	}
}

func TestTeamKickAccountNotFound(t *testing.T) {
	bodyBytes, statusCode := helpers.API_TeamsMembersKick(
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
	// delete the team
	db.Teams.DeleteOne(context.Background(), bson.M{"id": testTeamID})

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
