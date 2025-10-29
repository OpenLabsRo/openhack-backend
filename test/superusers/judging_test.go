package superusers

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/models"
	"backend/test/helpers"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	judgingTestSuperUserToken string
	judgingInitResp           struct {
		TeamOrderA      []string `json:"teamOrderA"`
		TeamOrderB      []string `json:"teamOrderB"`
		JudgeOrder      []string `json:"judgeOrder"`
		JudgeOffset     []int    `json:"judgeOffset"`
		JudgeMultiplier []int    `json:"judgeMultiplier"`
		JudgeBaseOrder  []int    `json:"judgeBaseOrder"`
		NumTeams        int      `json:"numTeams"`
		NumJudges       int      `json:"numJudges"`
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

	require.Len(t, judgingInitResp.TeamOrderA, numTeams, "should have %d teams in order A", numTeams)
	require.Len(t, judgingInitResp.TeamOrderB, numTeams, "should have %d teams in order B", numTeams)
	require.Len(t, judgingInitResp.JudgeOrder, numJudges, "should have %d judges in order", numJudges)
	require.Len(t, judgingInitResp.JudgeOffset, numJudges, "should have %d judge offsets", numJudges)
	require.Len(t, judgingInitResp.JudgeMultiplier, numJudges, "should have %d judge multipliers", numJudges)
	require.Len(t, judgingInitResp.JudgeBaseOrder, numJudges, "should have %d judge base order assignments", numJudges)
	require.Equal(t, numTeams, judgingInitResp.NumTeams)
	require.Equal(t, numJudges, judgingInitResp.NumJudges)

	// Validate that all multipliers are coprime to numTeams
	for i, mult := range judgingInitResp.JudgeMultiplier {
		require.Greater(t, mult, 0, "multiplier for judge %d should be positive", i)
		require.Less(t, mult, numTeams, "multiplier for judge %d should be less than numTeams", i)
		require.Equal(t, 1, gcd(mult, numTeams), "multiplier %d for judge %d should be coprime to %d", mult, i, numTeams)
	}

	fmt.Printf("Judging initialized with multipliers: %v\n", judgingInitResp.JudgeMultiplier)
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

	// Log rotation system configuration
	fmt.Printf("\n")
	fmt.Printf("========================================\n")
	fmt.Printf("    JUDGING ROTATION CONFIGURATION\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Number of Teams: %d\n", judgingInitResp.NumTeams)
	fmt.Printf("Number of Judges: %d\n", judgingInitResp.NumJudges)
	fmt.Printf("\n")
	fmt.Printf("Team Order A (first base):\n")
	for i, teamID := range judgingInitResp.TeamOrderA {
		fmt.Printf("  [%2d] %s\n", i, teamID)
	}
	fmt.Printf("\n")
	fmt.Printf("Team Order B (second base):\n")
	for i, teamID := range judgingInitResp.TeamOrderB {
		fmt.Printf("  [%2d] %s\n", i, teamID)
	}
	fmt.Printf("\n")
	fmt.Printf("Judge Order:\n")
	for i, judgeID := range judgingInitResp.JudgeOrder {
		fmt.Printf("  [%2d] %s\n", i, judgeID)
	}
	fmt.Printf("\n")
	fmt.Printf("Judge Settings:\n")
	for i := range judgingInitResp.JudgeOrder {
		baseOrderName := "A"
		if judgingInitResp.JudgeBaseOrder[i] == 1 {
			baseOrderName = "B"
		}
		fmt.Printf("  Judge %d (%s):\n", i, judgingInitResp.JudgeOrder[i])
		fmt.Printf("    Base Order: %s\n", baseOrderName)
		fmt.Printf("    Offset:     %d\n", judgingInitResp.JudgeOffset[i])
		fmt.Printf("    Multiplier: %d (coprime to %d)\n", judgingInitResp.JudgeMultiplier[i], judgingInitResp.NumTeams)
		fmt.Printf("    Formula:    teamIndex = (%d + step × %d) %% %d in order %s\n",
			judgingInitResp.JudgeOffset[i],
			judgingInitResp.JudgeMultiplier[i],
			judgingInitResp.NumTeams,
			baseOrderName)
	}
	fmt.Printf("\n")
	fmt.Printf("Expected comparisons per judge: %d\n", judgingInitResp.NumTeams-1)
	fmt.Printf("Total expected comparisons: %d\n", judgingInitResp.NumJudges*(judgingInitResp.NumTeams-1))
	fmt.Printf("========================================\n")
	fmt.Printf("\n")

	totalJudgmentsCreated := 0

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

		// Request exactly numTeams teams and create judgments
		teamsSeen := make(map[string]int)
		var previousTeamID string
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

			// Create judgment: compare current team with previous team
			if previousTeamID != "" {
				// Randomly decide winner (50/50 split)
				var winningTeamID, losingTeamID string
				if rand.Float32() < 0.5 {
					winningTeamID = teamID
					losingTeamID = previousTeamID
				} else {
					winningTeamID = previousTeamID
					losingTeamID = teamID
				}

				bodyBytes, statusCode = helpers.API_JudgeCreateJudgment(
					t,
					app,
					winningTeamID,
					losingTeamID,
					judgeToken,
				)
				require.Equal(t, http.StatusOK, statusCode)

				var judgment models.Judgment
				require.NoError(t, json.Unmarshal(bodyBytes, &judgment))
				require.NotEmpty(t, judgment.ID, "judgment should have an ID")
				require.Equal(t, winningTeamID, judgment.WinningTeamID)
				require.Equal(t, losingTeamID, judgment.LosingTeamID)
				require.Equal(t, judge.ID, judgment.JudgeID)

				fmt.Printf("Judge %d created judgment: %s beat %s\n", judgeIdx, winningTeamID, losingTeamID)
				totalJudgmentsCreated++
			}

			previousTeamID = teamID
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
	fmt.Printf("Total judgments created: %d\n", totalJudgmentsCreated)

	// Analyze judgment coverage
	fmt.Printf("\n")
	fmt.Printf("========================================\n")
	fmt.Printf("      JUDGMENT COVERAGE ANALYSIS\n")
	fmt.Printf("========================================\n")

	// Fetch all judgments from database
	cursor, err := db.Judgments.Find(db.Ctx, bson.M{})
	require.NoError(t, err)
	defer cursor.Close(db.Ctx)

	var allJudgments []models.Judgment
	err = cursor.All(db.Ctx, &allJudgments)
	require.NoError(t, err)

	fmt.Printf("Total judgments in database: %d\n", len(allJudgments))
	fmt.Printf("\n")

	// Build a map of unique pairs
	type TeamPair struct {
		TeamA string
		TeamB string
	}

	uniquePairs := make(map[TeamPair]int)
	judgmentsByJudge := make(map[string]int)

	for _, judgment := range allJudgments {
		// Count judgments per judge
		judgmentsByJudge[judgment.JudgeID]++

		// Normalize pair (always store in alphabetical order)
		pair := TeamPair{}
		if judgment.WinningTeamID < judgment.LosingTeamID {
			pair.TeamA = judgment.WinningTeamID
			pair.TeamB = judgment.LosingTeamID
		} else {
			pair.TeamA = judgment.LosingTeamID
			pair.TeamB = judgment.WinningTeamID
		}
		uniquePairs[pair]++
	}

	// Calculate statistics
	totalPossiblePairs := (numTeams * (numTeams - 1)) / 2
	uniquePairsCount := len(uniquePairs)
	coveragePercent := float64(uniquePairsCount) / float64(totalPossiblePairs) * 100

	fmt.Printf("Pair Coverage:\n")
	fmt.Printf("  Unique pairs compared:    %d\n", uniquePairsCount)
	fmt.Printf("  Total possible pairs:     %d\n", totalPossiblePairs)
	fmt.Printf("  Coverage:                 %.1f%%\n", coveragePercent)
	fmt.Printf("\n")

	// Show redundancy distribution
	redundancyDistribution := make(map[int]int)
	for _, count := range uniquePairs {
		redundancyDistribution[count]++
	}

	fmt.Printf("Redundancy Distribution:\n")
	for i := 1; i <= 10; i++ {
		if count, exists := redundancyDistribution[i]; exists {
			fmt.Printf("  Compared %d time(s):  %d pairs\n", i, count)
		}
	}
	fmt.Printf("\n")

	// Show judgments per judge
	fmt.Printf("Judgments Per Judge:\n")
	for _, judgeID := range judgingInitResp.JudgeOrder {
		count := judgmentsByJudge[judgeID]
		fmt.Printf("  %s: %d judgments\n", judgeID, count)
	}
	fmt.Printf("\n")

	// Calculate coverage quality score
	avgRedundancy := float64(len(allJudgments)) / float64(uniquePairsCount)

	fmt.Printf("Coverage Quality Metrics:\n")
	fmt.Printf("  Average redundancy:       %.2fx per pair\n", avgRedundancy)
	fmt.Printf("  Comparisons per judge:    %.1f\n", float64(len(allJudgments))/float64(numJudges))

	// Estimate max transitive distance
	// In a graph with this coverage, estimate worst-case path length
	var maxDistance int
	if coveragePercent < 20 {
		maxDistance = numTeams / 2
	} else if coveragePercent < 50 {
		maxDistance = numTeams / 3
	} else {
		maxDistance = numTeams / 4
	}

	fmt.Printf("  Est. max transitive dist: ~%d hops\n", maxDistance)
	fmt.Printf("\n")

	// Quality assessment
	fmt.Printf("Assessment:\n")
	if coveragePercent >= 60 {
		fmt.Printf("  ✓ EXCELLENT coverage for Gavel (>60%%)\n")
	} else if coveragePercent >= 40 {
		fmt.Printf("  ✓ GOOD coverage for Gavel (40-60%%)\n")
	} else if coveragePercent >= 20 {
		fmt.Printf("  ⚠ FAIR coverage for Gavel (20-40%%)\n")
	} else {
		fmt.Printf("  ✗ LOW coverage for Gavel (<20%%)\n")
	}

	if avgRedundancy >= 2.0 {
		fmt.Printf("  ✓ HIGH redundancy reduces noise\n")
	} else if avgRedundancy >= 1.5 {
		fmt.Printf("  ✓ GOOD redundancy\n")
	} else {
		fmt.Printf("  ⚠ LOW redundancy - more variance\n")
	}

	fmt.Printf("========================================\n")
	fmt.Printf("\n")

	// Analyze collision rate (judges at same team at same step)
	fmt.Printf("========================================\n")
	fmt.Printf("     COLLISION ANALYSIS\n")
	fmt.Printf("========================================\n")

	// Build a map of which teams each judge visited at each step
	type JudgeVisit struct {
		JudgeID string
		Step    int
		TeamID  string
	}

	// We need to reconstruct judge paths from the init configuration
	// For each judge, calculate their path based on offset, multiplier, and base order
	judgePathsByStep := make(map[int]map[string][]string) // step -> teamID -> []judgeIDs

	for judgeIdx, judgeID := range judgingInitResp.JudgeOrder {
		offset := judgingInitResp.JudgeOffset[judgeIdx]
		mult := judgingInitResp.JudgeMultiplier[judgeIdx]
		baseOrderIdx := judgingInitResp.JudgeBaseOrder[judgeIdx]

		var teamOrder []string
		if baseOrderIdx == 0 {
			teamOrder = judgingInitResp.TeamOrderA
		} else {
			teamOrder = judgingInitResp.TeamOrderB
		}

		// Calculate this judge's path
		for step := 0; step < numTeams; step++ {
			teamIndex := (offset + step*mult) % numTeams
			teamID := teamOrder[teamIndex]

			if judgePathsByStep[step] == nil {
				judgePathsByStep[step] = make(map[string][]string)
			}
			judgePathsByStep[step][teamID] = append(judgePathsByStep[step][teamID], judgeID)
		}
	}

	// Analyze collisions at each step
	totalCollisions := 0
	maxCollisionAtTeam := 0
	collisionSteps := 0

	fmt.Printf("Step-by-Step Collision Analysis:\n")
	for step := 0; step < numTeams; step++ {
		teamsAtStep := judgePathsByStep[step]
		stepCollisions := 0
		stepMaxAtTeam := 0

		for _, judgesAtTeam := range teamsAtStep {
			if len(judgesAtTeam) > 1 {
				stepCollisions += len(judgesAtTeam) - 1
				if len(judgesAtTeam) > stepMaxAtTeam {
					stepMaxAtTeam = len(judgesAtTeam)
				}
				if len(judgesAtTeam) > maxCollisionAtTeam {
					maxCollisionAtTeam = len(judgesAtTeam)
				}
			}
		}

		if stepCollisions > 0 {
			collisionSteps++
			fmt.Printf("  Step %2d: %d collision(s), max %d judges at one team\n", step, stepCollisions, stepMaxAtTeam)
		}
		totalCollisions += stepCollisions
	}

	avgCollisionsPerStep := float64(totalCollisions) / float64(numTeams)

	fmt.Printf("\n")
	fmt.Printf("Collision Summary:\n")
	fmt.Printf("  Total collisions:         %d\n", totalCollisions)
	fmt.Printf("  Steps with collisions:    %d / %d\n", collisionSteps, numTeams)
	fmt.Printf("  Avg collisions per step:  %.2f\n", avgCollisionsPerStep)
	fmt.Printf("  Max judges at one team:   %d\n", maxCollisionAtTeam)
	fmt.Printf("\n")

	// Assessment
	collisionRate := float64(totalCollisions) / float64(numJudges*numTeams) * 100
	fmt.Printf("Assessment:\n")
	fmt.Printf("  Collision rate: %.1f%% of all judge-step pairs\n", collisionRate)
	if maxCollisionAtTeam <= 2 {
		fmt.Printf("  ✓ LOW congestion (max 2 judges per team)\n")
	} else if maxCollisionAtTeam <= 3 {
		fmt.Printf("  ⚠ MODERATE congestion (max 3 judges per team)\n")
	} else {
		fmt.Printf("  ✗ HIGH congestion (max %d judges per team)\n", maxCollisionAtTeam)
	}

	fmt.Printf("========================================\n")
	fmt.Printf("\n")
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

// gcd computes the greatest common divisor using Euclidean algorithm
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
