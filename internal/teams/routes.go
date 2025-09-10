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

	teams.Get("", models.AccountMiddleware, TeamGetHandler)
	teams.Post("", models.AccountMiddleware, TeamCreateHandler)
	teams.Patch("", models.AccountMiddleware, TeamChangeHandler)
	teams.Delete("", models.AccountMiddleware, TeamDeleteHandler)

	teams.Get("/members", models.AccountMiddleware, TeamGetTeammatesHandler)
	teams.Patch("/join", models.AccountMiddleware, TeamJoinHandler)
	teams.Patch("/leave", models.AccountMiddleware, TeamLeaveHandler)
	teams.Patch("/kick", models.AccountMiddleware, TeamKickHandler)
}
