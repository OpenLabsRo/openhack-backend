package teams

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {
	r.Get("/meta/ping", teamPingHandler)
	r.Get("/meta/preview", TeamPreviewHandler)

	// team operations
	r.Get("", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read"}), TeamGetHandler)
	r.Post("", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "teams_write"}), TeamCreateHandler)
	r.Patch("/name", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "teams_write"}), TeamChangeNameHandler)
	r.Patch("/table", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "teams_write"}), TeamChangeTableHandler)
	r.Delete("", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "teams_write"}), TeamDeleteHandler)

	// members operations
	r.Get("/members", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read"}), TeamMembersGetHandler)
	r.Patch("/members/join", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "teams_write"}), TeamMembersJoinHandler)
	r.Patch("/members/leave", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "teams_write"}), TeamMembersLeaveHandler)
	r.Patch("/members/kick", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "teams_write"}), TeamMembersKickHandler)

	// submission operations
	r.Get("/submissions", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "submissions_read"}), TeamSubmissionGetHandler)
	r.Patch("/submissions/name", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "submissions_write"}), TeamSubmissionChangeNameHandler)
	r.Patch("/submissions/desc", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "submissions_write"}), TeamSubmissionChangeDescHandler)
	r.Patch("/submissions/repo", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "submissions_write"}), TeamSubmissionChangeRepoHandler)
	r.Patch("/submissions/pres", models.AccountMiddleware, models.FlagsMiddlewareBuilder([]string{"teams_read", "submissions_write"}), TeamSubmissionChangePresHandler)
}
