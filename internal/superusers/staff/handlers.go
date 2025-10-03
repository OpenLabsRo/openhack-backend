package staff

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// checkinScanParticipantHandler handles badge scans for participants.
// @Summary Scan a participant QR code
// @Description Looks up the account by ID, emits telemetry, and returns the account plus computed badge pile. This route also modifies the account to be checked-in.
// @Tags Superusers Staff
// @Security SuperUserAuth
// @Produce json
// @Param id query string true "Account ID"
// @Success 200 {object} StaffRegisterResponse
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 404 {object} errmsg._AccountNotFound
// @Router /superusers/staff/register [post]
func staffRegisterHandler(c fiber.Ctx) error {
	var su models.SuperUser
	utils.GetLocals(c, "superuser", &su)

	account := models.Account{ID: c.Query("id")}
	err := account.Get()
	if err != nil {
		return utils.StatusError(c,
			errmsg.AccountNotFound,
		)
	}

	events.Em.StaffRegister(su.Username, account.ID)

	return c.JSON(bson.M{
		"account": account,
		"pile":    utils.PileForAccount(account.ID, utils.BadgePileSalt()),
	})
}

// staffAccountGetHandler gets an account based on the accountID
// @Summary Gets an account based on the accountID
// @Description Looks up the account by ID and returns the account.
// @Tags Superusers Staff
// @Security SuperUserAuth
// @Produce json
// @Param accountID query string true "Account ID"
// @Success 200 {object} models.Account
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 404 {object} errmsg._AccountNotFound
// @Router /superusers/staff/account [get]
func staffAccountGetHandler(c fiber.Ctx) error {
	account := models.Account{ID: c.Query("accountID")}
	err := account.Get()
	if err != nil {
		return utils.StatusError(c,
			errmsg.AccountNotFound,
		)
	}

	return c.JSON(account)
}

// tagsGetHandler fetches a stored account by the linked tag ID.
// @Summary Fetch a tag's linked account
// @Description Pulls tag details from cache or Mongo and then fetches the linked account.
// @Tags Superusers Staff
// @Security SuperUserAuth
// @Produce json
// @Param id query string true "Tag ID"
// @Success 200 {object} models.Account
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 404 {object} errmsg._TagNotFound
// @Failure 404 {object} errmsg._AccountNotFound
// @Failure 409 {object} errmsg._TagIncomplete
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/staff/tags [get]
func staffTagGetHandler(c fiber.Ctx) error {
	id := c.Query("id")

	tag := models.Tag{ID: id}
	serr := tag.Get()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	account := models.Account{ID: tag.AccountID}
	err := account.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.AccountNotFound)
	}

	return c.JSON(account)
}

// tagsAssignHandler links a tag to an account.
// @Summary Assign a tag to an account
// @Description Persists the assignment and emits an event for check-in telemetry.
// @Tags Superusers Staff
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body models.Tag true "Tag assignment"
// @Success 200 {object} models.Tag
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 409 {object} errmsg._TagIncomplete
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/staff/tags [post]
func staffTagPostHandler(c fiber.Ctx) error {
	var su models.SuperUser
	utils.GetLocals(c, "superuser", &su)

	var tag models.Tag
	json.Unmarshal(c.Body(), &tag)

	serr := tag.Assign()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	events.Em.StaffTagAssign(su.Username, tag.AccountID, tag.ID)

	return c.JSON(tag)
}

// staffConsumablesPutHandler updates the consumables for an account.
// @Summary Update consumables for an account
// @Description Updates the consumables for an account and emits an event for telemetry.
// @Tags Superusers Staff
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body models.Consumables true "Consumables update"
// @Param accountID query string true "Account ID"
// @Success 200 {string} string "OK"
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 404 {object} errmsg._AccountNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/staff/consumables [put]
func staffConsumablesPutHandler(c fiber.Ctx) error {
	var consumables models.Consumables
	json.Unmarshal(c.Body(), &consumables)

	var su models.SuperUser
	utils.GetLocals(c, "superuser", &su)

	account := models.Account{ID: c.Query("accountID")}
	err := account.Get()
	if err != nil {
		return utils.StatusError(c,
			errmsg.AccountNotFound,
		)
	}
	err = account.UpdateConsumables(consumables)
	if err != nil {
		return utils.StatusError(c,
			errmsg.InternalServerError(err),
		)
	}

	events.Em.StaffConsumablesUpdated(su.Username, account.ID, consumables)

	return c.SendString("OK")
}

// staffPresentIn flags a participant as present.
// @Summary flags a participant as present
// @Description Creates a present in event, and changes the property of an account to present.
// @Tags Superusers Staff
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param accountID query string true "Account ID"
// @Success 200 {string} string "OK"
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 404 {object} errmsg._AccountNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/staff/in [patch]
func staffPresentIn(c fiber.Ctx) error {
	var su models.SuperUser
	utils.GetLocals(c, "superuser", &su)

	account := models.Account{ID: c.Query("accountID")}
	err := account.Get()
	if err != nil {
		return utils.StatusError(c,
			errmsg.AccountNotFound,
		)
	}
	err = account.UpdatePresent(true)
	if err != nil {
		return utils.StatusError(c,
			errmsg.InternalServerError(err),
		)
	}

	events.Em.StaffPresentIn(su.Username, account.ID)

	return c.SendString("OK")
}

// staffPresentOut flags a participant as not present.
// @Summary flags a participant as not present
// @Description Creates a present out event, and changes the property of an account to not present.
// @Tags Superusers Staff
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param accountID query string true "Account ID"
// @Success 200 {string} string "OK"
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 404 {object} errmsg._AccountNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/staff/out [patch]
func staffPresentOut(c fiber.Ctx) error {
	var su models.SuperUser
	utils.GetLocals(c, "superuser", &su)

	account := models.Account{ID: c.Query("accountID")}
	err := account.Get()
	if err != nil {
		return utils.StatusError(c,
			errmsg.AccountNotFound,
		)
	}
	err = account.UpdatePresent(false)
	if err != nil {
		return utils.StatusError(c,
			errmsg.InternalServerError(err),
		)
	}

	events.Em.StaffPresentOut(su.Username, account.ID)

	return c.SendString("OK")
}
