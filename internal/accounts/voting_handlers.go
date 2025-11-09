package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

var finalistSettingNames = []string{
	models.SettingFinalist1,
	models.SettingFinalist2,
	models.SettingFinalist3,
	models.SettingFinalist4,
	models.SettingFinalist5,
}

func loadFinalistTeams(requireAll bool) ([]models.Team, error) {
	teams := make([]models.Team, 0, len(finalistSettingNames))

	for _, name := range finalistSettingNames {
		setting := &models.Setting{Name: name}
		status := setting.Get()
		if status != errmsg.EmptyStatusError {
			if requireAll {
				return nil, fmt.Errorf("failed to load setting %s: %s", name, status.Message)
			}
			continue
		}

		teamID, ok := setting.Value.(string)
		if !ok || teamID == "" {
			if requireAll {
				return nil, fmt.Errorf("setting %s contains no team identifier", name)
			}
			continue
		}

		team := models.Team{ID: teamID}
		if err := team.Get(); err != nil {
			return nil, err
		}

		teams = append(teams, team)
	}

	if requireAll && len(teams) != len(finalistSettingNames) {
		return nil, fmt.Errorf("expected %d finalists, found %d", len(finalistSettingNames), len(teams))
	}

	return teams, nil
}

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

	finalistTeams, err := loadFinalistTeams(false)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	response := map[string]interface{}{
		"votingOpen": true,
		"hasVoted":   account.HasVoted,
		"finalists":  finalistTeams,
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

	finalistTeams, err := loadFinalistTeams(true)
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
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
	json.Unmarshal(c.Body(), &body)

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
