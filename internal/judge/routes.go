package judge

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {
	// judge token upgrade (exchange 2-minute token for 24-hour token)
	r.Post("/upgrade", JudgeUpgradeHandler)

	// judge operations (require judge authentication and judging flag)
	r.Post("/next-team",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		nextTeamHandler,
	)
	r.Get("/team",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		getTeamHandler,
	)
	r.Post("/judgment",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		createJudgmentHandler,
	)
}
