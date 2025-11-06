package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

var (
	_ = errmsg.StatusError{}
)

func Routes(r fiber.Router) {
	// accounts := app.Group("/accounts")

	r.Get("/meta/ping", accountPingHandler)

	r.Get("/meta/whoami", models.AccountMiddleware, accountWhoAmIHandler)

	// create
	r.Post("/auth/check", AccountCheckHandler)
	r.Post("/auth/register", AccountRegisterHandler)
	r.Post("/auth/login", AccountLoginHandler)

	// edit
	r.Patch("/me", models.AccountMiddleware, AccountEditHandler)

	// flags
	r.Get("/flags", models.AccountMiddleware, GetFlagsHandler)

	// promotionals
	r.Get("/promotionals", models.AccountMiddleware, GetPromotionalsHandler)

	// vouchers
	r.Get("/vouchers/:voucherType/:index", GetVoucherHandler)

	// voting
	r.Get("/voting/status", models.AccountMiddleware, votingStatusHandler)
	r.Get("/voting/finalists", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"voting"}), votingFinalistsHandler)
	r.Post("/voting/vote", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"voting"}), votingCastVoteHandler)
}
