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
	accounts.Post("/check", AccountCheckHandler)
	accounts.Post("/register", AccountRegisterHandler)
	accounts.Post("/login", AccountLoginHandler)

	// edit
	accounts.Patch("/", models.AccountMiddleware, AccountEditHandler)

	// flags
	accounts.Get("/flags", models.AccountMiddleware, GetFlagsHandler)
}
