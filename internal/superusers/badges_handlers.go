package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func badgeGetHandler(c fiber.Ctx) error {
	id := c.Query("id")

	badge := models.Badge{ID: id}
	serr := badge.Get()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	return c.JSON(badge)
}

func badgeAssignHandler(c fiber.Ctx) error {
	var badge models.Badge
	json.Unmarshal(c.Body(), &badge)

	serr := badge.Assign()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	return c.JSON(badge)
}
