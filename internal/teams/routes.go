package teams

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(app *fiber.App) {
	teams := app.Group("/teams")

	teams.Get("/ping", teamPingHandler)

	// team operations
	teams.Get("", models.AccountMiddleware, TeamGetHandler)
	teams.Post("", models.AccountMiddleware, TeamCreateHandler)
	teams.Patch("", models.AccountMiddleware, TeamChangeHandler)
	teams.Delete("", models.AccountMiddleware, TeamDeleteHandler)

	// teammate operations
	teams.Get("/members", models.AccountMiddleware, TeamGetTeammatesHandler)
	teams.Patch("/join", models.AccountMiddleware, TeamJoinHandler)
	teams.Patch("/leave", models.AccountMiddleware, TeamLeaveHandler)
	teams.Patch("/kick", models.AccountMiddleware, TeamKickHandler)

	// submission operations
	teams.Patch("/submissions/name", models.AccountMiddleware, TeamSubmissionChangeNameHandler)
	teams.Patch("/submissions/desc", models.AccountMiddleware, TeamSubmissionChangeDescHandler)
	teams.Patch("/submissions/repo", models.AccountMiddleware, TeamSubmissionChangeRepoHandler)
	teams.Patch("/submissions/pres", models.AccountMiddleware, TeamSubmissionChangePresHandler)

}

// teamPingHandler ensures the teams subsystem responds to health probes.
// @Summary Teams service health check
// @Description Returns a PONG from the teams group so orchestration checks can verify connectivity.
// @Tags Teams Health
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /teams/ping [get]
func teamPingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}
