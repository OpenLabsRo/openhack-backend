package judging

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {
	r.Post("/init",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		judgeInitHandler,
	)

	r.Post("/compute-rankings",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		models.FlagsMiddlewareBuilder([]string{"judging"}),
		computeRankingsHandler,
	)

	r.Get("/judges",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		getAllJudgesHandler,
	)

	r.Get("/finalists",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		getFinalistsHandler,
	)

	r.Get("/voting-results",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		getVotingResultsHandler,
	)

	judge := r.Group("/judge")

	judge.Post("",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		judgeCreateHandler,
	)

	judge.Post("/connect",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		judgeConnectHandler,
	)

	judge.Delete("",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		deleteJudgeHandler,
	)
}
