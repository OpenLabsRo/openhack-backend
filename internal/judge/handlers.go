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
// @Failure 200 {object} errmsg._JudgingFinished
// @Failure 401 {object} errmsg._AccountNoToken
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
