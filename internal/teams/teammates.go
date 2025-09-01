package teams

import (
	"backend/internal/models"
	"backend/internal/utils"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func TeamJoin(c fiber.Ctx) error {

	// get all info on the team
	team := models.Team{ID: c.Query("id")}
	err := team.Get()
	if err != nil {
		// return utils.Error(c, errors.New("could not find team"))
		return utils.Error(c, http.StatusNotFound, err)
	}

	// unmarshal the body
	var account models.Account
	utils.GetLocals(c, "account", &account)

	if account.TeamID != "" {
		return utils.Error(c, http.StatusConflict, errors.New("you already belong to a team"))
	}

	// adding the member to the team
	err = team.AddMember(account.ID)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, err)
	}

	err = account.AddToTeam(team.ID)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, errors.New("could not join the team"))
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

func TeamLeave(c fiber.Ctx) error {
	// get all info on the team
	team := models.Team{ID: c.Query("id")}
	err := team.Get()
	if err != nil {
		// return utils.Error(c, errors.New("could not find team"))
		return utils.Error(c, http.StatusNotFound, err)
	}

	// unmarshal the body
	var account models.Account
	utils.GetLocals(c, "account", &account)

	// removing the member to the team
	err = team.RemoveMember(account.ID)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, err)
	}

	err = account.RemoveFromTeam(team.ID)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, errors.New("could not leave the team"))
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

func TeamKick(c fiber.Ctx) error {
	// getting the account
	var account models.Account
	utils.GetLocals(c, "account", &account)

	// finding the account to remove
	accountToRemove := models.Account{ID: c.Query("id")}

	// the team
	team := models.Team{ID: account.TeamID}
	err := team.Get()
	if err != nil {
		return utils.Error(c, http.StatusNotFound, errors.New("could not find team"))
	}

	// removing the member to the team
	err = team.RemoveMember(accountToRemove.ID)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, err)
	}

	err = accountToRemove.RemoveFromTeam(team.ID)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, errors.New("could not leave the team"))
	}

	return c.Status(200).SendString("OK")
}
