package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

var (
	_ = errmsg.StatusError{}
)

func Routes(app *fiber.App) {
	superusers := app.Group("/superusers")

	superusers.Get("/ping", superUserPingHandler)

	superusers.Get("/whoami",
		models.SuperUserMiddlewareBuilder([]string{"admin"}),
		superUserWhoAmIHandler,
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
	superusers.Get("/staff/tags",
		models.SuperUserMiddlewareBuilder([]string{
			"staff", // as it should be a route available to all staff
		}),
		staffTagGetHandler,
	)
	superusers.Post("/staff/tags",
		models.SuperUserMiddlewareBuilder([]string{
			"staff", // as it should only be available for people at checking (and admin)
		}),
		staffTagPostHandler,
	)
	superusers.Post("/staff/register",
		models.SuperUserMiddlewareBuilder([]string{
			"staff",
		}),
		staffRegisterHandler,
	)
	superusers.Put("/staff/consumables",
		models.SuperUserMiddlewareBuilder([]string{
			"staff",
		}),
		staffConsumablesPutHandler,
	)
}

// superUserPingHandler responds to health probes for the superuser subsystem.
// @Summary Superuser service health check
// @Description Confirms the privileged routes segment is reachable by returning a simple PONG.
// @Tags Superusers Health
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /superusers/ping [get]
func superUserPingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}

// superUserWhoAmIHandler reveals the authenticated superuser context.
// @Summary Inspect the current superuser context
// @Description Echoes the active superuser payload so operators can verify their scopes.
// @Tags Superusers Health
// @Security SuperUserAuth
// @Produce json
// @Success 200 {object} models.SuperUser
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Router /superusers/whoami [get]
func superUserWhoAmIHandler(c fiber.Ctx) error {
	su := models.SuperUser{}
	utils.GetLocals(c, "superuser", &su)

	return c.JSON(su)
}
