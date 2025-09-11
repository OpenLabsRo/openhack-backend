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
	app := internal.SetupApp(deployment)

	app.Get("/testflags", models.FlagsMiddlewareBuilder([]string{"test"}))

	if app.Listen(fmt.Sprintf(":%v", env.PORT), fiber.ListenConfig{
		EnablePrefork: env.PREFORK,
	}) != nil {
		log.Fatalf("Error listening on port %v", env.PORT)
	}
}
