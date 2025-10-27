package superusers

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/models"
	"backend/test/helpers"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	judgingTestSuperUserToken string
	judgingInitResp           struct {
		TeamOrder   []string `json:"teamOrder"`
		JudgeOrder  []string `json:"judgeOrder"`
		JudgeOffset []int    `json:"judgeOffset"`
		NumTeams    int      `json:"numTeams"`
		NumJudges   int      `json:"numJudges"`
	}
	createdJudges          []models.Judge
	createdJudgingAccounts []models.Account
	createdJudgingTeams    []models.Team
)

const (
	numParticipants = 24
	numJudges       = 6
	numTeams        = numParticipants / teamSize
	teamSize        = 4
)

// var (
// 	judgingEnvRootFlag    = flag.String("judging-env-root", "", "directory containing environment files")
// 	judgingAppVersionFlag = flag.String("judging-app-version", "", "application version override")
// )

// // TestJudgingSetup initializes app for judging tests - runs first
// func TestJudgingSetup(t *testing.T) {
// 	flag.Parse()
// 	app = internal.SetupApp("test", *judgingEnvRootFlag, *judgingAppVersionFlag)
// 	helpers.ResetTestCache()
// 	helpers.ResetTestEvents()
// 	require.NotNil(t, app, "app should be initialized")
// }

// TestJudgingSetupSuperUser logs in as superuser
func TestJudgingSetupSuperUser(t *testing.T) {
	require.NotNil(t, app, "app should be initialized in TestJudgingSetup")

	bodyBytes, statusCode := helpers.API_SuperUsersAuthLogin(
		t,
		app,
		env.SUPERUSER_USERNAME,
		env.SUPERUSER_PASSWORD,
	)
	require.Equal(t, http.StatusOK, statusCode)

	var loginResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &loginResp))
	judgingTestSuperUserToken = loginResp.Token

	fmt.Printf("Superuser token acquired\n")

	// Reset flags by executing initialize flagstage
	_, statusCode = helpers.API_SuperUsersFlagStagesExecute(
		t,
		app,
		"0",
		judgingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)
}

// TestJudgingCreateJudges creates judges via API
func TestJudgingCreateJudges(t *testing.T) {
	require.NotEmpty(t, judgingTestSuperUserToken, "superuser token should be initialized")

	for i := range numJudges {
		judgeID := fmt.Sprintf("judge_%d", i)
		judgeName := fmt.Sprintf("Judge %d", i)

		bodyBytes, statusCode := helpers.API_SuperUsersJudgingCreate(
			t,
			app,
			judgeID,
			judgeName,
			judgingTestSuperUserToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var judge models.Judge
		require.NoError(t, json.Unmarshal(bodyBytes, &judge))
		createdJudges = append(createdJudges, judge)

		fmt.Printf("Created judge: %s (name: %s)\n", judge.ID, judge.Name)
	}

	require.Len(t, createdJudges, numJudges)
	fmt.Printf("Created %d judges\n", len(createdJudges))
}

// TestJudgingCreateAccounts creates accounts via API
func TestJudgingCreateAccounts(t *testing.T) {
	require.NotEmpty(t, judgingTestSuperUserToken, "superuser token should be initialized")

	for i := range numParticipants {
		email := fmt.Sprintf("test_account_%d@test.com", i)
		name := fmt.Sprintf("Test Account %d", i)

		bodyBytes, statusCode := helpers.API_SuperUsersParticipantsInitialize(
			t,
			app,
			email,
			name,
			judgingTestSuperUserToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var acc models.Account
		require.NoError(t, json.Unmarshal(bodyBytes, &acc))
		createdJudgingAccounts = append(createdJudgingAccounts, acc)
	}

	require.Len(t, createdJudgingAccounts, numParticipants)
	fmt.Printf("Created %d accounts\n", len(createdJudgingAccounts))
}

// TestJudgingFormTeams combines accounts into teams
func TestJudgingFormTeams(t *testing.T) {
	require.Len(t, createdJudgingAccounts, numParticipants, "should have exactly %d accounts", numParticipants)

	teamCount := numParticipants / teamSize

	for i := range teamCount {
		team := models.Team{
			ID:      fmt.Sprintf("test_team_%d", i),
			Name:    fmt.Sprintf("Test Team %d", i+1),
			Members: []string{},
			Deleted: false,
		}

		// Add members to the team
		for j := range teamSize {
			accountIdx := i*teamSize + j
			team.Members = append(team.Members, createdJudgingAccounts[accountIdx].ID)
		}

		_, err := db.Teams.InsertOne(db.Ctx, team)
		require.NoError(t, err)

		createdJudgingTeams = append(createdJudgingTeams, team)
	}

	require.Len(t, createdJudgingTeams, numTeams, "should have %d teams", numTeams)
	fmt.Printf("Created %d teams of %d members each\n", len(createdJudgingTeams), teamSize)
}

// TestJudgingInitialize runs the judging initialization
func TestJudgingInitialize(t *testing.T) {
	require.NotEmpty(t, judgingTestSuperUserToken, "superuser token should be initialized")
	require.Len(t, createdJudges, numJudges, "should have %d judges", numJudges)
	require.Len(t, createdJudgingTeams, numTeams, "should have %d teams", numTeams)

	bodyBytes, statusCode := helpers.API_SuperUsersJudgingInit(
		t,
		app,
		judgingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)

	require.NoError(t, json.Unmarshal(bodyBytes, &judgingInitResp))

	require.Len(t, judgingInitResp.TeamOrder, numTeams, "should have %d teams in order", numTeams)
	require.Len(t, judgingInitResp.JudgeOrder, numJudges, "should have %d judges in order", numJudges)
	require.Len(t, judgingInitResp.JudgeOffset, numJudges, "should have %d judge offsets", numJudges)
	require.Equal(t, numTeams, judgingInitResp.NumTeams)
	require.Equal(t, numJudges, judgingInitResp.NumJudges)

	fmt.Printf("Judging initialized\n")
}

// TestJudgingRotation tests the complete judging rotation for each judge
func TestJudgingRotation(t *testing.T) {
	require.NotEmpty(t, judgingTestSuperUserToken, "superuser token should be initialized")
	require.Len(t, createdJudges, numJudges, "should have %d judges", numJudges)
	require.Len(t, createdJudgingTeams, numTeams, "should have %d teams", numTeams)

	// Enable Stage 6 for judging operations
	_, statusCode := helpers.API_SuperUsersFlagStagesExecute(
		t,
		app,
		"6",
		judgingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)

	// For each judge, do the full rotation
	for judgeIdx, judge := range createdJudges {
		fmt.Printf("Judge %d: %s\n", judgeIdx, judge.ID)

		var bodyBytes []byte
		var statusCode int

		// Get connect token
		bodyBytes, statusCode = helpers.API_SuperUsersJudgingConnect(
			t,
			app,
			judge.ID,
			judgingTestSuperUserToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var connectResp struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(bodyBytes, &connectResp))
		require.NotEmpty(t, connectResp.Token, "should receive connect token")

		// Upgrade to full judge token
		bodyBytes, statusCode = helpers.API_JudgeUpgrade(
			t,
			app,
			connectResp.Token,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var upgradeResp struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(bodyBytes, &upgradeResp))
		judgeToken := upgradeResp.Token
		require.NotEmpty(t, judgeToken, "should receive full judge token")

		fmt.Printf("Judge %d upgraded\n", judgeIdx)

		// Request exactly numTeams teams
		teamsSeen := make(map[string]int)
		for _ = range judgingInitResp.NumTeams {
			bodyBytes, statusCode = helpers.API_JudgeNextTeam(
				t,
				app,
				judgeToken,
			)
			require.Equal(t, http.StatusOK, statusCode)

			var nextTeamResp struct {
				TeamID string `json:"teamID"`
			}
			require.NoError(t, json.Unmarshal(bodyBytes, &nextTeamResp))

			teamID := nextTeamResp.TeamID
			require.NotEmpty(t, teamID, "should receive a team ID")

			teamsSeen[teamID]++

			fmt.Printf("Judge %d checked %s\n", judgeIdx, teamID)

			// Get team info to verify it's valid
			bodyBytes, statusCode = helpers.API_JudgeTeamInfo(
				t,
				app,
				teamID,
				judgeToken,
			)
			require.Equal(t, http.StatusOK, statusCode)

			var teamInfo models.Team
			require.NoError(t, json.Unmarshal(bodyBytes, &teamInfo))
			require.Equal(t, teamID, teamInfo.ID, "team info should match requested team")
		}

		// Verify that judge saw each team exactly once
		require.Equal(t, judgingInitResp.NumTeams, len(teamsSeen), "judge should have seen exactly %d teams", judgingInitResp.NumTeams)
		for teamID, count := range teamsSeen {
			require.Equal(t, 1, count, "judge should have seen team %s exactly once, but saw it %d times", teamID, count)
		}

		// One more request should return judging finished
		bodyBytes, statusCode = helpers.API_JudgeNextTeam(
			t,
			app,
			judgeToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var finishedResp struct {
			Message string `json:"message"`
		}
		require.NoError(t, json.Unmarshal(bodyBytes, &finishedResp))
		require.Equal(t, "judging finished", finishedResp.Message)

		fmt.Printf("Judge %d success\n", judgeIdx)
	}

	fmt.Printf("\n=== All judges completed their rotations ===\n")
}

// TestJudgingCleanup cleans up test data
func TestJudgingCleanup(t *testing.T) {
	// Delete created judges
	for _, judge := range createdJudges {
		err := judge.Delete()
		require.NoError(t, err)
	}

	// Delete created accounts
	for _, acc := range createdJudgingAccounts {
		_, err := db.Accounts.DeleteOne(db.Ctx, bson.M{"id": acc.ID})
		require.NoError(t, err)
	}

	// Delete created teams
	for _, team := range createdJudgingTeams {
		_, err := db.Teams.DeleteOne(db.Ctx, bson.M{"id": team.ID})
		require.NoError(t, err)
	}

	// Delete all judgments
	_, err := db.Judgments.DeleteMany(db.Ctx, bson.M{})
	require.NoError(t, err)
}
