package participants

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {
	r.Post("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		initializeHandler,
	)
}
