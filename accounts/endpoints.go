package accounts

import (
	"backend/models"
	"backend/utils"
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func Endpoints(app *fiber.App) {
	accounts := app.Group("/accounts")

	accounts.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	accounts.Post("/initialize", func(c fiber.Ctx) error {
		var account models.Account
		json.Unmarshal(c.Body(), &account)

		err := account.Initialize()
		if err != nil {
			return utils.Error(c, errors.New("could not initialize account"))
		}

		return c.JSON(account)
	})

	accounts.Post("/register", func(c fiber.Ctx) error {
		id := c.Query("id")

		var body struct {
			Password string `json:"password" bson:"password"`
		}
		json.Unmarshal(c.Body(), &body)

		account := models.Account{
			ID: id,
		}

		err := account.CreatePassword(body.Password)
		if err != nil {
			return utils.Error(c, err)
		}

		token := account.GenToken()

		return c.JSON(bson.M{
			"token":   token,
			"account": account,
		})
	})

	accounts.Post("/login", func(c fiber.Ctx) error {
		var body struct {
			Email    string `json:"email" bson:"email"`
			Password string `json:"password" bson:"password"`
		}
		json.Unmarshal(c.Body(), &body)

		account := models.Account{}
		account.GetByEmail(body.Email)

		if bcrypt.CompareHashAndPassword(
			[]byte(account.Password),
			[]byte(body.Password),
		) != nil {
			return utils.Error(c, errors.New("incorrect password"))
		}

		token := account.GenToken()

		return c.JSON(bson.M{
			"token":   token,
			"account": account,
		})

		// return c.SendString("ok")
	})

	accounts.Get("/whoami", models.AccountMiddleware, func(c fiber.Ctx) error {
		account := models.Account{}
		utils.GetLocals(c, "account", &account)

		return c.JSON(account)
	})

	accounts.Post("/edit", models.AccountMiddleware, func(c fiber.Ctx) error {
		var fields models.AccountFields
		json.Unmarshal(c.Body(), &fields)

		account := models.Account{}
		utils.GetLocals(c, "account", &account)

		account.EditFields(fields)

		token := account.GenToken()

		return c.JSON(bson.M{
			"token":   token,
			"account": account,
		})
	})
}
