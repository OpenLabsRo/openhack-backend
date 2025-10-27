package participants

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// initializeHandler seeds an account shell for a participant.
// @Summary Initialize an account shell for a participant
// @Description Creates a passwordless account record so the participant can complete onboarding later.
// @Tags Superusers Participants
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body InitializeRequest true "Participant seed data"
// @Success 200 {object} models.Account
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 409 {object} errmsg._AccountAlreadyInitialized
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/participants [post]
func initializeHandler(c fiber.Ctx) error {
	superuser := models.SuperUser{}
	utils.GetLocals(c, "superuser", &superuser)

	account := models.Account{}
	json.Unmarshal(c.Body(), &account)

	serr := account.Initialize()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	// Set default flags for the account
	account.Flags = []string{"teams_read", "teams_write", "submissions_read", "submissions_write"}
	_, err := db.Accounts.UpdateOne(db.Ctx, bson.M{"id": account.ID}, bson.M{"$set": bson.M{"flags": account.Flags}})
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	events.Em.AccountInitialized(
		superuser.Username,
		account.ID,
	)

	return c.JSON(account)
}
