package main

import (
	"backend/accounts"
	"backend/db"
	"backend/env"
	"backend/teams"
	"log"

	"github.com/gofiber/fiber/v3"
)

func SetupApp(deployment string) *fiber.App {
	app := fiber.New()

	if err := db.InitDB(deployment); err != nil {
		log.Fatal("Could not connect to MongoDB")
	}

	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	app.Get("/version", func(c fiber.Ctx) error {
		return c.SendString(env.VERSION)
	})

	accounts.Endpoints(app)
	teams.Endpoints(app)

	return app
}
