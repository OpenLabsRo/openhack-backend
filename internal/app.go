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

func SetupApp(deployment string) *fiber.App {
	app := fiber.New()

	env.Init(deployment)

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
		events.Config{
			Buffer:     1000,
			BatchSize:  50,
			FlushEvery: 2 * time.Second,
		},
	)

	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	app.Get("/version", func(c fiber.Ctx) error {
		return c.SendString(env.VERSION)
	})

	accounts.Routes(app)
	teams.Routes(app)
	superusers.Routes(app)

	return app
}
