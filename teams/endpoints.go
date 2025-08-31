package teams

import (
	"backend/models"
	"backend/utils"
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func Endpoints(app *fiber.App) {
	teams := app.Group("/teams")

	teams.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	teams.Post("/", models.AccountMiddleware, func(c fiber.Ctx) error {
		var team models.Team
		json.Unmarshal(c.Body(), &team)

		account := models.Account{}
		utils.GetLocals(c, "account", &account)

		if account.TeamID != "" {
			return utils.Error(c, errors.New("user already has a team"))
		}

		err := team.Create(account.ID)
		if err != nil {
			return utils.Error(c, errors.New("could not create team"))
		}

		account.AddToTeam(team.ID)

		token := account.GenToken()

		return c.JSON(bson.M{
			"token":   token,
			"account": account,
		})
	})

	teams.Patch("/", models.AccountMiddleware, func(c fiber.Ctx) error {
		account := models.Account{}
		utils.GetLocals(c, "account", &account)

		if account.TeamID == "" {
			return utils.Error(c, errors.New("account does not belong to a team"))
		}

		var body struct {
			Name string `json:"name"`
		}
		json.Unmarshal(c.Body(), &body)

		team := models.Team{ID: account.TeamID}
		err := team.ChangeName(body.Name)
		if err != nil {
			return utils.Error(c, errors.New("could not change name of the team"))
		}

		err = team.Get()
		if err != nil {
			return utils.Error(c, err)
		}

		return c.JSON(team)
	})

	teams.Patch("/join", models.AccountMiddleware, func(c fiber.Ctx) error {
		// get all info on the team
		team := models.Team{ID: c.Query("id")}
		err := team.Get()
		if err != nil {
			// return utils.Error(c, errors.New("could not find team"))
			return utils.Error(c, err)
		}

		// unmarshal the body
		var account models.Account
		utils.GetLocals(c, "account", &account)

		if account.TeamID != "" {
			return utils.Error(c, errors.New("you already belong to a team"))
		}

		// adding the member to the team
		err = team.AddMember(account.ID)
		if err != nil {
			return utils.Error(c, err)
		}

		err = account.AddToTeam(team.ID)
		if err != nil {
			return utils.Error(c, errors.New("could not join the team"))
		}

		token := account.GenToken()

		return c.JSON(bson.M{
			"token":   token,
			"account": account,
		})

	})

	teams.Patch("/leave", models.AccountMiddleware, func(c fiber.Ctx) error {
		// get all info on the team
		team := models.Team{ID: c.Query("id")}
		err := team.Get()
		if err != nil {
			// return utils.Error(c, errors.New("could not find team"))
			return utils.Error(c, err)
		}

		// unmarshal the body
		var account models.Account
		utils.GetLocals(c, "account", &account)

		// removing the member to the team
		err = team.RemoveMember(account.ID)
		if err != nil {
			return utils.Error(c, err)
		}

		err = account.RemoveFromTeam(team.ID)
		if err != nil {
			return utils.Error(c, errors.New("could not leave the team"))
		}

		token := account.GenToken()

		return c.JSON(bson.M{
			"token":   token,
			"account": account,
		})
	})

	teams.Patch("/kick", models.AccountMiddleware, func(c fiber.Ctx) error {
		// getting the account
		var account models.Account
		utils.GetLocals(c, "account", &account)

		// finding the account to remove
		accountToRemove := models.Account{ID: c.Query("id")}

		// the team
		team := models.Team{ID: account.TeamID}
		err := team.Get()
		if err != nil {
			return utils.Error(c, errors.New("could not find team"))
		}

		// removing the member to the team
		err = team.RemoveMember(accountToRemove.ID)
		if err != nil {
			return utils.Error(c, err)
		}

		err = accountToRemove.RemoveFromTeam(team.ID)
		if err != nil {
			return utils.Error(c, errors.New("could not leave the team"))
		}

		return c.Status(200).SendString("OK")
	})

	teams.Delete("", models.AccountMiddleware, func(c fiber.Ctx) error {
		account := models.Account{}
		utils.GetLocals(c, "account", &account)

		team := models.Team{ID: account.TeamID}
		err := team.Get()

		if len(team.Members) > 1 {
			return utils.Error(c, errors.New("teammates still in team"))
		}

		err = account.RemoveFromTeam(account.TeamID)
		if err != nil {
			return utils.Error(c, errors.New("could not remove from team"))
		}

		err = team.Delete()
		if err != nil {
			return utils.Error(c, errors.New("could not delete"))
		}

		token := account.GenToken()

		return c.JSON(bson.M{
			"token":   token,
			"account": account,
		})
	})
}
