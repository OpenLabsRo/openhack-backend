package accounts

import (
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

func Routes(app *fiber.App) {
	accounts := app.Group("/accounts")

	accounts.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	accounts.Get("/whoami", models.AccountMiddleware, func(c fiber.Ctx) error {
		account := models.Account{}
		utils.GetLocals(c, "account", &account)

		return c.JSON(account)
	})

	// create
	accounts.Post("/check", AccountCheck)
	accounts.Post("/register", AccountRegister)
	accounts.Post("/login", AccountLogin)

	// edit
	accounts.Patch("/", models.AccountMiddleware, AccountEdit)
}
