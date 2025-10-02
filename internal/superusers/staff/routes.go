package staff

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {
	r.Get("/tags",
		models.SuperUserMiddlewareBuilder([]string{
			"staff", // as it should be a route available to all staff
		}),
		staffTagGetHandler,
	)
	r.Post("/tags",
		models.SuperUserMiddlewareBuilder([]string{
			"staff", // as it should only be available for people at checking (and admin)
		}),
		staffTagPostHandler,
	)
	r.Post("/register",
		models.SuperUserMiddlewareBuilder([]string{
			"staff",
		}),
		staffRegisterHandler,
	)
	r.Get("/account",
		models.SuperUserMiddlewareBuilder([]string{
			"staff",
		}),
		staffAccountGetHandler,
	)
	r.Put("/consumables",
		models.SuperUserMiddlewareBuilder([]string{
			"staff",
		}),
		staffConsumablesPutHandler,
	)
	r.Patch("/in",
		models.SuperUserMiddlewareBuilder([]string{
			"staff",
		}),
		staffPresentIn,
	)
	r.Patch("/out",
		models.SuperUserMiddlewareBuilder([]string{
			"staff",
		}),
		staffPresentOut,
	)
}
