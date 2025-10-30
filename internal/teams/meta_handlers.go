package teams

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

// teamPingHandler ensures the teams subsystem responds to health probes.
// @Summary Teams service health check
// @Description Returns a PONG from the teams group so orchestration checks can verify connectivity.
// @Tags Teams Meta
// @Produce plain
// @Success 200 {string} string "PONG"
// @Router /teams/meta/ping [get]
func teamPingHandler(c fiber.Ctx) error {
	return c.SendString("PONG")
}

// TeamPreviewHandler returns preview details about a team by ID.
// @Summary Get team preview details
// @Description Retrieves basic team information including name, members, submission details, and table assignment for a given team ID. Useful for participants reviewing team details before joining via a join link.
// @Tags Teams Meta
// @Produce json
// @Param id query string true "Team ID"
// @Success 200 {object} TeamPreviewResponse
// @Failure 404 {object} errmsg._TeamNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/meta/preview [get]
func TeamPreviewHandler(c fiber.Ctx) error {
	teamID := c.Query("id")
	if teamID == "" {
		return utils.StatusError(
			c, errmsg.TeamNotFound,
		)
	}

	team := models.Team{ID: teamID}
	err := team.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.TeamNotFound,
		)
	}

	members, err := team.GetMembers()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	response := TeamPreviewResponse{
		ID:           team.ID,
		Name:         team.Name,
		Table:        team.Table,
		MembersCount: len(team.Members),
		Members:      members,
		Submission: SubmissionResponse{
			Name: team.Submission.Name,
			Desc: team.Submission.Desc,
			Repo: team.Submission.Repo,
			Pres: team.Submission.Pres,
		},
	}

	return c.JSON(response)
}
