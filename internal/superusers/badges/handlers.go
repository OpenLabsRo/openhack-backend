package badges

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// pilesGetHandler computes badge pile assignments for all accounts.
// @Summary Retrieve badge pile assignments for all accounts
// @Description Hashes each account into deterministic piles so on-site staff can stage badge pickup.
// @Tags Superusers Badges
// @Security SuperUserAuth
// @Produce json
// @Success 200 {object} PilesResponse
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/badges [get]
func pilesGetHandler(c fiber.Ctx) error {
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

	salt, serr := loadBadgePileSalt()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	sieve := make([][]models.Account, env.BADGE_PILES)

	for _, v := range accounts {
		pile := utils.PileForAccount(v.ID, salt)
		sieve[pile] = append(sieve[pile], v)
	}

	return c.JSON(sieve)
}

// pilesGetHandler computes badge pile assignments for all accounts.
// @Summary Retrieve badge pile assignments for all accounts
// @Description Hashes each account into deterministic piles so on-site staff can stage badge pickup.
// @Tags Superusers Badges
// @Security SuperUserAuth
// @Produce json
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/badges [get]
func pilesComputeHandler(c fiber.Ctx) error {
	var body struct {
		Trials int `json:"trials"`
	}

	if len(c.Body()) > 0 {
		_ = json.Unmarshal(c.Body(), &body)
	}

	var accounts []models.Account
	cursor, err := db.Accounts.Find(db.Ctx, bson.M{})
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	if err = cursor.All(db.Ctx, &accounts); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	salt, counts, serr := computeAndPersistBadgePileSalt(accounts, body.Trials)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	return c.JSON(bson.M{
		"salt":   salt,
		"counts": counts,
	})
}
