package teams

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func TeamJoinHandler(c fiber.Ctx) error {
	// get all info on the team
	team := models.Team{ID: c.Query("id")}
	err := team.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.TeamNotFound,
		)
	}

	// unmarshal the body
	var account models.Account
	utils.GetLocals(c, "account", &account)

	if account.TeamID != "" {
		return utils.StatusError(
			c, errmsg.AccountAlreadyHasTeam,
		)
	}

	// adding the member to the team
	serr := team.AddMember(account.ID)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	err = account.AddToTeam(team.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError,
		)
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

func TeamLeaveHandler(c fiber.Ctx) error {
	// unmarshal the body
	var account models.Account
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	team := models.Team{ID: account.TeamID}
	err := team.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.TeamNotFound)
	}

	// removing the member to the team
	serr := team.RemoveMember(account.ID)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	err = account.RemoveFromTeam(team.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError,
		)
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

func TeamKickHandler(c fiber.Ctx) error {
	// getting the local account
	var account models.Account
	utils.GetLocals(c, "account", &account)

	// finding the account to remove
	accountToRemove := models.Account{ID: c.Query("id")}
	err := accountToRemove.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.AccountNotFound,
		)
	}

	// the team
	team := models.Team{ID: account.TeamID}
	err = team.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.TeamNotFound,
		)
	}

	// removing the member to the team
	serr := team.RemoveMember(accountToRemove.ID)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	err = accountToRemove.RemoveFromTeam(team.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError,
		)
	}

	return c.Status(200).SendString("OK")
}
