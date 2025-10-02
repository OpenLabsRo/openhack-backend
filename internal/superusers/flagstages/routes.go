package flagstages

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {

	// flagstages
	r.Get("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagStagesGetHandler,
	)
	r.Post("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagStagesCreateHandler,
	)
	r.Delete("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagStagesDeleteHandler,
	)
	r.Post("/execute",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagStagesExecuteHandler,
	)
}
