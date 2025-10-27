package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/superusers/badges"
	"backend/internal/superusers/flags"
	"backend/internal/superusers/flagstages"
	"backend/internal/superusers/judging"
	"backend/internal/superusers/participants"
	"backend/internal/superusers/staff"

	"github.com/gofiber/fiber/v3"
)

var (
	_ = errmsg.StatusError{}
)

func Routes(r fiber.Router) {

	r.Get("/meta/ping", superUserPingHandler)
	r.Get("/meta/whoami",
		models.SuperUserMiddlewareBuilder([]string{"admin"}),
		superUserWhoAmIHandler,
	)

	// login for supersusers
	r.Post("/auth/login", loginHandler)

	flags.Routes(r.Group("/flags"))
	flagstages.Routes(r.Group("/flagstages"))
	badges.Routes(r.Group("/badges"))
	judging.Routes(r.Group("/judging"))
	participants.Routes(r.Group("/participants"))

	staff.Routes(r.Group("/staff"))
}
