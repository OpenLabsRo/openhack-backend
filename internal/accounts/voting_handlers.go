package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// votingStatusHandler returns voting status and finalists for the participant
// @Summary Get voting status
// @Description Returns whether voting is open, if user has voted, and the list of finalists
// @Tags Accounts Voting
// @Security AccountAuth
// @Produce json
// @Success 200 {object} VotingStatusResponse
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /accounts/voting/status [get]
func votingStatusHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)
	err := account.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	// Get finalists from settings
	finalist1Setting := &models.Setting{Name: models.SettingFinalist1}
	finalist2Setting := &models.Setting{Name: models.SettingFinalist2}
	finalist3Setting := &models.Setting{Name: models.SettingFinalist3}

	var finalists []string
	if finalist1Setting.Get() == errmsg.EmptyStatusError {
		if val, ok := finalist1Setting.Value.(string); ok {
			finalists = append(finalists, val)
		}
	}
	if finalist2Setting.Get() == errmsg.EmptyStatusError {
		if val, ok := finalist2Setting.Value.(string); ok {
			finalists = append(finalists, val)
		}
	}
	if finalist3Setting.Get() == errmsg.EmptyStatusError {
		if val, ok := finalist3Setting.Value.(string); ok {
			finalists = append(finalists, val)
		}
	}

	response := map[string]interface{}{
		"votingOpen": true,
		"hasVoted":   account.HasVoted,
		"finalists":  finalists,
	}

	return c.JSON(response)
}

// votingFinalistsHandler returns the list of finalists with team information
// @Summary Get finalists
// @Description Returns the 3 finalist teams (middle, right, left order)
// @Tags Accounts Voting
// @Security AccountAuth
// @Produce json
// @Success 200 {object} VotingFinalistsResponse
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 403 {object} errmsg._FlagRequired
// @Failure 500 {object} errmsg._InternalServerError
// @Router /accounts/voting/finalists [get]
func votingFinalistsHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	// Check if user has already voted
	if account.HasVoted {
		statusErr := errmsg.NewStatusError(409, "you have already voted")
		return utils.StatusError(c, statusErr)
	}

	// Get finalists from settings
	finalist1Setting := &models.Setting{Name: models.SettingFinalist1}
	finalist2Setting := &models.Setting{Name: models.SettingFinalist2}
	finalist3Setting := &models.Setting{Name: models.SettingFinalist3}

	finalists1 := finalist1Setting.Get() == errmsg.EmptyStatusError
	finalists2 := finalist2Setting.Get() == errmsg.EmptyStatusError
	finalists3 := finalist3Setting.Get() == errmsg.EmptyStatusError

	if !finalists1 || !finalists2 || !finalists3 {
		return utils.StatusError(c, errmsg.InternalServerError(nil))
	}

	var finalistTeams []map[string]interface{}
	for _, setting := range []*models.Setting{finalist1Setting, finalist2Setting, finalist3Setting} {
		if teamID, ok := setting.Value.(string); ok {
			team := &models.Team{ID: teamID}
			if err := team.Get(); err == nil {
				finalistTeams = append(finalistTeams, map[string]interface{}{
					"id":   team.ID,
					"name": team.Name,
				})
			}
		}
	}

	return c.JSON(map[string]interface{}{
		"finalists": finalistTeams,
	})
}

// votingCastVoteHandler casts a vote for a finalist
// @Summary Cast a vote
// @Description Records a vote for one of the finalists
// @Tags Accounts Voting
// @Security AccountAuth
// @Accept json
// @Produce json
// @Param payload body VotingCastRequest true "Team ID to vote for"
// @Success 200 {object} VotingCastResponse
// @Failure 400 {object} errmsg._BadRequest
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 403 {object} errmsg._FlagRequired
// @Failure 409 {object} errmsg._AccountAlreadyRegistered
// @Failure 500 {object} errmsg._InternalServerError
// @Router /accounts/voting/vote [post]
func votingCastVoteHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)
	err := account.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	// Check if user has already voted
	if account.HasVoted {
		statusErr := errmsg.NewStatusError(409, "you have already voted")
		return utils.StatusError(c, statusErr)
	}

	// Parse request body
	var body struct {
		TeamID string `json:"teamID"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		statusErr := errmsg.NewStatusError(400, "invalid request body")
		return utils.StatusError(c, statusErr)
	}

	if body.TeamID == "" {
		statusErr := errmsg.NewStatusError(400, "teamID is required")
		return utils.StatusError(c, statusErr)
	}

	// Validate team ID is one of the finalists
	finalist1Setting := &models.Setting{Name: models.SettingFinalist1}
	finalist2Setting := &models.Setting{Name: models.SettingFinalist2}
	finalist3Setting := &models.Setting{Name: models.SettingFinalist3}

	finalist1Setting.Get()
	finalist2Setting.Get()
	finalist3Setting.Get()

	isFinalist := false
	if val, ok := finalist1Setting.Value.(string); ok && val == body.TeamID {
		isFinalist = true
	} else if val, ok := finalist2Setting.Value.(string); ok && val == body.TeamID {
		isFinalist = true
	} else if val, ok := finalist3Setting.Value.(string); ok && val == body.TeamID {
		isFinalist = true
	}

	if !isFinalist {
		statusErr := errmsg.NewStatusError(400, "invalid finalist team")
		return utils.StatusError(c, statusErr)
	}

	// Set hasVoted = true and cache it
	errSetVote := account.SetHasVoted()
	if errSetVote != nil {
		return utils.StatusError(c, errmsg.InternalServerError(errSetVote))
	}

	// Create anonymous vote
	vote := &models.Vote{
		Choice: body.TeamID,
	}
	errCreate := vote.Create()
	if errCreate != nil {
		return utils.StatusError(c, errmsg.InternalServerError(errCreate))
	}

	return c.Status(http.StatusOK).JSON(map[string]string{
		"message": "Vote recorded successfully",
	})
}
