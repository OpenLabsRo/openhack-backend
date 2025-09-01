package teams

import (
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func TeamCreate(c fiber.Ctx) error {
	var team models.Team
	json.Unmarshal(c.Body(), &team)

	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID != "" {
		return utils.Error(c, http.StatusConflict, errors.New("user already has a team"))
	}

	err := team.Create(account.ID)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, errors.New("could not create team"))
	}

	account.AddToTeam(team.ID)

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

func TeamUpdate(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.Error(c, http.StatusConflict, errors.New("account does not belong to a team"))
	}

	var body struct {
		Name string `json:"name"`
	}
	json.Unmarshal(c.Body(), &body)

	team := models.Team{ID: account.TeamID}
	err := team.ChangeName(body.Name)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, errors.New("could not change name of the team"))
	}

	err = team.Get()
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, err)
	}

	return c.JSON(team)
}

func TeamDelete(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	team := models.Team{ID: account.TeamID}
	err := team.Get()

	if len(team.Members) > 1 {
		return utils.Error(c, http.StatusConflict, errors.New("teammates still in team"))
	}

	err = account.RemoveFromTeam(account.TeamID)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, errors.New("could not remove from team"))
	}

	err = team.Delete()
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, errors.New("could not delete"))
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}
