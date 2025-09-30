package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

// tagsGetHandler fetches a stored check-in tag by ID.
// @Summary Fetch a check-in tag by ID
// @Description Pulls tag details from cache or Mongo so staff can verify issued tags.
// @Tags Superusers Check-In
// @Security SuperUserAuth
// @Produce json
// @Param id query string true "Tag ID"
// @Success 200 {object} models.Tag
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 404 {object} swagger.StatusErrorDoc
// @Failure 409 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /superusers/checkin/tags [get]
func tagsGetHandler(c fiber.Ctx) error {
	id := c.Query("id")

	tag := models.Tag{ID: id}
	serr := tag.Get()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	return c.JSON(tag)
}

// tagsAssignHandler links a tag to an account.
// @Summary Assign a tag to an account
// @Description Persists the assignment and emits an event for check-in telemetry.
// @Tags Superusers Check-In
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body TagAssignRequest true "Tag assignment"
// @Success 200 {object} models.Tag
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 404 {object} swagger.StatusErrorDoc
// @Failure 409 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /superusers/checkin/tags [post]
func tagsAssignHandler(c fiber.Ctx) error {
	var su models.SuperUser
	utils.GetLocals(c, "superuser", &su)

	var tag models.Tag
	json.Unmarshal(c.Body(), &tag)

	serr := tag.Assign()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	events.Em.CheckInTagAssign(su.Username, tag.AccountID, tag.ID)

	return c.JSON(tag)
}
