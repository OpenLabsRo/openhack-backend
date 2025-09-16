package teams

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func TeamGetTeammatesHandler(c fiber.Ctx) error {
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

	members, err := team.GetMembers()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(members)
}

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
	serr := team.AddMember(account.ID, account)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	err = account.AddToTeam(team.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	token := account.GenToken()

	events.Em.TeamMemberJoin(
		account.ID,
		team.ID,
	)

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
			c, errmsg.InternalServerError(err),
		)
	}

	token := account.GenToken()

	events.Em.TeamMemberLeave(
		account.ID,
		team.ID,
	)
	events.Em.AccountTeamExit(
		account.ID,
		team.ID,
	)

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

	events.Em.TeamMemberKick(
		account.ID,
		team.ID,
		accountToRemove.ID,
	)

	err = accountToRemove.RemoveFromTeam(team.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	events.Em.AccountTeamExit(
		accountToRemove.ID,
		team.ID,
	)

	return c.Status(200).SendString("OK")
}
