package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// checkinScanParticipantHandler handles badge scans for participants.
// @Summary Scan a participant badge
// @Description Looks up the account by ID, emits telemetry, and returns the account plus computed badge pile.
// @Tags Superusers Check-In
// @Security SuperUserAuth
// @Produce json
// @Param id query string true "Account ID"
// @Success 200 {object} CheckInScanResponse
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 404 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /superusers/checkin/scan [post]
func checkinScanParticipantHandler(c fiber.Ctx) error {
	var su models.SuperUser
	utils.GetLocals(c, "superuser", &su)

	account := models.Account{ID: c.Query("id")}
	err := account.Get()
	if err != nil {
		return utils.StatusError(c,
			errmsg.AccountNotFound,
		)
	}

	events.Em.CheckInScan(su.Username, account.ID)

	return c.JSON(bson.M{
		"account": account,
		"pile":    utils.PileForAccount(account.ID, utils.BadgePileSalt()),
	})
}
