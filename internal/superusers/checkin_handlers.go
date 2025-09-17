package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

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
		"pile":    utils.PileForAccount(account.ID, 1),
	})
}
