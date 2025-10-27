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
}
