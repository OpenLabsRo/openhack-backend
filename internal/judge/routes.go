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
	r.Get("/previous-team",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		previousTeamHandler,
	)
	r.Get("/current-team",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		currentTeamHandler,
	)
	r.Get("/team",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		getTeamHandler,
	)
	r.Get("/all-teams",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		getAllTeamsHandler,
	)

	r.Get("/me",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		judgeInfoHandler,
	)

	r.Post("/judgment",
		models.JudgeMiddleware,
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		createJudgmentHandler,
	)
}
