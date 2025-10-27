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

	// Extract team IDs and shuffle
	teamIDs := make([]string, numTeams)
	for i, team := range teams {
		teamIDs[i] = team.ID
	}
	rand.Shuffle(len(teamIDs), func(i, j int) {
		teamIDs[i], teamIDs[j] = teamIDs[j], teamIDs[i]
	})

	// Create or update teamOrder setting (JSON string)
	teamOrderJSON, err := json.Marshal(teamIDs)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	teamOrderSetting := models.Setting{
		Name:  models.SettingTeamOrder,
		Value: string(teamOrderJSON),
	}
	if serr := teamOrderSetting.Save(); serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	events.Em.JudgeInitTeamOrderSet(superuser.Username, teamIDs)

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

	// Calculate and create judgeOffset
	// Each judge index is offset by its position, wrapping around based on number of teams
	judgeOffsets := make([]int, numJudges)
	for i := range numJudges {
		judgeOffsets[i] = i % numTeams
	}

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

	events.Em.JudgeInitOffsetSet(superuser.Username, judgeOffsets, numTeams)

	return c.JSON(JudgeInitResponse{
		TeamOrder:   teamIDs,
		JudgeOrder:  judgeIDs,
		JudgeOffset: judgeOffsets,
		NumTeams:    numTeams,
		NumJudges:   numJudges,
	})
}

// errorMessage is a simple error wrapper for the InternalServerError function
type errorMessage struct {
	message string
}

func (e *errorMessage) Error() string {
	return e.message
}
