package judge

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// JudgeUpgradeHandler exchanges a 2-minute connect token for a full 24-hour judge token.
// @Summary Upgrade connect token to full session token
// @Description Validates the 2-minute connect token and mints a new 24-hour token for judge platform access.
// @Tags Judges Auth
// @Accept json
// @Produce json
// @Param payload body JudgeUpgradeRequest true "Connect token"
// @Success 200 {object} JudgeUpgradeResponse
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /judge/upgrade [post]
func JudgeUpgradeHandler(c fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}
	json.Unmarshal(c.Body(), &body)

	judge := models.Judge{}
	err := judge.ParseToken(body.Token)
	if err != nil {
		return utils.StatusError(c, errmsg.AccountNoToken)
	}

	if judge.ID == "" {
		return utils.StatusError(c, errmsg.AccountNoToken)
	}

	fullToken := judge.GenToken()

	events.Em.JudgeTokenUpgraded(judge.ID)

	return c.JSON(bson.M{
		"token": fullToken,
		"judge": judge,
	})
}

// nextTeamHandler retrieves the next team for the authenticated judge.
// @Summary Get next team for judging
// @Description Returns the next team ID for the judge to evaluate. Applies offset on first call, then cycles through teams until judging is complete.
// @Tags Judges
// @Security JudgeAuth
// @Produce json
// @Success 200 {object} NextTeamResponse
// @Failure 202 {object} errmsg._JudgeResting
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 410 {object} errmsg._JudgingFinished
// @Failure 500 {object} errmsg._InternalServerError
// @Router /judge/next-team [post]
func nextTeamHandler(c fiber.Ctx) error {
	judge := models.Judge{}
	utils.GetLocals(c, "judge", &judge)

	teamID, serr := judge.GetNextTeam()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	events.Em.JudgeNextTeamRequested(judge.ID, teamID)

	return c.JSON(bson.M{
		"teamID": teamID,
	})
}

// currentTeamHandler retrieves the currently assigned team for the judge.
// @Summary Get current team for judging
// @Description Returns the full team details for the judge's current assignment based on their rotation state.
// @Tags Judges
// @Security JudgeAuth
// @Produce json
// @Success 200 {object} models.Team
// @Failure 202 {object} errmsg._JudgeResting
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 404 {object} errmsg._TeamNotFound
// @Failure 410 {object} errmsg._JudgingFinished
// @Failure 500 {object} errmsg._InternalServerError
// @Router /judge/current-team [get]
func currentTeamHandler(c fiber.Ctx) error {
	judge := models.Judge{}
	utils.GetLocals(c, "judge", &judge)

	if judge.ID == "" {
		return utils.StatusError(c, errmsg.AccountNoToken)
	}

	if err := judge.Get(); err != nil {
		return utils.StatusError(c, errmsg.JudgeNotFound)
	}

	teamID, serr := judge.GetCurrentTeamID()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	team := models.Team{ID: teamID}
	if err := team.Get(); err != nil {
		return utils.StatusError(c, errmsg.TeamNotFound)
	}

	return c.JSON(team)
}

// getTeamHandler retrieves team information by team ID for the authenticated judge.
// @Summary Get team information
// @Description Returns detailed information about a team including submission data.
// @Tags Judges
// @Security JudgeAuth
// @Produce json
// @Param id query string true "Team ID"
// @Success 200 {object} models.Team
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 404 {object} errmsg._TeamNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /judge/team [get]
func getTeamHandler(c fiber.Ctx) error {
	teamID := c.Query("id")
	if teamID == "" {
		return utils.StatusError(c, errmsg.TeamNotFound)
	}

	team := models.Team{ID: teamID}
	err := team.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.TeamNotFound)
	}

	return c.JSON(team)
}

// createJudgmentHandler records a pairwise comparison judgment between two teams.
// @Summary Create a judgment
// @Description Records a judgment where a judge compares two teams and selects a winner.
// @Tags Judges
// @Security JudgeAuth
// @Accept json
// @Produce json
// @Param payload body CreateJudgmentRequest true "Judgment details"
// @Success 200 {object} models.Judgment
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /judge/judgment [post]
func createJudgmentHandler(c fiber.Ctx) error {
	var body struct {
		WinningTeamID string `json:"winningTeamID"`
		LosingTeamID  string `json:"losingTeamID"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	judge := models.Judge{}
	utils.GetLocals(c, "judge", &judge)

	judgment := models.Judgment{
		WinningTeamID: body.WinningTeamID,
		LosingTeamID:  body.LosingTeamID,
		JudgeID:       judge.ID,
	}

	err := judgment.Create()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	events.Em.JudgmentCreated(judge.ID, judgment.WinningTeamID, judgment.LosingTeamID)

	return c.JSON(judgment)
}
