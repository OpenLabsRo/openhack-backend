package main

import (
	"backend/env"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := SetupApp("dev")

	if app.Listen(fmt.Sprintf(":%v", env.PORT), fiber.ListenConfig{
		EnablePrefork: env.PREFORK,
	}) != nil {
		log.Fatalf("Error listening on port %v", env.PORT)
	}
}
