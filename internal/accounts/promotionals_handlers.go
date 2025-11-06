package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

// GetPromotionalsHandler retrieves the personalized promotional codes for the authenticated participant.
// @Summary Get promotional codes
// @Description Returns a key-value map of promotional codes assigned to the participant.
// @Tags Accounts Profile
// @Security AccountAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /accounts/promotionals [get]
func GetPromotionalsHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	err := account.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	return c.JSON(account.Promotionals)
}
