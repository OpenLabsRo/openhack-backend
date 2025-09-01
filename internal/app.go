package internal

import (
	"backend/internal/accounts"
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/teams"
	"log"

	"github.com/gofiber/fiber/v3"
)

func SetupApp(deployment string) *fiber.App {
	app := fiber.New()

	if err := db.InitDB(deployment); err != nil {
		log.Fatal("Could not connect to MongoDB")
		return nil
	}

	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	app.Get("/version", func(c fiber.Ctx) error {
		return c.SendString(env.VERSION)
	})

	accounts.Routes(app)
	teams.Routes(app)

	return app
}
