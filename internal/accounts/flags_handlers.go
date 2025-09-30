package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

// GetFlagsHandler returns feature flag assignments for the current account.
// @Summary Retrieve current feature flags
// @Description Provides the participant's active stage and boolean flags to drive feature toggles in the client.
// @Tags Accounts Flags
// @Security AccountAuth
// @Produce json
// @Success 200 {object} models.Flags
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /accounts/flags [get]
func GetFlagsHandler(c fiber.Ctx) error {
	flags := models.Flags{}
	err := flags.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	return c.JSON(flags)
}
