package superusers

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/test/helpers"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"

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
	for solo := 0; solo < numSoloJudges; solo++ {
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
	for pair := 0; pair < numPairJudges; pair++ {
		groupID := numSoloJudges + pair
		pairAttr := fmt.Sprintf("group_%02d", groupID)
		for member := 0; member < 2; member++ {
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
	for trio := 0; trio < numTrioJudges; trio++ {
		groupID := numSoloJudges + numPairJudges + trio
		pairAttr := fmt.Sprintf("group_%02d", groupID)
		for member := 0; member < 3; member++ {
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
	for i := 0; i < numPairingTeams; i++ {
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

// TestJudgingPairsPairingValidation validates the judge pairing groups
func TestJudgingPairsPairingValidation(t *testing.T) {
	require.NotEmpty(t, pairingTestSuperUserToken, "superuser token should be initialized")
	require.Equal(t, totalPairingJudgeGroups, pairingInitMatrix.Groups, "should have %d pair groups", totalPairingJudgeGroups)

	fmt.Printf("\n========================================\n")
	fmt.Printf("      JUDGE PAIR GROUPS\n")
	fmt.Printf("========================================\n")

	// Fetch pair groups from settings
	pairGroupsSetting := &models.Setting{Name: models.SettingJudgePairGroups}
	errStatus := pairGroupsSetting.Get()
	require.Equal(t, errmsg.EmptyStatusError, errStatus)

	var pairGroups []struct {
		GroupID   int      `json:"groupId"`
		JudgeIDs  []string `json:"judgeIds"`
		PairAttr  string   `json:"pairAttr"`
		NumJudges int      `json:"numJudges"`
	}
	require.NoError(t, json.Unmarshal([]byte(pairGroupsSetting.Value.(string)), &pairGroups))

	// Build expected group sizes dynamically
	var expectedGroupSizes []int
	for i := 0; i < numSoloJudges; i++ {
		expectedGroupSizes = append(expectedGroupSizes, 1)
	}
	for i := 0; i < numPairJudges; i++ {
		expectedGroupSizes = append(expectedGroupSizes, 2)
	}
	for i := 0; i < numTrioJudges; i++ {
		expectedGroupSizes = append(expectedGroupSizes, 3)
	}

	for i, group := range pairGroups {
		fmt.Printf("Group %d (pair='%s'):\n", group.GroupID, group.PairAttr)
		fmt.Printf("  Judges: %v\n", group.JudgeIDs)
		fmt.Printf("  Count: %d\n", group.NumJudges)

		if i < len(expectedGroupSizes) {
			require.Equal(t, expectedGroupSizes[i], group.NumJudges, "group %d should have %d judges", i, expectedGroupSizes[i])
		}
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
			require.Equal(t, http.StatusOK, statusCode)

			var nextTeamResp struct {
				TeamID  string `json:"teamID"`
				Message string `json:"message"`
			}
			require.NoError(t, json.Unmarshal(bodyBytes, &nextTeamResp))

			groupIdx := judgeToGroupIdx[judge.ID]

			if nextTeamResp.Message == "judging finished" {
				judgesResting++
				continue
			}

			teamID := nextTeamResp.TeamID

			// Empty teamID means rest step (blank cell in matrix)
			if teamID == "" {
				judgesResting++
				continue
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

				// Randomly decide who wins
				var winningTeam, losingTeam string
				if rand.Float32() < 0.5 {
					winningTeam = teamID
					losingTeam = previousTeam
				} else {
					winningTeam = previousTeam
					losingTeam = teamID
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
				fmt.Printf("    → Created judgment: %s vs %s (winner: %s)\n", previousTeam, teamID, winningTeam)
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

	fmt.Printf("Cleanup complete\n")
}
