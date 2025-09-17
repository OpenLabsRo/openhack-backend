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

	superusers.Get("/whoami",
		models.SuperUserMiddlewareBuilder([]string{"admin"}),
		func(c fiber.Ctx) error {
			su := models.SuperUser{}
			utils.GetLocals(c, "superuser", &su)

			return c.JSON(bson.M{
				"superuser": su,
			})
		},
	)

	// login for supersusers
	superusers.Post("/login", loginHandler)

	// initializing accounts
	superusers.Post("/accounts/initialize",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		initializeAccountHandler,
	)

	// flags
	superusers.Get("/flags",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsGetHandler,
	)
	superusers.Post("/flags",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsSetHandler,
	)
	superusers.Put("/flags",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsSetBulkHandler,
	)
	superusers.Put("/flags/reset",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsResetHandler,
	)
	superusers.Delete("/flags",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagsUnsetHandler,
	)

	// testing the flags middleware
	superusers.Get("/flags/test",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		models.FlagsMiddlewareBuilder([]string{
			"test", "testing",
		}),
		flagsTestHandler,
	)

	// flagstages
	superusers.Get("/flags/stages",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagStagesGetHandler,
	)
	superusers.Post("/flags/stages",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagStagesCreateHandler,
	)
	superusers.Delete("/flags/stages",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagStagesDeleteHandler,
	)
	superusers.Post("/flags/stages/execute",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		flagStagesExecuteHandler,
	)

	// checkin
	superusers.Get("/checkin/badges",
		models.SuperUserMiddlewareBuilder([]string{
			"admin",
		}),
		badgePilesGetHandler,
	)
	superusers.Get("/checkin/tags",
		models.SuperUserMiddlewareBuilder([]string{
			"staff", // as it should be a route available to all staff
		}),
		tagsGetHandler,
	)
	superusers.Post("/checkin/tags",
		models.SuperUserMiddlewareBuilder([]string{
			"staff.checkin", // as it should only be available for people at checking (and admin)
		}),
		tagsAssignHandler,
	)
	superusers.Post("/checkin/scan",
		models.SuperUserMiddlewareBuilder([]string{
			"staff.checkin",
		}),
		checkinScanParticipantHandler,
	)
}
