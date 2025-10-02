package meta

import (
	"backend/internal/env"

	"github.com/gofiber/fiber/v3"
)

// pingHandler answers with a plain "PONG" for service uptime checks.
// @Summary Health check
// @Description Lightweight heartbeat used by load balancers to confirm the OpenHack API is alive.
// @Tags General
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /meta/ping [get]
func pingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}

// versionHandler prints the current deployment version for observability.
// @Summary Current deployment version
// @Description Exposes the semantic version bundled with the running process for smoke tests.
// @Tags General
// @Produce plain
// @Success 200 {string} string "25.10.01.0"
// @Router /meta/version [get]
func versionHandler(c fiber.Ctx) error {
	return c.SendString(env.VERSION)
}
