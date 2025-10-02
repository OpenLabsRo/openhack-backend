package teams

import "github.com/gofiber/fiber/v3"

// teamPingHandler ensures the teams subsystem responds to health probes.
// @Summary Teams service health check
// @Description Returns a PONG from the teams group so orchestration checks can verify connectivity.
// @Tags Teams Meta
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /teams/meta/ping [get]
func teamPingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}
