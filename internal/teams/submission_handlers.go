package teams

import (
	"backend/internal/errmsg"
	"backend/internal/events"
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
	oldName, serr := team.ChangeSubmissionName(body.Name)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	events.Em.SubmissionChangeName(
		account.ID,
		team.ID,
		oldName,
		team.Submission.Name,
	)

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
	oldDesc, serr := team.ChangeSubmissionDesc(body.Desc)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	events.Em.SubmissionChangeDesc(
		account.ID,
		team.ID,
		oldDesc,
		team.Submission.Desc,
	)

	return c.JSON(team)
}

func TeamSubmissionChangeRepoHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	var body struct {
		Repo string `json:"repo"`
	}
	json.Unmarshal(c.Body(), &body)

	team := models.Team{ID: account.TeamID}
	oldRepo, serr := team.ChangeSubmissionRepo(body.Repo)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	events.Em.SubmissionChangeRepo(
		account.ID,
		team.ID,
		oldRepo,
		team.Submission.Repo,
	)

	return c.JSON(team)
}

func TeamSubmissionChangePresHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	var body struct {
		Pres string `json:"pres"`
	}
	json.Unmarshal(c.Body(), &body)

	team := models.Team{ID: account.TeamID}
	oldPres, serr := team.ChangeSubmissionPres(body.Pres)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	events.Em.SubmissionChangePres(
		account.ID,
		team.ID,
		oldPres,
		team.Submission.Pres,
	)

	return c.JSON(team)
}
