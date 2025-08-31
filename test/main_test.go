package test

import (
	"backend/internal"
	"backend/internal/models"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

var (
	app       *fiber.App
	accounts  []models.Account
	passwords []string
	tokens    []string
)

func init() {
	app = internal.SetupApp("dev")
	fmt.Println("Setup complete!")
}
