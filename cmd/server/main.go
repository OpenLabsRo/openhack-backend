package main

import (
	"backend/internal"
	"backend/internal/env"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := internal.SetupApp("dev")

	if app.Listen(fmt.Sprintf(":%v", env.PORT), fiber.ListenConfig{
		EnablePrefork: env.PREFORK,
	}) != nil {
		log.Fatalf("Error listening on port %v", env.PORT)
	}
}
