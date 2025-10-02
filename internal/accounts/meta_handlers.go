package accounts

import (
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

// accountPingHandler verifies the accounts subsystem is responsive.
// @Summary Accounts service health check
// @Description Responds with PONG so callers can verify the accounts service group is reachable.
// @Tags Accounts Meta
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /accounts/meta/ping [get]
func accountPingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}

// accountWhoAmIHandler returns the authenticated account profile.
// @Summary Get current account profile
// @Description Reads the bearer token context and echoes the hydrated participant document.
// @Tags Accounts Meta
// @Security AccountAuth
// @Produce json
// @Success 200 {object} models.Account
// @Failure 401 {object} errmsg._AccountNoToken
// @Router /accounts/meta/whoami [get]
func accountWhoAmIHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	return c.JSON(account)
}
