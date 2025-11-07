package superusers

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"backend/test/helpers"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	pairingTestSuperUserToken string
	pairingInitResp           struct {
		Message            string  `json:"message"`
		NumTeams           int     `json:"numTeams"`
		NumJudges          int     `json:"numJudges"`
		NumPairGroups      int     `json:"numPairGroups"`
		NumSteps           int     `json:"numSteps"`
		Collisions         int     `json:"collisions"`
		GavelScore         float64 `json:"gavelScore"`
		UniquePairs        int     `json:"uniquePairs"`
		TotalPossiblePairs int     `json:"totalPossiblePairs"`
		AverageRedundancy  float64 `json:"averageRedundancy"`
	}
	pairingInitMatrix struct {
		Steps         int        `json:"steps"`
		Groups        int        `json:"groups"`
		Teams         int        `json:"teams"`
		Assignments   int        `json:"assignments"`
		BlankCells    int        `json:"blankCells"`
		Collisions    int        `json:"collisions"`
		GavelScore    float64    `json:"gavelScore"`
		UniquePairs   int        `json:"uniquePairs"`
		TotalPairs    int        `json:"totalPairs"`
		AvgRedundancy float64    `json:"avgRedundancy"`
		Matrix        [][]string `json:"matrix"`
	}
	createdPairingJudges     []models.Judge
	createdPairingAccounts   []models.Account
	createdPairingTeams      []models.Team
	createdPairingJudgesByID map[string]*models.Judge
)

// Configuration for judge pairing test (customize here)
const (
	numPairingParticipants = 52
	numPairingTeams        = 16
	numSoloJudges          = 8
	numPairJudges          = 4  // number of pairs (each with 2 judges)
	numTrioJudges          = 1  // number of trios (each with 3 judges)
	numTeamsSize3          = 12 // teams with 3 members
	numTeamsSize4          = 4  // teams with 4 members
)

// Calculated values based on configuration
var (
	totalPairingJudges      = numSoloJudges + (numPairJudges * 2) + (numTrioJudges * 3) // 5 + 6 + 3 = 14
	totalPairingJudgeGroups = numSoloJudges + numPairJudges + numTrioJudges             // 5 + 3 + 1 = 9
)

// TestJudgingPairsSetupSuperUser logs in as superuser for pairing tests
func TestJudgingPairsSetupSuperUser(t *testing.T) {
	require.NotNil(t, app, "app should be initialized")

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
	pairingTestSuperUserToken = loginResp.Token

	fmt.Printf("\n=== JUDGE PAIRING TEST SUITE ===\n")
	fmt.Printf("Scenario: %d solo + %d pairs of 2 + %d trio(s) = %d judges, %d teams, %d participants\n",
		numSoloJudges, numPairJudges, numTrioJudges, totalPairingJudges, numPairingTeams, numPairingParticipants)
	fmt.Printf("Superuser token acquired\n")

	// Reset flags
	_, statusCode = helpers.API_SuperUsersFlagStagesExecute(
		t,
		app,
		"0",
		pairingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)
}

// TestJudgingPairsCreateJudgesWithPairs creates judges with pair attributes
func TestJudgingPairsCreateJudgesWithPairs(t *testing.T) {
	require.NotEmpty(t, pairingTestSuperUserToken, "superuser token should be initialized")

	createdPairingJudgesByID = make(map[string]*models.Judge)
	judgeIdx := 0

	// Create solo judges (each solo judge is their own group)
	fmt.Printf("\nCreating %d solo judges:\n", numSoloJudges)
	for solo := range numSoloJudges {
		judgeID := fmt.Sprintf("judge_%d", judgeIdx)
		judgeName := fmt.Sprintf("Judge Solo %d", solo+1)

		bodyBytes, statusCode := helpers.API_SuperUsersJudgingCreate(
			t,
			app,
			judgeID,
			judgeName,
			pairingTestSuperUserToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var judge models.Judge
		require.NoError(t, json.Unmarshal(bodyBytes, &judge))

		// Solo judges each get their own group ID
		groupID := solo
		judge.Pair = fmt.Sprintf("group_%02d", groupID)
		_, err := db.Judges.UpdateOne(db.Ctx, bson.M{"id": judge.ID}, bson.M{"$set": bson.M{"pair": judge.Pair}})
		require.NoError(t, err)

		// Fetch updated judge
		err = judge.Get()
		require.NoError(t, err)

		createdPairingJudges = append(createdPairingJudges, judge)
		j := judge
		createdPairingJudgesByID[judge.ID] = &j

		fmt.Printf("  Created: %s (group='%s')\n", judge.ID, judge.Pair)
		judgeIdx++
	}

	// Create pairs of 2 judges
	fmt.Printf("\nCreating %d pairs of 2 judges:\n", numPairJudges)
	for pair := range numPairJudges {
		groupID := numSoloJudges + pair
		pairAttr := fmt.Sprintf("group_%02d", groupID)
		for member := range 2 {
			judgeID := fmt.Sprintf("judge_%d", judgeIdx)
			judgeName := fmt.Sprintf("Judge Pair %d-%d", pair+1, member+1)

			bodyBytes, statusCode := helpers.API_SuperUsersJudgingCreate(
				t,
				app,
				judgeID,
				judgeName,
				pairingTestSuperUserToken,
			)
			require.Equal(t, http.StatusOK, statusCode)

			var judge models.Judge
			require.NoError(t, json.Unmarshal(bodyBytes, &judge))

			// Update with pair attribute
			judge.Pair = pairAttr
			_, err := db.Judges.UpdateOne(db.Ctx, bson.M{"id": judge.ID}, bson.M{"$set": bson.M{"pair": pairAttr}})
			require.NoError(t, err)

			// Fetch updated judge
			err = judge.Get()
			require.NoError(t, err)

			createdPairingJudges = append(createdPairingJudges, judge)
			j := judge
			createdPairingJudgesByID[judge.ID] = &j

			fmt.Printf("  Created: %s (pair='%s')\n", judge.ID, judge.Pair)
			judgeIdx++
		}
	}

	// Create trios of 3 judges
	fmt.Printf("\nCreating %d trio(s) of 3 judges:\n", numTrioJudges)
	for trio := range numTrioJudges {
		groupID := numSoloJudges + numPairJudges + trio
		pairAttr := fmt.Sprintf("group_%02d", groupID)
		for member := range 3 {
			judgeID := fmt.Sprintf("judge_%d", judgeIdx)
			judgeName := fmt.Sprintf("Judge Trio %d-%d", trio+1, member+1)

			bodyBytes, statusCode := helpers.API_SuperUsersJudgingCreate(
				t,
				app,
				judgeID,
				judgeName,
				pairingTestSuperUserToken,
			)
			require.Equal(t, http.StatusOK, statusCode)

			var judge models.Judge
			require.NoError(t, json.Unmarshal(bodyBytes, &judge))

			// Update with pair attribute
			judge.Pair = pairAttr
			_, err := db.Judges.UpdateOne(db.Ctx, bson.M{"id": judge.ID}, bson.M{"$set": bson.M{"pair": pairAttr}})
			require.NoError(t, err)

			// Fetch updated judge
			err = judge.Get()
			require.NoError(t, err)

			createdPairingJudges = append(createdPairingJudges, judge)
			j := judge
			createdPairingJudgesByID[judge.ID] = &j

			fmt.Printf("  Created: %s (pair='%s')\n", judge.ID, judge.Pair)
			judgeIdx++
		}
	}

	require.Len(t, createdPairingJudges, totalPairingJudges)
	fmt.Printf("\nTotal judges created: %d\n", len(createdPairingJudges))
}

// TestJudgingPairsCreateAccounts creates accounts for teams
func TestJudgingPairsCreateAccounts(t *testing.T) {
	require.NotEmpty(t, pairingTestSuperUserToken, "superuser token should be initialized")

	for i := range numPairingParticipants {
		email := fmt.Sprintf("test_pair_account_%d@test.com", i)
		firstName := "Test Pair Account"
		lastName := fmt.Sprintf("%d", i)

		bodyBytes, statusCode := helpers.API_SuperUsersParticipantsInitialize(
			t,
			app,
			email,
			firstName,
			lastName,
			pairingTestSuperUserToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var acc models.Account
		require.NoError(t, json.Unmarshal(bodyBytes, &acc))
		createdPairingAccounts = append(createdPairingAccounts, acc)
	}

	require.Len(t, createdPairingAccounts, numPairingParticipants)
	fmt.Printf("Created %d accounts\n", len(createdPairingAccounts))
}

// TestJudgingPairsFormTeams combines accounts into teams
func TestJudgingPairsFormTeams(t *testing.T) {
	require.Len(t, createdPairingAccounts, numPairingParticipants)

	// Create teams with configured sizes
	accountIdx := 0
	for i := range numPairingTeams {
		team := models.Team{
			ID:      fmt.Sprintf("test_pair_team_%d", i),
			Name:    fmt.Sprintf("Test Pair Team %d", i+1),
			Members: []string{},
			Deleted: false,
		}

		// Determine team size based on configuration
		teamSize := 3
		if i >= numTeamsSize3 {
			teamSize = 4
		}

		for j := 0; j < teamSize; j++ {
			if accountIdx < len(createdPairingAccounts) {
				team.Members = append(team.Members, createdPairingAccounts[accountIdx].ID)
				accountIdx++
			}
		}

		_, err := db.Teams.InsertOne(db.Ctx, team)
		require.NoError(t, err)

		createdPairingTeams = append(createdPairingTeams, team)
	}

	require.Len(t, createdPairingTeams, numPairingTeams)
	fmt.Printf("Created %d teams: %d teams of 3 + %d teams of 4 members\n", len(createdPairingTeams), numTeamsSize3, numTeamsSize4)
}

// TestJudgingPairsInitialize runs the new pairing-based judging initialization
func TestJudgingPairsInitialize(t *testing.T) {
	require.NotEmpty(t, pairingTestSuperUserToken, "superuser token should be initialized")
	require.Len(t, createdPairingJudges, totalPairingJudges)
	require.Len(t, createdPairingTeams, numPairingTeams)

	bodyBytes, statusCode := helpers.API_SuperUsersJudgingInit(
		t,
		app,
		pairingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)

	require.NoError(t, json.Unmarshal(bodyBytes, &pairingInitResp))

	fmt.Printf("\n========================================\n")
	fmt.Printf("      INITIALIZATION RESPONSE\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Message: %s\n", pairingInitResp.Message)
	fmt.Printf("Num Teams: %d\n", pairingInitResp.NumTeams)
	fmt.Printf("Num Judges: %d\n", pairingInitResp.NumJudges)
	fmt.Printf("Num Pair Groups: %d\n", pairingInitResp.NumPairGroups)
	fmt.Printf("Num Steps: %d\n", pairingInitResp.NumSteps)
	fmt.Printf("Collisions: %d\n", pairingInitResp.Collisions)
	fmt.Printf("Gavel Score: %.1f%%\n", pairingInitResp.GavelScore)
	fmt.Printf("Unique Pairs Compared: %d / %d\n", pairingInitResp.UniquePairs, pairingInitResp.TotalPossiblePairs)
	fmt.Printf("Average Redundancy: %.2fx\n", pairingInitResp.AverageRedundancy)
	fmt.Printf("\n")

	require.Equal(t, numPairingTeams, pairingInitResp.NumTeams)
	require.Equal(t, totalPairingJudges, pairingInitResp.NumJudges)
	require.Equal(t, totalPairingJudgeGroups, pairingInitResp.NumPairGroups)
	require.Equal(t, pairingInitResp.Collisions, 0) // No collisions allowed

	// Fetch matrix from settings
	matrixSetting := &models.Setting{Name: models.SettingJudgeInitMatrix}
	errStatus := matrixSetting.Get()
	require.Equal(t, errmsg.EmptyStatusError, errStatus)

	require.NoError(t, json.Unmarshal([]byte(matrixSetting.Value.(string)), &pairingInitMatrix))

	// Log matrix in compact format
	fmt.Printf("\n========================================\n")
	fmt.Printf("      LATIN RECTANGLE ASSIGNMENTS\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Per-group team assignments by step:\n")
	fmt.Printf("\n")

	for groupIdx := 0; groupIdx < pairingInitMatrix.Groups; groupIdx++ {
		assignments := []string{}
		blanks := 0
		for step := 0; step < pairingInitMatrix.Steps; step++ {
			teamID := pairingInitMatrix.Matrix[step][groupIdx]
			if teamID == "" {
				blanks++

				assignments = append(assignments, fmt.Sprintf("S%d:%s", step, fmt.Sprintf("%2s", "x")))
			} else {

				assignments = append(assignments,
					fmt.Sprintf("S%d:%s", step,
						fmt.Sprintf("%2s",
							teamID[strings.LastIndex(teamID, "_")+1:],
						),
					),
				)
			}
		}
		fmt.Printf("Group %2d: %v (rest: %d steps)\n", groupIdx, assignments, blanks)
	}
	fmt.Printf("\n")

	fmt.Printf("========================================\n")
	fmt.Printf("      MATRIX METRICS\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Total Cells: %d\n", pairingInitMatrix.Steps*pairingInitMatrix.Groups)
	fmt.Printf("Assignments: %d\n", pairingInitMatrix.Assignments)
	fmt.Printf("Blank Cells: %d\n", pairingInitMatrix.BlankCells)
	fmt.Printf("Collisions: %d\n", pairingInitMatrix.Collisions)
	fmt.Printf("Coverage: %.2f%%\n", float64(pairingInitMatrix.Assignments)/float64(pairingInitMatrix.Steps*pairingInitMatrix.Groups)*100)
	fmt.Printf("\n")

	fmt.Printf("========================================\n")
	fmt.Printf("      COLLISION VERIFICATION\n")
	fmt.Printf("========================================\n")
	if pairingInitMatrix.Collisions == 0 {
		fmt.Printf("✓ ZERO COLLISIONS VERIFIED\n")
		fmt.Printf("  No team appears more than once per step\n")
		fmt.Printf("  Latin rectangle constraint: SATISFIED\n")
	} else {
		fmt.Printf("✗ COLLISIONS DETECTED: %d\n", pairingInitMatrix.Collisions)
	}
	fmt.Printf("\n")

	fmt.Printf("========================================\n")
	fmt.Printf("      GAVEL SCORE (PAIRWISE COVERAGE)\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Unique pairs compared:    %d\n", pairingInitMatrix.UniquePairs)
	fmt.Printf("Total possible pairs:     %d\n", pairingInitMatrix.TotalPairs)
	fmt.Printf("Gavel Score:              %.1f%%\n", pairingInitMatrix.GavelScore)
	fmt.Printf("Average redundancy:       %.2fx per pair\n", pairingInitMatrix.AvgRedundancy)
	fmt.Printf("\n")

	if pairingInitMatrix.GavelScore >= 60 {
		fmt.Printf("Assessment: [EXCELLENT] coverage for Gavel (>60%%)\n")
	} else if pairingInitMatrix.GavelScore >= 40 {
		fmt.Printf("Assessment: [GOOD] coverage for Gavel (40-60%%)\n")
	} else if pairingInitMatrix.GavelScore >= 20 {
		fmt.Printf("Assessment: [FAIR] coverage for Gavel (20-40%%)\n")
	} else {
		fmt.Printf("Assessment: [LOW] coverage for Gavel (<20%%)\n")
	}
	fmt.Printf("\n")
}

// TestJudgingPairsFullSimulation runs a complete simulation through all steps
func TestJudgingPairsFullSimulation(t *testing.T) {
	require.NotEmpty(t, pairingTestSuperUserToken, "superuser token should be initialized")
	require.Len(t, createdPairingJudges, totalPairingJudges)
	require.Len(t, createdPairingTeams, numPairingTeams)

	// Enable judging stage
	_, statusCode := helpers.API_SuperUsersFlagStagesExecute(
		t,
		app,
		"6",
		pairingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)

	fmt.Printf("\n========================================\n")
	fmt.Printf("    FULL JUDGING SIMULATION\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Testing all %d steps for %d judges (%d solo + %d pairs + %d trios) across %d teams\n\n",
		pairingInitMatrix.Steps, len(createdPairingJudges), numSoloJudges, numPairJudges, numTrioJudges, len(createdPairingTeams))

	judgeTokens := make(map[string]string)

	// Upgrade all judges and track their tokens
	fmt.Printf("Upgrading judges...\n")
	for _, judge := range createdPairingJudges {
		bodyBytes, statusCode := helpers.API_SuperUsersJudgingConnect(
			t,
			app,
			judge.ID,
			pairingTestSuperUserToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var connectResp struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(bodyBytes, &connectResp))

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
		judgeTokens[judge.ID] = upgradeResp.Token
	}
	fmt.Printf("All %d judges upgraded\n\n", len(createdPairingJudges))

	// Fetch judge to group mapping
	judgeToGroupIndexSetting := &models.Setting{Name: models.SettingJudgeToGroupIndex}
	errStatus := judgeToGroupIndexSetting.Get()
	require.Equal(t, errmsg.EmptyStatusError, errStatus)

	var judgeToGroupIdx map[string]int
	require.NoError(t, json.Unmarshal([]byte(judgeToGroupIndexSetting.Value.(string)), &judgeToGroupIdx))

	// Build reverse mapping: group -> judges for debugging
	groupToJudges := make(map[int][]string)
	for judgeID, groupIdx := range judgeToGroupIdx {
		groupToJudges[groupIdx] = append(groupToJudges[groupIdx], judgeID)
	}

	// Identify unreliable judges in the trio group with opposite biases
	// Judge 0: always votes for PRIOR team (left bias)
	// Judge 1: always votes for CURRENT team (right bias)
	// Judge 2: control with random voting
	unreliableJudges := make(map[string]string)   // judgeID -> "prior" or "current"
	trioGroupIdx := numSoloJudges + numPairJudges // Should be 12
	if judges, ok := groupToJudges[trioGroupIdx]; ok && len(judges) >= 2 {
		// Make first judge vote for prior (left bias)
		unreliableJudges[judges[0]] = "prior"
		// Make second judge vote for current (right bias)
		unreliableJudges[judges[1]] = "current"
		fmt.Printf("\n⚠ UNRELIABLE JUDGES DETECTED FOR TESTING:\n")
		fmt.Printf("  %s - will always vote for PRIOR team (left bias)\n", judges[0])
		fmt.Printf("  %s - will always vote for CURRENT team (right bias)\n", judges[1])
		if len(judges) > 2 {
			fmt.Printf("  %s - acts as control (normal random voting)\n", judges[2])
		}
		fmt.Printf("\n")
	}

	// Track statistics
	type StepStats struct {
		AssignedGroups   map[int]string // groupIdx -> teamID
		RestingGroups    []int
		JudgeAssignments map[string]string // judgeID -> teamID
	}

	allStepsStats := make(map[int]*StepStats)
	globalTeamAssignments := make(map[string]int) // teamID -> count
	globalGroupTeamPairs := make(map[string]bool) // "groupIdx-teamID" -> true if assigned
	totalCollisionsAcrossSteps := 0
	judgeTeamHistory := make(map[string][]string) // judgeID -> list of team IDs in order
	totalJudgmentsCreated := 0
	unreliableJudgmentCount := make(map[string]int)   // Track biased judgments per unreliable judge
	unreliableJudgmentBias := make(map[string]string) // Track bias direction per unreliable judge

	// Run simulation for all steps
	fmt.Printf("========================================\n")
	fmt.Printf("      STEP-BY-STEP SIMULATION\n")
	fmt.Printf("========================================\n\n")

	for step := 0; step < pairingInitMatrix.Steps; step++ {
		stepStats := &StepStats{
			AssignedGroups:   make(map[int]string),
			RestingGroups:    []int{},
			JudgeAssignments: make(map[string]string),
		}

		fmt.Printf("=== STEP %d ===\n", step)

		judgesResting := 0
		judgesWorking := 0

		for _, judge := range createdPairingJudges {
			bodyBytes, statusCode := helpers.API_JudgeNextTeam(
				t,
				app,
				judgeTokens[judge.ID],
			)

			groupIdx := judgeToGroupIdx[judge.ID]
			var teamID string

			switch statusCode {
			case http.StatusOK:
				var nextTeam models.Team
				require.NoError(t, json.Unmarshal(bodyBytes, &nextTeam))
				teamID = nextTeam.ID
				require.NotEmpty(t, teamID, "expected team assignment when status is 200")

				refreshedJudge := models.Judge{ID: judge.ID}
				require.NoError(t, refreshedJudge.Get())
				require.Equal(t, step, refreshedJudge.CurrentTeam, "judge current step should match loop step when working")

				currentTeam := nextTeam
				currentBytes, currentStatus := helpers.API_JudgeCurrentTeam(
					t,
					app,
					judgeTokens[judge.ID],
				)
				require.Equal(t, http.StatusOK, currentStatus)

				require.NoError(t, json.Unmarshal(currentBytes, &currentTeam))
				require.Equal(t, teamID, currentTeam.ID, "current team should match recently assigned team")
			case http.StatusAccepted:
				var restResp struct {
					Message string `json:"message"`
				}
				require.NoError(t, json.Unmarshal(bodyBytes, &restResp))
				require.Equal(t, errmsg.JudgeResting.Message, restResp.Message)
				refreshedJudge := models.Judge{ID: judge.ID}
				require.NoError(t, refreshedJudge.Get())
				require.Equal(t, step, refreshedJudge.CurrentTeam, "judge current step should advance even when resting")
				judgesResting++
				continue
			case http.StatusGone:
				var finishedResp struct {
					Message string `json:"message"`
				}
				require.NoError(t, json.Unmarshal(bodyBytes, &finishedResp))
				require.Equal(t, errmsg.JudgingFinished.Message, finishedResp.Message)
				judgesResting++
				continue
			default:
				require.Failf(t, "unexpected status from /judge/next-team", "status=%d body=%s", statusCode, string(bodyBytes))
			}

			stepStats.JudgeAssignments[judge.ID] = teamID
			judgesWorking++

			// Track group assignment
			if existing, ok := stepStats.AssignedGroups[groupIdx]; ok {
				// All judges in same group should get same team
				require.Equal(t, existing, teamID, "judges in group %d should get same team", groupIdx)
			} else {
				stepStats.AssignedGroups[groupIdx] = teamID
			}

			globalTeamAssignments[teamID]++
			globalGroupTeamPairs[fmt.Sprintf("%d-%s", groupIdx, teamID)] = true

			// Create a judgment if this judge has seen a previous team
			if len(judgeTeamHistory[judge.ID]) > 0 {
				previousTeam := judgeTeamHistory[judge.ID][len(judgeTeamHistory[judge.ID])-1]

				// Determine who wins based on judge reliability
				var winningTeam, losingTeam string
				biasMarker := ""

				if bias, isUnreliable := unreliableJudges[judge.ID]; isUnreliable {
					// Unreliable judges have systematic biases
					unreliableJudgmentCount[judge.ID]++
					unreliableJudgmentBias[judge.ID] = bias
					if bias == "prior" {
						// Always vote for previous team
						winningTeam = previousTeam
						losingTeam = teamID
						biasMarker = " [BIASED LEFT: always votes for prior team]"
					} else if bias == "current" {
						// Always vote for current team
						winningTeam = teamID
						losingTeam = previousTeam
						biasMarker = " [BIASED RIGHT: always votes for current team]"
					}
				} else {
					// Reliable judges vote randomly
					if rand.Float32() < 0.5 {
						winningTeam = teamID
						losingTeam = previousTeam
					} else {
						winningTeam = previousTeam
						losingTeam = teamID
					}
				}

				_, statusCode := helpers.API_JudgeCreateJudgment(
					t,
					app,
					winningTeam,
					losingTeam,
					judgeTokens[judge.ID],
				)
				require.Equal(t, http.StatusOK, statusCode)
				totalJudgmentsCreated++
				fmt.Printf("    → Created judgment: %s vs %s (winner: %s)%s\n", previousTeam, teamID, winningTeam, biasMarker)
			}

			// Add to history for next comparison
			judgeTeamHistory[judge.ID] = append(judgeTeamHistory[judge.ID], teamID)
		}

		// Identify resting groups
		for groupIdx := 0; groupIdx < pairingInitMatrix.Groups; groupIdx++ {
			if _, working := stepStats.AssignedGroups[groupIdx]; !working {
				stepStats.RestingGroups = append(stepStats.RestingGroups, groupIdx)
			}
		}

		allStepsStats[step] = stepStats

		// Log step summary
		fmt.Printf("[STEP PARTICIPATION SUMMARY]\n")
		fmt.Printf("  Groups working: %d, Groups resting: %d\n", len(stepStats.AssignedGroups), len(stepStats.RestingGroups))
		fmt.Printf("  Judges working: %d, Judges resting: %d\n", judgesWorking, judgesResting)

		// Verify no collisions: each team should appear at most once per step (per-group check)
		teamsAssignedThisStep := make(map[string][]int) // teamID -> list of group indices
		for groupIdx, teamID := range stepStats.AssignedGroups {
			if teamID != "" {
				teamsAssignedThisStep[teamID] = append(teamsAssignedThisStep[teamID], groupIdx)
			}
		}

		fmt.Printf("\n[COLLISION CHECK - Per-Group Per-Step]\n")
		hasCollisions := false
		for teamID, groupIndices := range teamsAssignedThisStep {
			if len(groupIndices) > 1 {
				fmt.Printf("  ✗ Team %s assigned to %d groups: %v ← COLLISION!\n", teamID, len(groupIndices), groupIndices)
				hasCollisions = true
				totalCollisionsAcrossSteps++
			} else {
				fmt.Printf("  ✓ Team %s assigned to 1 group only\n", teamID)
			}
		}

		if !hasCollisions && len(teamsAssignedThisStep) > 0 {
			fmt.Printf("  ✓ No collisions in step %d\n", step)
		}

		// Show final group->team assignments for this step with judge details
		fmt.Printf("\n[FINAL ASSIGNMENTS THIS STEP]\n")
		for groupIdx := 0; groupIdx < pairingInitMatrix.Groups; groupIdx++ {
			if teamID, ok := stepStats.AssignedGroups[groupIdx]; ok {
				judges := groupToJudges[groupIdx]
				fmt.Printf("  Group %2d (judges: %v) -> %s\n", groupIdx, judges, teamID)
			} else {
				fmt.Printf("  Group %2d -> [REST]\n", groupIdx)
			}
		}
		fmt.Printf("\n")
	}

	// Final verification and summary
	fmt.Printf("========================================\n")
	fmt.Printf("      SIMULATION SUMMARY\n")
	fmt.Printf("========================================\n\n")

	fmt.Printf("Global Statistics:\n")
	fmt.Printf("  Total steps: %d\n", pairingInitMatrix.Steps)
	fmt.Printf("  Total judges: %d\n", len(createdPairingJudges))
	fmt.Printf("  Total groups: %d\n", pairingInitMatrix.Groups)
	fmt.Printf("  Total teams: %d\n", len(createdPairingTeams))
	fmt.Printf("\n")

	// Verify all teams got judged
	teamsJudged := make(map[string]bool)
	for step := 0; step < pairingInitMatrix.Steps; step++ {
		if stats, ok := allStepsStats[step]; ok {
			for _, teamID := range stats.JudgeAssignments {
				teamsJudged[teamID] = true
			}
		}
	}

	fmt.Printf("Teams judged: %d / %d\n", len(teamsJudged), len(createdPairingTeams))
	for _, team := range createdPairingTeams {
		if teamsJudged[team.ID] {
			fmt.Printf("  ✓ %s\n", team.ID)
		} else {
			fmt.Printf("  ✗ %s (NOT JUDGED)\n", team.ID)
		}
	}
	fmt.Printf("\n")

	// Verify all groups got work
	fmt.Printf("Group Activity Summary:\n")
	for groupIdx := 0; groupIdx < pairingInitMatrix.Groups; groupIdx++ {
		assignmentCount := 0
		for step := 0; step < pairingInitMatrix.Steps; step++ {
			if stats, ok := allStepsStats[step]; ok {
				if _, ok := stats.AssignedGroups[groupIdx]; ok {
					assignmentCount++
				}
			}
		}
		fmt.Printf("  Group %2d: %d assignments\n", groupIdx, assignmentCount)
		require.Greater(t, assignmentCount, 0, "group %d should have at least one assignment", groupIdx)
	}
	fmt.Printf("\n")

	// Collision summary
	fmt.Printf("========================================\n")
	fmt.Printf("      COLLISION ANALYSIS\n")
	fmt.Printf("========================================\n\n")
	fmt.Printf("Per-step collisions detected: %d\n", totalCollisionsAcrossSteps)
	if totalCollisionsAcrossSteps == 0 {
		fmt.Printf("✓ Zero collisions - Latin rectangle satisfied\n")
	} else {
		fmt.Printf("✗ Collisions found - matrix generation issue\n")
	}
	fmt.Printf("\n")

	// Judgment summary
	fmt.Printf("========================================\n")
	fmt.Printf("      JUDGMENT SUMMARY\n")
	fmt.Printf("========================================\n\n")
	fmt.Printf("Total judgments created during simulation: %d\n", totalJudgmentsCreated)

	// Fetch actual judgments from database to verify
	cursor, err := db.Judgments.Find(db.Ctx, bson.M{})
	require.NoError(t, err)
	defer cursor.Close(db.Ctx)

	var judgments []models.Judgment
	err = cursor.All(db.Ctx, &judgments)
	require.NoError(t, err)

	fmt.Printf("Judgments persisted in database: %d\n", len(judgments))
	if len(judgments) > 0 {
		fmt.Printf("✓ Pairwise comparisons recorded\n")
	} else {
		fmt.Printf("✗ No judgments found\n")
	}
	fmt.Printf("\n")

	// Unreliable judges summary
	if len(unreliableJudgmentCount) > 0 {
		fmt.Printf("========================================\n")
		fmt.Printf("  UNRELIABLE JUDGES BEHAVIOR SUMMARY\n")
		fmt.Printf("========================================\n\n")
		for judgeID, count := range unreliableJudgmentCount {
			bias := unreliableJudgmentBias[judgeID]
			if bias == "prior" {
				fmt.Printf("  %s: Made %d biased judgments (LEFT bias: always voted for prior team)\n", judgeID, count)
			} else if bias == "current" {
				fmt.Printf("  %s: Made %d biased judgments (RIGHT bias: always voted for current team)\n", judgeID, count)
			}
		}
		fmt.Printf("\nExpected Gavel Algorithm Result:\n")
		fmt.Printf("  • Both unreliable judges should show LOW α/β ratio (high beta value)\n")
		fmt.Printf("  • LEFT bias judge: systematically voted for prior, contradicting skill rankings\n")
		fmt.Printf("  • RIGHT bias judge: systematically voted for current, contradicting skill rankings\n")
		fmt.Printf("  • Algorithm should infer: α ↓ (reduced reliability), β ↑ (increased noise) for both\n")
		fmt.Printf("  • Control judge (if present) should maintain HIGH α/β ratio\n\n")
	}

	fmt.Printf("========================================\n")
	if len(teamsJudged) == len(createdPairingTeams) && totalCollisionsAcrossSteps == 0 && len(judgments) > 0 {
		fmt.Printf("✓ FULL SIMULATION COMPLETED SUCCESSFULLY\n")
		fmt.Printf("  All teams judged, zero collisions, pairwise judgments created\n")
	} else {
		fmt.Printf("✗ SIMULATION COMPLETED WITH ISSUES\n")
		if len(teamsJudged) < len(createdPairingTeams) {
			fmt.Printf("  Teams not judged: %d\n", len(createdPairingTeams)-len(teamsJudged))
		}
		if totalCollisionsAcrossSteps > 0 {
			fmt.Printf("  Per-step collisions: %d\n", totalCollisionsAcrossSteps)
		}
		if len(judgments) == 0 {
			fmt.Printf("  No judgments created\n")
		}
	}
	fmt.Printf("========================================\n")
}

// TestJudgingPairsCrowdBTScoring runs the Crowd Bradley-Terry algorithm on the created judgments
func TestJudgingPairsCrowdBTScoring(t *testing.T) {
	fmt.Printf("\n========================================\n")
	fmt.Printf("    CROWD BT SCORING ALGORITHM\n")
	fmt.Printf("========================================\n\n")

	// Fetch all judgments from database
	cursor, err := db.Judgments.Find(db.Ctx, bson.M{})
	require.NoError(t, err)
	defer cursor.Close(db.Ctx)

	var dbJudgments []models.Judgment
	err = cursor.All(db.Ctx, &dbJudgments)
	require.NoError(t, err)

	fmt.Printf("Loaded %d judgments from database\n", len(dbJudgments))
	if len(dbJudgments) == 0 {
		fmt.Printf("⚠ No judgments found - skipping scoring\n")
		return
	}

	// Convert to utils.JudgmentWithJudge format
	judgments := make([]utils.JudgmentWithJudge, len(dbJudgments))
	for i, j := range dbJudgments {
		judgments[i] = utils.JudgmentWithJudge{
			WinningTeamID: j.WinningTeamID,
			LosingTeamID:  j.LosingTeamID,
			JudgeID:       j.JudgeID,
		}
	}

	fmt.Printf("Running Crowd BT algorithm...\n")

	// Time the algorithm
	startTime := time.Now()

	// Run the algorithm
	scorer := utils.NewCrowdBTScorer()
	scores := scorer.Score(judgments)
	ranking := scorer.RankTeams()
	judgeReliability := scorer.GetJudgeReliabilityAll()
	teamUncertainty := scorer.GetTeamUncertainty()

	elapsedTime := time.Since(startTime)

	fmt.Printf("✓ Algorithm completed in %v\n\n", elapsedTime)

	// Print team rankings
	fmt.Printf("========================================\n")
	fmt.Printf("        TEAM RANKINGS (by Mu)\n")
	fmt.Printf("========================================\n\n")
	fmt.Printf("Rank | Team       | μ (Skill)  | σ² (Unc.)    | Confidence | Notes\n")
	fmt.Printf("-----|------------|------------|--------------|------------|---------------------\n")

	for i, teamID := range ranking {
		mu := scores[teamID]
		sigmaSq := teamUncertainty[teamID]
		confidence := 1.0 / (1.0 + sigmaSq)
		note := ""

		if confidence > 0.9 {
			note = "Very high confidence"
		} else if confidence > 0.8 {
			note = "High confidence"
		} else if confidence > 0.6 {
			note = "Medium confidence"
		} else if confidence > 0.4 {
			note = "Low confidence"
		} else {
			note = "Very low confidence"
		}

		fmt.Printf("%4d | %-10s | %10.4f | %12.4f | %10.4f | %s\n",
			i+1, teamID, mu, sigmaSq, confidence, note)
	}

	fmt.Printf("\n")
	fmt.Printf("━━━ EXPLANATION OF SIGMA-SQUARED (σ²) ━━━\n")
	fmt.Printf("σ² = Variance/Uncertainty in the team's skill estimate\n")
	fmt.Printf("  • σ² ≈ 0.1-0.5: HIGH CONFIDENCE (many judgments, strong consensus)\n")
	fmt.Printf("  • σ² ≈ 0.5-1.5: MEDIUM CONFIDENCE (moderate data, some variation)\n")
	fmt.Printf("  • σ² > 1.5:     LOW CONFIDENCE (few judgments or lots of conflict)\n\n")

	fmt.Printf("━━━ EXPLANATION OF CONFIDENCE ━━━\n")
	fmt.Printf("Confidence = 1 / (1 + σ²)  [ranges from 0 to 1]\n")
	fmt.Printf("  • 0.0 = completely uncertain\n")
	fmt.Printf("  • 0.5 = 50%% confident\n")
	fmt.Printf("  • 1.0 = perfect confidence (σ²=0, infinite judgments)\n")
	fmt.Printf("Example: σ²=1.0 → confidence=0.5 (50%% sure of ranking)\n\n")

	// Print judge reliability
	fmt.Printf("========================================\n")
	fmt.Printf("       JUDGE RELIABILITY (α, β)\n")
	fmt.Printf("========================================\n\n")

	// Count judgments per judge
	judgeCount := make(map[string]int)
	for _, j := range dbJudgments {
		judgeCount[j.JudgeID]++
	}

	// Sort judges by ID for consistent output
	var judgeIDs []string
	for judgeID := range judgeReliability {
		judgeIDs = append(judgeIDs, judgeID)
	}
	sort.Strings(judgeIDs)

	fmt.Printf("Judge      | α (Alpha)  | β (Beta)   | α/β Ratio  | #Judgments | Class\n")
	fmt.Printf("-----------|------------|------------|------------|------------|----------------\n")

	for _, judgeID := range judgeIDs {
		params := judgeReliability[judgeID]
		alpha := params[0]
		beta := params[1]
		ratio := 1.0
		if beta > 0 {
			ratio = alpha / beta
		}
		count := judgeCount[judgeID]

		// Interpret reliability
		reliability := "unknown"
		if ratio > 2.0 {
			reliability = "highly_reliable"
		} else if ratio > 1.0 {
			reliability = "reliable"
		} else if ratio < 0.5 {
			reliability = "noisy"
		} else {
			reliability = "neutral"
		}

		fmt.Printf("%-10s | %10.4f | %10.4f | %10.4f | %10d | %s\n",
			judgeID, alpha, beta, ratio, count, reliability)
	}

	fmt.Printf("\n")
	fmt.Printf("━━━ WHY JUDGES APPEAR RELIABLE DESPITE RANDOM JUDGMENTS ━━━\n")
	fmt.Printf("Initial Priors: α=10.0, β=1.0\n")
	fmt.Printf("  These priors assume judges are generally reliable.\n\n")

	fmt.Printf("What Happens with Random Judgments:\n")
	fmt.Printf("  1. Random votes create roughly balanced wins/losses per team\n")
	fmt.Printf("  2. Team skill estimates converge to ~equal values (all ≈0)\n")
	fmt.Printf("  3. When comparing ~equal-skilled teams, RANDOM votes look \"expected\"\n")
	fmt.Printf("  4. The algorithm interprets this as: \"judge voted for likely winner\"\n")
	fmt.Printf("  5. Result: α stays high, β stays low → high α/β ratio\n\n")

	fmt.Printf("What WOULD show a judge is unreliable (high β):\n")
	fmt.Printf("  • Consistently voting for clearly-weaker teams\n")
	fmt.Printf("  • Pattern of voting opposite to established skill differences\n")
	fmt.Printf("  • Systematic bias that contradicts the learned rankings\n\n")

	// Summary statistics
	fmt.Printf("========================================\n")
	fmt.Printf("         ALGORITHM SUMMARY\n")
	fmt.Printf("========================================\n\n")

	fmt.Printf("Execution time:           %v\n", elapsedTime)
	fmt.Printf("Teams scored:             %d\n", len(scores))
	fmt.Printf("Judges evaluated:         %d\n", len(judgeReliability))
	fmt.Printf("Total judgments:          %d\n", len(judgments))
	fmt.Printf("Avg judgments per judge:  %.1f\n", float64(len(judgments))/float64(len(judgeReliability)))

	// Calculate average confidence
	var totalConfidence float64
	for _, sigmaSq := range teamUncertainty {
		confidence := 1.0 / (1.0 + sigmaSq)
		totalConfidence += confidence
	}
	avgConfidence := totalConfidence / float64(len(teamUncertainty))
	fmt.Printf("Average ranking confidence: %.4f\n", avgConfidence)

	// Find most and least reliable judges
	var mostReliable, leastReliable string
	var mostRatio, leastRatio float64 = -1, 1000000

	for _, judgeID := range judgeIDs {
		params := judgeReliability[judgeID]
		ratio := 1.0
		if params[1] > 0 {
			ratio = params[0] / params[1]
		}

		if ratio > mostRatio {
			mostRatio = ratio
			mostReliable = judgeID
		}
		if ratio < leastRatio {
			leastRatio = ratio
			leastReliable = judgeID
		}
	}

	fmt.Printf("\nMost reliable judge:      %s (α/β = %.4f)\n", mostReliable, mostRatio)
	fmt.Printf("Least reliable judge:     %s (α/β = %.4f)\n", leastReliable, leastRatio)

	// Find highest and lowest ranked teams
	if len(ranking) > 0 {
		bestTeam := ranking[0]
		worstTeam := ranking[len(ranking)-1]
		bestSigmaSq := teamUncertainty[bestTeam]
		worstSigmaSq := teamUncertainty[worstTeam]
		bestConfidence := 1.0 / (1.0 + bestSigmaSq)
		worstConfidence := 1.0 / (1.0 + worstSigmaSq)

		fmt.Printf("\nHighest ranked team:      %s (μ=%.4f, σ²=%.4f, conf=%.4f)\n",
			bestTeam, scores[bestTeam], bestSigmaSq, bestConfidence)
		fmt.Printf("Lowest ranked team:       %s (μ=%.4f, σ²=%.4f, conf=%.4f)\n",
			worstTeam, scores[worstTeam], worstSigmaSq, worstConfidence)
	}

	fmt.Printf("========================================\n")
	fmt.Printf("✓ CROWD BT SCORING COMPLETE\n")
	fmt.Printf("========================================\n\n")
}

// TestJudgingPairsComputeRankingsEndpoint tests the compute-rankings endpoint
func TestJudgingPairsComputeRankingsEndpoint(t *testing.T) {
	require.NotNil(t, app, "app should be initialized")

	fmt.Printf("\n========================================\n")
	fmt.Printf("    COMPUTE RANKINGS ENDPOINT TEST\n")
	fmt.Printf("========================================\n\n")

	// Call the compute rankings endpoint
	bodyBytes, statusCode := helpers.API_SuperUsersJudgingComputeRankings(
		t,
		app,
		pairingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode, "compute-rankings endpoint should return 200")

	// Parse response
	var response struct {
		RankedTeams      []string                      `json:"rankedTeams"`
		TeamScores       map[string]float64            `json:"teamScores"`
		TeamUncertainty  map[string]float64            `json:"teamUncertainty"`
		JudgeReliability map[string]map[string]float64 `json:"judgeReliability"`
		JudgmentCount    int                           `json:"judgmentCount"`
	}

	err := json.Unmarshal(bodyBytes, &response)
	require.NoError(t, err, "response should be valid JSON")

	fmt.Printf("Response parsed successfully\n")
	fmt.Printf("Judgments processed: %d\n", response.JudgmentCount)
	fmt.Printf("Teams ranked: %d\n", len(response.RankedTeams))
	fmt.Printf("Judges evaluated: %d\n", len(response.JudgeReliability))

	// Validate response structure
	require.Greater(t, len(response.RankedTeams), 0, "should have ranked teams")
	require.Equal(t, len(response.RankedTeams), len(response.TeamScores), "ranked teams should match team scores")
	require.Equal(t, len(response.RankedTeams), len(response.TeamUncertainty), "ranked teams should match team uncertainty")
	require.Greater(t, response.JudgmentCount, 0, "should have processed judgments")

	// Verify team scores are in descending order
	for i := 0; i < len(response.RankedTeams)-1; i++ {
		team1 := response.RankedTeams[i]
		team2 := response.RankedTeams[i+1]
		score1 := response.TeamScores[team1]
		score2 := response.TeamScores[team2]
		require.GreaterOrEqual(t, score1, score2, "teams should be sorted by score descending")
	}

	// Verify all scores have uncertainty values
	for teamID := range response.TeamScores {
		_, hasUncertainty := response.TeamUncertainty[teamID]
		require.True(t, hasUncertainty, "all teams should have uncertainty values")
	}

	// Verify judge reliability parameters are positive
	for judgeID, params := range response.JudgeReliability {
		alpha, hasAlpha := params["alpha"]
		beta, hasBeta := params["beta"]
		require.True(t, hasAlpha, "judge %s should have alpha parameter", judgeID)
		require.True(t, hasBeta, "judge %s should have beta parameter", judgeID)
		require.Greater(t, alpha, 0.0, "judge %s alpha should be positive", judgeID)
		require.Greater(t, beta, 0.0, "judge %s beta should be positive", judgeID)
	}

	// Print full rankings
	fmt.Printf("\n========================================\n")
	fmt.Printf("         FULL RANKINGS\n")
	fmt.Printf("========================================\n\n")
	fmt.Printf("Rank | Team ID              | Score (μ) | Uncertainty | Confidence\n")
	fmt.Printf("-----|----------------------|-----------|-------------|------------\n")
	for i, teamID := range response.RankedTeams {
		score := response.TeamScores[teamID]
		uncertainty := response.TeamUncertainty[teamID]
		confidence := 1.0 / (1.0 + uncertainty)
		fmt.Printf("%4d | %-20s | %9.4f | %11.4f | %.4f\n", i+1, teamID, score, uncertainty, confidence)
	}

	// Print finalists (top 3)
	fmt.Printf("\n========================================\n")
	fmt.Printf("         🏆 FINALISTS 🏆\n")
	fmt.Printf("========================================\n\n")
	numFinalists := 3
	if len(response.RankedTeams) < 3 {
		numFinalists = len(response.RankedTeams)
	}
	for i := 0; i < numFinalists; i++ {
		teamID := response.RankedTeams[i]
		score := response.TeamScores[teamID]
		uncertainty := response.TeamUncertainty[teamID]
		confidence := 1.0 / (1.0 + uncertainty)
		fmt.Printf("%d. %s (μ=%.4f, σ²=%.4f, confidence=%.4f)\n", i+1, teamID, score, uncertainty, confidence)
	}

	// Print judge confidence analysis
	fmt.Printf("\n========================================\n")
	fmt.Printf("    JUDGE RELIABILITY ANALYSIS\n")
	fmt.Printf("========================================\n\n")
	fmt.Printf("Judge ID             | Alpha      | Beta       | Alpha/Beta | Reliability\n")
	fmt.Printf("---------------------|------------|------------|------------|------------------\n")

	var judgeIDs []string
	for judgeID := range response.JudgeReliability {
		judgeIDs = append(judgeIDs, judgeID)
	}
	sort.Strings(judgeIDs)

	for _, judgeID := range judgeIDs {
		params := response.JudgeReliability[judgeID]
		alpha := params["alpha"]
		beta := params["beta"]
		ratio := alpha / beta
		reliability := "unknown"
		if ratio > 2.0 {
			reliability = "highly_reliable"
		} else if ratio > 1.0 {
			reliability = "reliable"
		} else if ratio < 0.5 {
			reliability = "noisy"
		} else {
			reliability = "neutral"
		}
		fmt.Printf("%-20s | %10.4f | %10.4f | %10.4f | %s\n", judgeID, alpha, beta, ratio, reliability)
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("✓ COMPUTE RANKINGS ENDPOINT TEST PASSED\n")
	fmt.Printf("========================================\n\n")
}

// TestJudgingPairsParticipantVoting tests the participant voting system
func TestJudgingPairsParticipantVoting(t *testing.T) {
	require.NotNil(t, app, "app should be initialized")

	fmt.Printf("\n========================================\n")
	fmt.Printf("    PARTICIPANT VOTING TEST\n")
	fmt.Printf("========================================\n\n")

	// Enable judging stage (stage 6) to compute rankings
	_, statusCode := helpers.API_SuperUsersFlagStagesExecute(
		t,
		app,
		"6",
		pairingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode, "should be able to execute judging stage")

	// Get finalists from compute rankings endpoint
	rankingBody, statusCode := helpers.API_SuperUsersJudgingComputeRankings(
		t,
		app,
		pairingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)

	// Now switch to voting stage (stage 8: Public Voting)
	_, statusCode = helpers.API_SuperUsersFlagStagesExecute(
		t,
		app,
		"8",
		pairingTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode, "should be able to execute voting stage")

	var rankingResp struct {
		RankedTeams []string `json:"rankedTeams"`
	}
	errUnmarshal := json.Unmarshal(rankingBody, &rankingResp)
	require.NoError(t, errUnmarshal)
	require.GreaterOrEqual(t, len(rankingResp.RankedTeams), 3, "should have at least 3 teams to rank")

	// Get top 3 finalists
	finalist1 := rankingResp.RankedTeams[0]
	finalist2 := rankingResp.RankedTeams[1]
	finalist3 := rankingResp.RankedTeams[2]

	fmt.Printf("Finalists:\n")
	fmt.Printf("  1. %s\n", finalist1)
	fmt.Printf("  2. %s\n", finalist2)
	fmt.Printf("  3. %s\n", finalist3)
	fmt.Printf("\n")

	// Verify finalists are saved to settings
	f1Setting := &models.Setting{Name: models.SettingFinalist1}
	f2Setting := &models.Setting{Name: models.SettingFinalist2}
	f3Setting := &models.Setting{Name: models.SettingFinalist3}

	serr := f1Setting.Get()
	require.Equal(t, errmsg.EmptyStatusError, serr)
	require.Equal(t, finalist1, f1Setting.Value.(string))

	serr = f2Setting.Get()
	require.Equal(t, errmsg.EmptyStatusError, serr)
	require.Equal(t, finalist2, f2Setting.Value.(string))

	serr = f3Setting.Get()
	require.Equal(t, errmsg.EmptyStatusError, serr)
	require.Equal(t, finalist3, f3Setting.Value.(string))

	fmt.Printf("✓ Finalists saved to settings\n\n")

	// Cast votes for all participants
	fmt.Printf("========================================\n")
	fmt.Printf("      VOTING PHASE\n")
	fmt.Printf("========================================\n\n")

	voteCastLog := make(map[string]string) // accountID -> finalist team ID

	finalists := []string{finalist1, finalist2, finalist3}

	// First, register all accounts with passwords
	const participantPassword = "testpassword123"
	for _, account := range createdPairingAccounts {
		// Register account with password
		_, statusCode := helpers.API_AccountsAuthRegister(
			t,
			app,
			account.Email,
			participantPassword,
		)
		require.Equal(t, http.StatusOK, statusCode, "should be able to register account %s", account.Email)
	}

	for _, account := range createdPairingAccounts {
		// Login as participant
		loginBody, statusCode := helpers.API_AccountsAuthLogin(
			t,
			app,
			account.Email,
			participantPassword,
		)
		require.Equal(t, http.StatusOK, statusCode, "should be able to login")

		var loginResp struct {
			Token string `json:"token"`
		}
		err := json.Unmarshal(loginBody, &loginResp)
		require.NoError(t, err)
		participantToken := loginResp.Token

		// Get voting status
		statusBody, statusCode := helpers.API_AccountsVotingStatus(
			t,
			app,
			participantToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var statusResp struct {
			VotingOpen bool          `json:"votingOpen"`
			HasVoted   bool          `json:"hasVoted"`
			Finalists  []models.Team `json:"finalists"`
		}
		err = json.Unmarshal(statusBody, &statusResp)
		require.NoError(t, err)
		require.True(t, statusResp.VotingOpen, "voting should be open")
		require.False(t, statusResp.HasVoted, "participant should not have voted yet")
		require.Equal(t, 3, len(statusResp.Finalists), "should have 3 finalists")

		// Get finalists for this participant
		finalistsBody, statusCode := helpers.API_AccountsVotingFinalists(
			t,
			app,
			participantToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var finalistsResp struct {
			Finalists []models.Team `json:"finalists"`
		}
		err = json.Unmarshal(finalistsBody, &finalistsResp)
		require.NoError(t, err)
		require.Equal(t, 3, len(finalistsResp.Finalists), "should receive 3 finalists")

		// Randomly select a finalist to vote for
		selectedFinalist := finalists[rand.Intn(len(finalists))]

		// Cast vote
		voteBody, statusCode := helpers.API_AccountsVotingCastVote(
			t,
			app,
			selectedFinalist,
			participantToken,
		)
		require.Equal(t, http.StatusOK, statusCode, "vote should be recorded for %s", account.ID)

		var voteResp struct {
			Message string `json:"message"`
		}
		err = json.Unmarshal(voteBody, &voteResp)
		require.NoError(t, err)

		voteCastLog[account.ID] = selectedFinalist

		fmt.Printf("  %s voted for %s\n", account.Email, selectedFinalist)

		// Verify hasVoted is now true
		statusBody2, statusCode := helpers.API_AccountsVotingStatus(
			t,
			app,
			participantToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var statusResp2 struct {
			HasVoted bool `json:"hasVoted"`
		}
		err = json.Unmarshal(statusBody2, &statusResp2)
		require.NoError(t, err)
		require.True(t, statusResp2.HasVoted, "participant should have voted after casting vote")

		// Try to vote again - should fail
		_, statusCode = helpers.API_AccountsVotingCastVote(
			t,
			app,
			selectedFinalist,
			participantToken,
		)
		require.Equal(t, http.StatusConflict, statusCode, "double vote should be rejected")
	}

	fmt.Printf("\n✓ %d participants voted successfully\n\n", len(voteCastLog))

	// Print vote log
	fmt.Printf("========================================\n")
	fmt.Printf("         VOTES CAST LOG\n")
	fmt.Printf("========================================\n\n")
	for accountID, teamID := range voteCastLog {
		fmt.Printf("  %s → %s\n", accountID, teamID)
	}
	fmt.Printf("\n")

	// Verify votes in database
	fmt.Printf("========================================\n")
	fmt.Printf("      VOTE VERIFICATION\n")
	fmt.Printf("========================================\n\n")

	cursor, err := db.Votes.Find(db.Ctx, bson.M{})
	require.NoError(t, err)
	defer cursor.Close(db.Ctx)

	var dbVotes []models.Vote
	err = cursor.All(db.Ctx, &dbVotes)
	require.NoError(t, err)

	fmt.Printf("Votes in database: %d\n", len(dbVotes))
	fmt.Printf("Votes logged: %d\n", len(voteCastLog))
	require.Equal(t, len(voteCastLog), len(dbVotes), "vote count should match")

	// Count votes per finalist
	voteCount := make(map[string]int)
	for _, vote := range dbVotes {
		voteCount[vote.Choice]++
	}

	fmt.Printf("\nVote Distribution:\n")
	for _, finalist := range finalists {
		count := voteCount[finalist]
		fmt.Printf("  %s: %d votes\n", finalist, count)
	}

	// Verify all votes are for valid finalists
	for _, vote := range dbVotes {
		isValid := false
		for _, finalist := range finalists {
			if vote.Choice == finalist {
				isValid = true
				break
			}
		}
		require.True(t, isValid, "vote for %s should be for a valid finalist", vote.Choice)
	}

	fmt.Printf("\n✓ All votes verified\n")
	fmt.Printf("========================================\n")
	fmt.Printf("✓ PARTICIPANT VOTING TEST PASSED\n")
	fmt.Printf("========================================\n\n")
}

// TestJudgingPairsCleanup cleans up resources
func TestJudgingPairsCleanup(t *testing.T) {
	_, err := db.Judges.DeleteMany(db.Ctx, bson.M{})
	require.NoError(t, err)

	_, err = db.Accounts.DeleteMany(db.Ctx, bson.M{})
	require.NoError(t, err)

	_, err = db.Teams.DeleteMany(db.Ctx, bson.M{})
	require.NoError(t, err)

	_, err = db.Judgments.DeleteMany(db.Ctx, bson.M{})
	require.NoError(t, err)

	_, err = db.Votes.DeleteMany(db.Ctx, bson.M{})
	require.NoError(t, err)

	fmt.Printf("Cleanup complete\n")
}
