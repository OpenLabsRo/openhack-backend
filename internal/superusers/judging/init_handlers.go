package judging

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"math/rand"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// getCoprimeMultipliers returns a list of numbers coprime to n
// These are used to generate diverse permutations while maintaining some structure
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

// judgeInitHandler initializes judging settings: team order, judge order, and judge offsets.
// @Summary Initialize judging configuration
// @Description Scrambles teams and judges randomly, creates teamOrder and judgeOrder settings, and calculates judge offsets with wrapping.
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Produce json
// @Success 200 {object} JudgeInitResponse
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judging/init [post]
func judgeInitHandler(c fiber.Ctx) error {
	superuser := models.SuperUser{}
	utils.GetLocals(c, "superuser", &superuser)

	// Fetch all teams
	cursor, err := db.Teams.Find(db.Ctx, bson.M{})
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	defer cursor.Close(db.Ctx)

	var teams []models.Team
	if err = cursor.All(db.Ctx, &teams); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	// Fetch all judges
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

	// Extract team IDs and create TWO base orders for better coverage
	teamIDs := make([]string, numTeams)
	for i, team := range teams {
		teamIDs[i] = team.ID
	}

	// Create first base order (shuffled)
	teamOrderA := make([]string, numTeams)
	copy(teamOrderA, teamIDs)
	rand.Shuffle(len(teamOrderA), func(i, j int) {
		teamOrderA[i], teamOrderA[j] = teamOrderA[j], teamOrderA[i]
	})

	// Create second base order (different shuffle)
	teamOrderB := make([]string, numTeams)
	copy(teamOrderB, teamIDs)
	rand.Shuffle(len(teamOrderB), func(i, j int) {
		teamOrderB[i], teamOrderB[j] = teamOrderB[j], teamOrderB[i]
	})

	// Ensure the two orders are actually different
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
		// Re-shuffle B if identical to A
		rand.Shuffle(len(teamOrderB), func(i, j int) {
			teamOrderB[i], teamOrderB[j] = teamOrderB[j], teamOrderB[i]
		})
	}

	// Store both base orders as JSON arrays in a single setting
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

	// Extract judge IDs and shuffle
	judgeIDs := make([]string, numJudges)
	for i, judge := range judges {
		judgeIDs[i] = judge.ID
	}
	rand.Shuffle(len(judgeIDs), func(i, j int) {
		judgeIDs[i], judgeIDs[j] = judgeIDs[j], judgeIDs[i]
	})

	// Create or update judgeOrder setting (JSON string)
	judgeOrderJSON, err := json.Marshal(judgeIDs)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	judgeOrderSetting := models.Setting{
		Name:  models.SettingJudgeOrder,
		Value: string(judgeOrderJSON),
	}
	if serr := judgeOrderSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	events.Em.JudgeInitJudgeOrderSet(superuser.Username, judgeIDs)

	// Calculate judge offsets, multipliers, and base order assignments
	// Split judges between two base orders for better coverage
	judgeOffsets := make([]int, numJudges)
	judgeMultipliers := make([]int, numJudges)
	judgeBaseOrders := make([]int, numJudges) // 0 = use teamOrderA, 1 = use teamOrderB

	// Get all numbers coprime to numTeams
	coprimes := getCoprimeMultipliers(numTeams)
	if len(coprimes) == 0 {
		return utils.StatusError(c, errmsg.InternalServerError(
			&errorMessage{message: "no coprime multipliers found"},
		))
	}

	for i := range numJudges {
		judgeOffsets[i] = i % numTeams
		// Cycle through coprime multipliers
		judgeMultipliers[i] = coprimes[i%len(coprimes)]
		// Alternate between base orders
		judgeBaseOrders[i] = i % 2
	}

	// Shuffle multipliers for more randomness
	rand.Shuffle(len(judgeMultipliers), func(i, j int) {
		judgeMultipliers[i], judgeMultipliers[j] = judgeMultipliers[j], judgeMultipliers[i]
	})

	// Shuffle base order assignments
	rand.Shuffle(len(judgeBaseOrders), func(i, j int) {
		judgeBaseOrders[i], judgeBaseOrders[j] = judgeBaseOrders[j], judgeBaseOrders[i]
	})

	// Create or update judgeOffset setting (JSON string)
	judgeOffsetJSON, err := json.Marshal(judgeOffsets)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	judgeOffsetSetting := models.Setting{
		Name:  models.SettingJudgeOffset,
		Value: string(judgeOffsetJSON),
	}
	if serr := judgeOffsetSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	// Create or update judgeMultiplier setting (JSON string)
	judgeMultiplierJSON, err := json.Marshal(judgeMultipliers)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	judgeMultiplierSetting := models.Setting{
		Name:  models.SettingJudgeMultiplier,
		Value: string(judgeMultiplierJSON),
	}
	if serr := judgeMultiplierSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	// Create or update judgeBaseOrder setting (JSON string)
	judgeBaseOrderJSON, err := json.Marshal(judgeBaseOrders)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	judgeBaseOrderSetting := models.Setting{
		Name:  models.SettingJudgeBaseOrder,
		Value: string(judgeBaseOrderJSON),
	}
	if serr := judgeBaseOrderSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	events.Em.JudgeInitOffsetSet(superuser.Username, judgeOffsets, numTeams)

	return c.JSON(JudgeInitResponse{
		TeamOrderA:      teamOrderA,
		TeamOrderB:      teamOrderB,
		JudgeOrder:      judgeIDs,
		JudgeOffset:     judgeOffsets,
		JudgeMultiplier: judgeMultipliers,
		JudgeBaseOrder:  judgeBaseOrders,
		NumTeams:        numTeams,
		NumJudges:       numJudges,
	})
}

// errorMessage is a simple error wrapper for the InternalServerError function
type errorMessage struct {
	message string
}

func (e *errorMessage) Error() string {
	return e.message
}
