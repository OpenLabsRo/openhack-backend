package superusers

import (
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func Routes(app *fiber.App) {
	superusers := app.Group("/superusers")

	superusers.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("PONG")
	})

	superusers.Get("/whoami", models.SuperUserMiddleware, func(c fiber.Ctx) error {
		su := models.SuperUser{}
		utils.GetLocals(c, "superuser", &su)

		return c.JSON(bson.M{
			"superuser": su,
		})
	})

	// login for supersusers
	superusers.Post("/login", loginHandler)

	// initializing accounts
	superusers.Post("/accounts/initialize", models.SuperUserMiddleware, initializeAccountHandler)

	// flags
	superusers.Get("/flags", models.SuperUserMiddleware, flagsGetHandler)
	superusers.Post("/flags", models.SuperUserMiddleware, flagsSetHandler)
	superusers.Put("/flags", models.SuperUserMiddleware, flagsSetBulkHandler)
	superusers.Put("/flags/reset", models.SuperUserMiddleware, flagsResetHandler)
	superusers.Delete("/flags", models.SuperUserMiddleware, flagsUnsetHandler)

	// testing the flags middleware
	superusers.Get("/flags/test", models.SuperUserMiddleware, models.FlagsMiddlewareBuilder([]string{
		"test", "testing",
	}), flagsTestHandler)

	// flagstages
	superusers.Get("/flags/stages", models.SuperUserMiddleware, flagStagesGetHandler)
	superusers.Post("/flags/stages", models.SuperUserMiddleware, flagStagesCreateHandler)
	superusers.Delete("/flags/stages", models.SuperUserMiddleware, flagStagesDeleteHandler)
	superusers.Post("/flags/stages/execute", models.SuperUserMiddleware, flagStagesExecuteHandler)

}
