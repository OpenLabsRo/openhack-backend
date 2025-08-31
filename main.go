package main

import (
	"backend/accounts"
	"backend/db"
	"backend/env"
	"backend/teams"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	if db.InitDB() != nil {
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

	if app.Listen(fmt.Sprintf(":%v", env.PORT), fiber.ListenConfig{
		EnablePrefork: env.PREFORK,
	}) != nil {
		log.Fatalf("Error listening on port %v", env.PORT)
	}
}
