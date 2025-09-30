package teams

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// TeamGetHandler returns the authenticated account's team document.
// @Summary Fetch the team for the authenticated account
// @Description Looks up the caller's team by membership and returns submission metadata and members.
// @Tags Teams Core
// @Security AccountAuth
// @Produce json
// @Success 200 {object} models.Team
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 409 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /teams [get]
func TeamGetHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	team := models.Team{ID: account.TeamID}
	err := team.Get()

	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(team)
}

// TeamCreateHandler creates a fresh team led by the caller.
// @Summary Create a new team led by the authenticated account
// @Description Seeds a new team with the caller as captain and returns a refreshed token.
// @Tags Teams Core
// @Security AccountAuth
// @Accept json
// @Produce json
// @Param payload body models.Team false "Optional team seed"
// @Success 200 {object} AccountTokenResponse
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 409 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /teams [post]
func TeamCreateHandler(c fiber.Ctx) error {
	var team models.Team
	json.Unmarshal(c.Body(), &team)

	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID != "" {
		return utils.StatusError(
			c, errmsg.AccountAlreadyHasTeam,
		)
	}

	err := team.Create(account.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	account.AddToTeam(team.ID)

	token := account.GenToken()

	events.Em.TeamCreate(
		account.ID,
		team.ID,
	)

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

// TeamChangeHandler updates the team's display name.
// @Summary Rename the current team
// @Description Applies a new team name and broadcasts the change to the event stream.
// @Tags Teams Core
// @Security AccountAuth
// @Accept json
// @Produce json
// @Param payload body TeamRenameRequest true "New team name"
// @Success 200 {object} models.Team
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 409 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /teams [patch]
func TeamChangeHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	var body struct {
		Name string `json:"name"`
	}
	json.Unmarshal(c.Body(), &body)

	team := models.Team{ID: account.TeamID}
	oldName, serr := team.ChangeName(body.Name)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	events.Em.TeamChangeName(
		account.ID,
		team.ID,
		oldName,
		team.Name,
	)

	return c.JSON(team)
}

// TeamDeleteHandler disbands the caller's team when they are the last member.
// @Summary Delete the current team
// @Description Removes the final member from the roster and drops the team document when safe.
// @Tags Teams Core
// @Security AccountAuth
// @Produce json
// @Success 200 {object} AccountTokenResponse
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 409 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /teams [delete]
func TeamDeleteHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	team := models.Team{ID: account.TeamID}
	err := team.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	if len(team.Members) > 1 {
		return utils.StatusError(
			c, errmsg.TeamNotEmpty,
		)
	}

	err = account.RemoveFromTeam(account.TeamID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	oldID, err := team.Delete()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	token := account.GenToken()

	events.Em.TeamDelete(
		account.ID,
		oldID,
	)

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}
