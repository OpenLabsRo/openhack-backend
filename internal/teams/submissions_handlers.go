package teams

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

// TeamSubmissionChangeNameHandler updates the submission name field.
// @Summary Update the team submission name
// @Description Records a new project name and emits an event so judges see the latest label.
// @Tags Teams Submissions
// @Security AccountAuth
// @Accept json
// @Produce json
// @Param payload body SubmissionNameRequest true "Submission name"
// @Success 200 {object} models.Team
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 409 {object} errmsg._AccountHasNoTeam
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/submissions/name [patch]
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

// TeamSubmissionChangeDescHandler updates the submission summary text.
// @Summary Update the team submission description
// @Description Persists revised blurb content so reviewers receive the latest write-up.
// @Tags Teams Submissions
// @Security AccountAuth
// @Accept json
// @Produce json
// @Param payload body SubmissionDescRequest true "Submission description"
// @Success 200 {object} models.Team
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 409 {object} errmsg._AccountHasNoTeam
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/submissions/desc [patch]
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

// TeamSubmissionChangeRepoHandler updates the repository link.
// @Summary Update the team submission repo
// @Description Stores a repository URL so judges can inspect source material from the dashboard.
// @Tags Teams Submissions
// @Security AccountAuth
// @Accept json
// @Produce json
// @Param payload body SubmissionRepoRequest true "Submission repository"
// @Success 200 {object} models.Team
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 409 {object} errmsg._AccountHasNoTeam
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/submissions/repo [patch]
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

// TeamSubmissionChangePresHandler updates the presentation artifact link.
// @Summary Update the team submission presentation
// @Description Keeps the presentation URL in sync for the demo day kiosk.
// @Tags Teams Submissions
// @Security AccountAuth
// @Accept json
// @Produce json
// @Param payload body SubmissionPresRequest true "Submission presentation"
// @Success 200 {object} models.Team
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 409 {object} errmsg._AccountHasNoTeam
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/submissions/pres [patch]
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
