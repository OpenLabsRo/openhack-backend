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

func TeamDeleteHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	team := models.Team{ID: account.TeamID}
	err := team.Get()

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
