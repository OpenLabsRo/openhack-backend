package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

var PORT string
var PREFORK bool
var VERSION string

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	PORT = os.Getenv("PORT")
	PREFORK = os.Getenv("PREFORK") == "true"
	VERSION = "2025.08.08.1"

	app := fiber.New()

	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	app.Get("/version", func(c fiber.Ctx) error {
		return c.SendString(VERSION)
	})

	if app.Listen(fmt.Sprintf(":%v", PORT), fiber.ListenConfig{
		EnablePrefork: PREFORK,
	}) != nil {
		log.Fatal("Error listening on port %v", PORT)
	}
}
