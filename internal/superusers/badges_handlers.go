package superusers

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func badgePilesGetHandler(c fiber.Ctx) error {
	// get all Accounts
	//
	var accounts []models.Account
	cursor, err := db.Accounts.Find(db.Ctx, bson.M{})
	if err != nil {
		return utils.StatusError(c,
			errmsg.InternalServerError(err),
		)
	}

	if err = cursor.All(db.Ctx, &accounts); err != nil {
		return utils.StatusError(c,
			errmsg.InternalServerError(err),
		)
	}

	if env.BADGE_PILES <= 0 {
		return utils.StatusError(c,
			errmsg.InternalServerError(fmt.Errorf("badge piles misconfigured")),
		)
	}

	salt := utils.BadgePileSalt()
	sieve := make([][]models.Account, env.BADGE_PILES)

	for _, v := range accounts {
		pile := utils.PileForAccount(v.ID, salt)
		sieve[pile] = append(sieve[pile], v)
	}

	return c.JSON(sieve)
}
