package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// AccountEditHandler updates the display name for the authenticated participant.
// @Summary Update display name
// @Description Stores new first and last names and returns a refreshed token so clients keep their session in sync.
// @Tags Accounts Profile
// @Security AccountAuth
// @Accept json
// @Produce json
// @Param payload body AccountEditRequest true "First and last name"
// @Success 200 {object} AccountTokenResponse
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /accounts/me [patch]
func AccountEditHandler(c fiber.Ctx) error {
	var body struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	json.Unmarshal(c.Body(), &body)

	account := models.Account{}
	utils.GetLocals(c, "account", &account)
	oldFirstName := account.FirstName
	oldLastName := account.LastName

	err := account.EditName(body.FirstName, body.LastName)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	token := account.GenToken()

	events.Em.AccountNameChanged(
		account.ID,
		oldFirstName+" "+oldLastName,
		account.FirstName+" "+account.LastName,
	)

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}
