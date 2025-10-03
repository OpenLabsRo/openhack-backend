package badges

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {
	r.Get("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		pilesGetHandler,
	)
	r.Post("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		pilesComputeHandler,
	)
}
