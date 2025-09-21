package main

import (
	"backend/internal"
	"backend/internal/env"
	"backend/internal/models"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: server <deployment-type>")
		os.Exit(1)
	}
	deployment := os.Args[1]

	port := ""
	switch deployment {
	case "test":
		port = "9000"
	case "dev":
		port = "9001"
	case "prod":
		port = "9002"
	default:
		log.Fatalf("Invalid deployment type: %s", deployment)
	}

	app := internal.SetupApp(deployment)

	app.Get("/testflags", models.FlagsMiddlewareBuilder([]string{"test"}))

	if app.Listen(fmt.Sprintf(":%s", port), fiber.ListenConfig{
		EnablePrefork: env.PREFORK,
	}) != nil {
		log.Fatalf("Error listening on port %s", port)
	}
}
