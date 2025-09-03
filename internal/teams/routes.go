package teams

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(app *fiber.App) {
	teams := app.Group("/teams")

	teams.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	teams.Get("", models.AccountMiddleware, TeamGet)
	teams.Post("", models.AccountMiddleware, TeamCreate)
	teams.Patch("", models.AccountMiddleware, TeamChange)
	teams.Delete("", models.AccountMiddleware, TeamDelete)

	teams.Patch("/join", models.AccountMiddleware, TeamJoin)
	teams.Patch("/leave", models.AccountMiddleware, TeamLeave)
	teams.Patch("/kick", models.AccountMiddleware, TeamKick)
}
