package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func initializeAccountHandler(c fiber.Ctx) error {
	var account models.Account
	json.Unmarshal(c.Body(), &account)

	serr := account.Initialize()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	return c.JSON(account)
}
