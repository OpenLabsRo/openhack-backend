package internal

import (
	"backend/internal/accounts"
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/events"
	"backend/internal/meta"
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

	meta.Routes(app.Group("/meta"))
	superusers.Routes(app.Group("/superusers"))
	accounts.Routes(app.Group("/accounts"))
	teams.Routes(app.Group("/teams"))

	return app
}
