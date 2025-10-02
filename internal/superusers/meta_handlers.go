package superusers

import (
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

// superUserPingHandler responds to health probes for the superuser subsystem.
// @Summary Superuser service health check
// @Description Confirms the privileged routes segment is reachable by returning a simple PONG.
// @Tags Superusers Meta
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /superusers/meta/ping [get]
func superUserPingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}

// superUserWhoAmIHandler reveals the authenticated superuser context.
// @Summary Inspect the current superuser context
// @Description Echoes the active superuser payload so operators can verify their scopes.
// @Tags Superusers Meta
// @Security SuperUserAuth
// @Produce json
// @Success 200 {object} models.SuperUser
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Router /superusers/meta/whoami [get]
func superUserWhoAmIHandler(c fiber.Ctx) error {
	su := models.SuperUser{}
	utils.GetLocals(c, "superuser", &su)

	return c.JSON(su)
}
