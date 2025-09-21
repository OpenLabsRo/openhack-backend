package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

func GetFlagsHandler(c fiber.Ctx) error {
	flags := models.Flags{}
	err := flags.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	return c.JSON(flags)
}
