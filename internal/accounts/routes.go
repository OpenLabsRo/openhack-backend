package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

var (
	_ = errmsg.StatusError{}
)

func Routes(app *fiber.App) {
	accounts := app.Group("/accounts")

	accounts.Get("/ping", accountPingHandler)

	accounts.Get("/whoami", models.AccountMiddleware, accountWhoAmIHandler)

	// create
	accounts.Post("/check", AccountCheckHandler)
	accounts.Post("/register", AccountRegisterHandler)
	accounts.Post("/login", AccountLoginHandler)

	// edit
	accounts.Patch("/", models.AccountMiddleware, AccountEditHandler)

	// flags
	accounts.Get("/flags", models.AccountMiddleware, GetFlagsHandler)
}

// accountPingHandler verifies the accounts subsystem is responsive.
// @Summary Accounts service health check
// @Description Responds with PONG so callers can verify the accounts service group is reachable.
// @Tags Accounts Health
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /accounts/ping [get]
func accountPingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}

// accountWhoAmIHandler returns the authenticated account profile.
// @Summary Get current account profile
// @Description Reads the bearer token context and echoes the hydrated participant document.
// @Tags Accounts Identity
// @Security AccountAuth
// @Produce json
// @Success 200 {object} models.Account
// @Failure 401 {object} swagger.StatusErrorDoc
// @Router /accounts/whoami [get]
func accountWhoAmIHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	return c.JSON(account)
}
