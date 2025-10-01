package internal

import (
	"backend/internal/accounts"
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/events"
	"backend/internal/superusers"
	"backend/internal/teams"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

func getEmitterConfig(deployment string) events.Config {
	switch deployment {
	case "test":
		return events.Config{
			Buffer:     1000,
			BatchSize:  50,
			FlushEvery: 100 * time.Millisecond,
		}
	case "dev":
		return events.Config{
			Buffer:     1000,
			BatchSize:  50,
			FlushEvery: 1 * time.Second,
		}
	default:
		return events.Config{
			Buffer:     1000,
			BatchSize:  50,
			FlushEvery: 2 * time.Second,
		}
	}
}

func SetupApp(deployment string, envRoot string, appVersion string) *fiber.App {
	app := fiber.New()

	env.Init(envRoot, appVersion)

	if err := db.InitDB(deployment); err != nil {
		log.Fatal("Could not connect to MongoDB")
		return nil
	}

	if err := db.InitCache(deployment); err != nil {
		log.Fatal("Could not connect to Redis")
		return nil
	}

	events.Em = events.NewEmitter(
		db.Events,
		getEmitterConfig(deployment),
		deployment,
	)

	app.Get("/ping", pingHandler)

	app.Get("/version", versionHandler)

	accounts.Routes(app)
	teams.Routes(app)
	superusers.Routes(app)

	return app
}

// pingHandler answers with a plain "PONG" for service uptime checks.
// @Summary Health check
// @Description Lightweight heartbeat used by load balancers to confirm the OpenHack API is alive.
// @Tags General
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /ping [get]
func pingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}

// versionHandler prints the current deployment version for observability.
// @Summary Current deployment version
// @Description Exposes the semantic version bundled with the running process for smoke tests.
// @Tags General
// @Produce plain
// @Success 200 {string} string "25.09.29.0"
// @Router /version [get]
func versionHandler(c fiber.Ctx) error {
	return c.SendString(env.VERSION)
}
