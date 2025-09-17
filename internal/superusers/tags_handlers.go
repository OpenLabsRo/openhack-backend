package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func tagsGetHandler(c fiber.Ctx) error {
	id := c.Query("id")

	tag := models.Tag{ID: id}
	serr := tag.Get()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	return c.JSON(tag)
}

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
