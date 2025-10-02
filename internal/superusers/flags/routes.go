package flags

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {
	r.Get("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsGetHandler,
	)
	r.Post("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsSetHandler,
	)
	r.Put("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsSetBulkHandler,
	)
	r.Post("/reset",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsResetHandler,
	)
	r.Delete("/",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsUnsetHandler,
	)

	// testing the flags middleware
	r.Get("/test",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		models.FlagsMiddlewareBuilder([]string{
			"test", "testing",
		}),
		flagsTestHandler,
	)
}
