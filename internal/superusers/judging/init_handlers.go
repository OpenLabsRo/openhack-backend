package judging

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"math/rand"
	"sort"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// getCoprimeMultipliers returns a list of numbers coprime to n
func getCoprimeMultipliers(n int) []int {
	coprimes := []int{}
	for i := 1; i < n; i++ {
		if gcd(i, n) == 1 {
			coprimes = append(coprimes, i)
		}
	}
	return coprimes
}

// gcd computes the greatest common divisor using Euclidean algorithm
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// JudgeInitMatrix represents the step-by-group assignment matrix
type JudgeInitMatrix struct {
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
	Matrix        [][]string `json:"matrix"` // matrix[step][groupIdx] = teamID or ""
}

// judgeInitHandler initializes judging settings with judge pairing system.
// @Summary Initialize judging configuration with judge pairing
// @Description Groups judges by pair attribute and creates Latin rectangle assignment
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judging/init [post]
func judgeInitHandler(c fiber.Ctx) error {
	superuser := models.SuperUser{}
	utils.GetLocals(c, "superuser", &superuser)

	// Fetch all teams and judges
	cursor, err := db.Teams.Find(db.Ctx, bson.M{})
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	defer cursor.Close(db.Ctx)

	var teams []models.Team
	if err = cursor.All(db.Ctx, &teams); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	cursorJudges, err := db.Judges.Find(db.Ctx, bson.M{})
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	defer cursorJudges.Close(db.Ctx)

	var judges []models.Judge
	if err = cursorJudges.All(db.Ctx, &judges); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	numTeams := len(teams)
	numJudges := len(judges)

	if numTeams == 0 || numJudges == 0 {
		return utils.StatusError(c, errmsg.InternalServerError(
			&errorMessage{message: "no teams or judges found"},
		))
	}

	// === PHASE 1: GROUP JUDGES BY PAIR ATTRIBUTE ===
	pairGroups := make(map[string][]string)
	judgeIDToGroupIdx := make(map[string]int)

	for _, judge := range judges {
		pairAttr := judge.Pair
		if pairAttr == "" {
			pairAttr = judge.ID
		}
		pairGroups[pairAttr] = append(pairGroups[pairAttr], judge.ID)
	}

	var pairAttrKeys []string
	for attr := range pairGroups {
		pairAttrKeys = append(pairAttrKeys, attr)
	}
	sort.Strings(pairAttrKeys)

	var judgePairGroups []JudgePairGroup
	for groupID, attr := range pairAttrKeys {
		group := JudgePairGroup{
			GroupID:   groupID,
			JudgeIDs:  pairGroups[attr],
			PairAttr:  attr,
			NumJudges: len(pairGroups[attr]),
		}
		judgePairGroups = append(judgePairGroups, group)

		for _, judgeID := range group.JudgeIDs {
			judgeIDToGroupIdx[judgeID] = groupID
		}
	}

	numPairGroups := len(judgePairGroups)

	// === PHASE 2: CREATE SHUFFLED TEAM ORDERS ===
	teamIDs := make([]string, numTeams)
	for i, team := range teams {
		teamIDs[i] = team.ID
	}

	teamOrderA := make([]string, numTeams)
	copy(teamOrderA, teamIDs)
	rand.Shuffle(len(teamOrderA), func(i, j int) {
		teamOrderA[i], teamOrderA[j] = teamOrderA[j], teamOrderA[i]
	})

	teamOrderB := make([]string, numTeams)
	copy(teamOrderB, teamIDs)
	rand.Shuffle(len(teamOrderB), func(i, j int) {
		teamOrderB[i], teamOrderB[j] = teamOrderB[j], teamOrderB[i]
	})

	// Ensure different
	for {
		different := false
		for i := range teamOrderA {
			if teamOrderA[i] != teamOrderB[i] {
				different = true
				break
			}
		}
		if different {
			break
		}
		rand.Shuffle(len(teamOrderB), func(i, j int) {
			teamOrderB[i], teamOrderB[j] = teamOrderB[j], teamOrderB[i]
		})
	}

	baseOrders := [][]string{teamOrderA, teamOrderB}
	baseOrdersJSON, err := json.Marshal(baseOrders)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	teamOrderSetting := models.Setting{
		Name:  models.SettingTeamOrder,
		Value: string(baseOrdersJSON),
	}
	if serr := teamOrderSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	events.Em.JudgeInitTeamOrderSet(superuser.Username, teamOrderA)

	// === PHASE 3: CALCULATE STEPS ===
	numSteps := numPairGroups
	if numTeams > numPairGroups {
		numSteps = numTeams
	}

	// === PHASE 4: BUILD LATIN RECTANGLE WITH EVEN BLANK DISTRIBUTION ===
	// Strategy: For each group, randomly shuffle step assignments with blanks evenly distributed
	// Then assign teams per-step ensuring no collisions (same team in same step for different groups)

	matrix := make([][]string, numSteps)
	for i := range numSteps {
		matrix[i] = make([]string, numPairGroups)
	}

	pairGroupTeamsSeen := make(map[int]map[string]bool)
	for i := range numPairGroups {
		pairGroupTeamsSeen[i] = make(map[string]bool)
	}

	// For each group, determine which steps it will be active (rest of steps are blank)
	// Distribute blanks evenly
	activeStepsPerGroup := make([][]int, numPairGroups)
	blanksPerGroup := numSteps - numTeams // How many blanks each group should have
	if blanksPerGroup < 0 {
		blanksPerGroup = 0
	}

	for groupIdx := 0; groupIdx < numPairGroups; groupIdx++ {
		// Create list of all steps
		allSteps := make([]int, numSteps)
		for i := 0; i < numSteps; i++ {
			allSteps[i] = i
		}

		// Shuffle and take the first (numSteps - blanksPerGroup) as active
		rand.Shuffle(len(allSteps), func(i, j int) {
			allSteps[i], allSteps[j] = allSteps[j], allSteps[i]
		})

		activeCount := numSteps - blanksPerGroup
		activeStepsPerGroup[groupIdx] = allSteps[:activeCount]
		sort.Ints(activeStepsPerGroup[groupIdx]) // Keep in order for consistency
	}

	// Now assign teams to active steps, ensuring no collisions
	teamsAssignedPerStep := make(map[int]map[string]bool)
	for i := 0; i < numSteps; i++ {
		teamsAssignedPerStep[i] = make(map[string]bool)
	}

	coprimes := getCoprimeMultipliers(numTeams)
	if len(coprimes) == 0 {
		return utils.StatusError(c, errmsg.InternalServerError(
			&errorMessage{message: "no coprime multipliers found"},
		))
	}

	// Assign offset and multiplier to each group for team selection
	pairOffsets := make([]int, numPairGroups)
	pairMultipliers := make([]int, numPairGroups)

	for i := range numPairGroups {
		pairOffsets[i] = rand.Intn(numTeams)
		pairMultipliers[i] = coprimes[i%len(coprimes)]
	}

	rand.Shuffle(len(pairMultipliers), func(i, j int) {
		pairMultipliers[i], pairMultipliers[j] = pairMultipliers[j], pairMultipliers[i]
	})

	// Assign teams step-by-step to prevent collisions
	// For each step, go through groups that are active in that step
	for step := 0; step < numSteps; step++ {
		for groupIdx := 0; groupIdx < numPairGroups; groupIdx++ {
			// Check if this group is active in this step
			isActiveInStep := false
			stepInGroup := -1
			for idx, activeStep := range activeStepsPerGroup[groupIdx] {
				if activeStep == step {
					isActiveInStep = true
					stepInGroup = idx
					break
				}
			}

			if !isActiveInStep {
				// This group is resting in this step
				matrix[step][groupIdx] = ""
				continue
			}

			// This group is active in this step, find a team
			offset := pairOffsets[groupIdx]
			multiplier := pairMultipliers[groupIdx]

			// Calculate team index using offset + step*multiplier
			teamIndex := (offset + stepInGroup*multiplier) % numTeams

			// Try to find a team that:
			// 1. Hasn't been assigned to this group yet
			// 2. Isn't already assigned to ANY other group in this step
			found := false
			attempts := 0
			maxAttempts := numTeams

			for attempts < maxAttempts {
				candidateTeamID := teamOrderA[teamIndex]

				if !pairGroupTeamsSeen[groupIdx][candidateTeamID] && !teamsAssignedPerStep[step][candidateTeamID] {
					matrix[step][groupIdx] = candidateTeamID
					pairGroupTeamsSeen[groupIdx][candidateTeamID] = true
					teamsAssignedPerStep[step][candidateTeamID] = true
					found = true
					break
				}

				teamIndex = (teamIndex + 1) % numTeams
				attempts++
			}

			if !found {
				// No valid team found for this group in this step, leave blank
				matrix[step][groupIdx] = ""
			}
		}
	}

	// === PHASE 5: CALCULATE METRICS ===
	assignments := 0
	blanks := 0
	collisions := 0

	for step := 0; step < numSteps; step++ {
		teamCounts := make(map[string]int)
		for groupIdx := 0; groupIdx < numPairGroups; groupIdx++ {
			teamID := matrix[step][groupIdx]
			if teamID == "" {
				blanks++
			} else {
				assignments++
				teamCounts[teamID]++
			}
		}
		for _, count := range teamCounts {
			if count > 1 {
				collisions += (count - 1)
			}
		}
	}

	// Calculate Gavel score (pairwise coverage)
	type TeamPair struct {
		TeamA string
		TeamB string
	}

	uniquePairs := make(map[TeamPair]int)
	for step := 0; step < numSteps; step++ {
		assignedTeams := []string{}
		for groupIdx := 0; groupIdx < numPairGroups; groupIdx++ {
			if matrix[step][groupIdx] != "" {
				assignedTeams = append(assignedTeams, matrix[step][groupIdx])
			}
		}

		// Record pairwise comparisons
		for i := 0; i < len(assignedTeams); i++ {
			for j := i + 1; j < len(assignedTeams); j++ {
				pair := TeamPair{}
				if assignedTeams[i] < assignedTeams[j] {
					pair.TeamA = assignedTeams[i]
					pair.TeamB = assignedTeams[j]
				} else {
					pair.TeamA = assignedTeams[j]
					pair.TeamB = assignedTeams[i]
				}
				uniquePairs[pair]++
			}
		}
	}

	totalPossiblePairs := (numTeams * (numTeams - 1)) / 2
	uniquePairsCount := len(uniquePairs)
	gavelScore := float64(uniquePairsCount) / float64(totalPossiblePairs) * 100

	totalComparisons := 0
	for _, count := range uniquePairs {
		totalComparisons += count
	}

	var avgRedundancy float64
	if uniquePairsCount > 0 {
		avgRedundancy = float64(totalComparisons) / float64(uniquePairsCount)
	}

	// === PHASE 6: SAVE SETTINGS ===
	pairGroupsJSON, err := json.Marshal(judgePairGroups)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	pairGroupsSetting := models.Setting{
		Name:  models.SettingJudgePairGroups,
		Value: string(pairGroupsJSON),
	}
	if serr := pairGroupsSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	pairOffsetJSON, err := json.Marshal(pairOffsets)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	pairOffsetSetting := models.Setting{
		Name:  models.SettingJudgePairOffset,
		Value: string(pairOffsetJSON),
	}
	if serr := pairOffsetSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	pairMultiplierJSON, err := json.Marshal(pairMultipliers)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	pairMultiplierSetting := models.Setting{
		Name:  models.SettingJudgePairMultiplier,
		Value: string(pairMultiplierJSON),
	}
	if serr := pairMultiplierSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	judgeToGroupIndexJSON, err := json.Marshal(judgeIDToGroupIdx)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	judgeToGroupIndexSetting := models.Setting{
		Name:  models.SettingJudgeToGroupIndex,
		Value: string(judgeToGroupIndexJSON),
	}
	if serr := judgeToGroupIndexSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	// Save the matrix
	matrixObj := JudgeInitMatrix{
		Steps:         numSteps,
		Groups:        numPairGroups,
		Teams:         numTeams,
		Assignments:   assignments,
		BlankCells:    blanks,
		Collisions:    collisions,
		GavelScore:    gavelScore,
		UniquePairs:   uniquePairsCount,
		TotalPairs:    totalPossiblePairs,
		AvgRedundancy: avgRedundancy,
		Matrix:        matrix,
	}

	matrixJSON, err := json.Marshal(matrixObj)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	matrixSetting := models.Setting{
		Name:  models.SettingJudgeInitMatrix,
		Value: string(matrixJSON),
	}
	if serr := matrixSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	// Save waitMinutes setting (1 minute for testing)
	waitMinutesSetting := models.Setting{
		Name:  models.SettingWaitMinutes,
		Value: "5",
	}
	if serr := waitMinutesSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	return c.JSON(bson.M{
		"message":            "judging initialized with judge pairing",
		"numTeams":           numTeams,
		"numJudges":          numJudges,
		"numPairGroups":      numPairGroups,
		"numSteps":           numSteps,
		"collisions":         collisions,
		"gavelScore":         gavelScore,
		"uniquePairs":        uniquePairsCount,
		"totalPossiblePairs": totalPossiblePairs,
		"averageRedundancy":  avgRedundancy,
		"waitMinutes":        waitMinutesSetting.Value,
	})
}

// errorMessage is a simple error wrapper for the InternalServerError function
type errorMessage struct {
	message string
}

func (e *errorMessage) Error() string {
	return e.message
}
