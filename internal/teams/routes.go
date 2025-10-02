package teams

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func Routes(r fiber.Router) {
	r.Get("/meta/ping", teamPingHandler)

	// team operations
	r.Get("", models.AccountMiddleware, TeamGetHandler)
	r.Post("", models.AccountMiddleware, TeamCreateHandler)
	r.Patch("", models.AccountMiddleware, TeamChangeHandler)
	r.Delete("", models.AccountMiddleware, TeamDeleteHandler)

	// members operations
	r.Get("/members", models.AccountMiddleware, TeamMembersGetHandler)
	r.Patch("/members/join", models.AccountMiddleware, TeamMembersJoinHandler)
	r.Patch("/members/leave", models.AccountMiddleware, TeamMembersLeaveHandler)
	r.Patch("/members/kick", models.AccountMiddleware, TeamMembersKickHandler)

	// submission operations
	r.Patch("/submissions/name", models.AccountMiddleware, TeamSubmissionChangeNameHandler)
	r.Patch("/submissions/desc", models.AccountMiddleware, TeamSubmissionChangeDescHandler)
	r.Patch("/submissions/repo", models.AccountMiddleware, TeamSubmissionChangeRepoHandler)
	r.Patch("/submissions/pres", models.AccountMiddleware, TeamSubmissionChangePresHandler)
}
