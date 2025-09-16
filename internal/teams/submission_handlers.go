package teams

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func TeamSubmissionChangeNameHandler(c fiber.Ctx) error {
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
	serr := team.ChangeSubmissionName(body.Name)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	return c.JSON(team)
}

func TeamSubmissionChangeDescHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	var body struct {
		Desc string `json:"desc"`
	}
	json.Unmarshal(c.Body(), &body)

	team := models.Team{ID: account.TeamID}
	serr := team.ChangeSubmissionDesc(body.Desc)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	return c.JSON(team)
}

func TeamSubmissionChangeLinkHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	var body struct {
		Link string `json:"link"`
	}
	json.Unmarshal(c.Body(), &body)

	team := models.Team{ID: account.TeamID}
	serr := team.ChangeSubmissionLink(body.Link)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	return c.JSON(team)
}
