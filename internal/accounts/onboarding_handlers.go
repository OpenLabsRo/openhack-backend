package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// AccountCheckHandler verifies whether the participant has already registered.
// @Summary Check registration status
// @Description Confirms if an initialized account already set a password so the UI can branch between login and signup.
// @Tags Accounts Auth
// @Accept json
// @Produce json
// @Param payload body AccountCheckRequest true "Account email"
// @Success 200 {object} AccountCheckResponse
// @Failure 404 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /accounts/check [post]
func AccountCheckHandler(c fiber.Ctx) error {
	var body struct {
		Email string `json:"email" bson:"email"`
	}
	json.Unmarshal(c.Body(), &body)

	// checking if the account is initialized
	account := models.Account{}
	serr := account.GetByEmail(body.Email)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	// if the account already has a password, it's registered
	// if the account doesn't have a password, redirect to the registration page
	return c.JSON(bson.M{
		"registered": account.Password != "",
	})
}

// AccountRegisterHandler finalizes registration by persisting a hashed password.
// @Summary Complete participant registration
// @Description Accepts credentials for an initialized participant and issues a signed session token upon success.
// @Tags Accounts Auth
// @Accept json
// @Produce json
// @Param payload body CredentialRequest true "Account credentials"
// @Success 200 {object} AccountTokenResponse
// @Failure 404 {object} swagger.StatusErrorDoc
// @Failure 409 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /accounts/register [post]
func AccountRegisterHandler(c fiber.Ctx) error {
	var body struct {
		Email    string `json:"email" bson:"email"`
		Password string `json:"password" bson:"password"`
	}
	json.Unmarshal(c.Body(), &body)

	// checking again if the account is initialized
	account := models.Account{}
	serr := account.GetByEmail(body.Email)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	if account.Password != "" {

		events.Em.AccountRegisterFailure(
			account.ID,
			errmsg.AccountAlreadyRegistered.Message,
		)
		return utils.StatusError(c, errmsg.AccountAlreadyRegistered)
	}

	serr = account.CreatePassword(body.Password)
	if serr != errmsg.EmptyStatusError {

		events.Em.AccountRegisterFailure(
			account.ID,
			serr.Message,
		)

		return utils.StatusError(c, serr)
	}

	token := account.GenToken()

	events.Em.AccountRegisterSuccess(
		account.ID,
	)

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

// AccountLoginHandler authenticates a participant and mints a new JWT.
// @Summary Authenticate a participant
// @Description Validates submitted credentials against the stored hash and returns a refreshed token plus account snapshot.
// @Tags Accounts Auth
// @Accept json
// @Produce json
// @Param payload body CredentialRequest true "Login payload"
// @Success 200 {object} AccountTokenResponse
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 404 {object} swagger.StatusErrorDoc
// @Router /accounts/login [post]
func AccountLoginHandler(c fiber.Ctx) error {
	var body struct {
		Email    string `json:"email" bson:"email"`
		Password string `json:"password" bson:"password"`
	}
	json.Unmarshal(c.Body(), &body)

	account := models.Account{}
	serr := account.GetByEmail(body.Email)
	if serr != errmsg.EmptyStatusError {
		events.Em.AccountLoginFailure(
			account.ID,
			serr.Message,
		)
		return utils.StatusError(c, serr)
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(account.Password),
		[]byte(body.Password),
	) != nil {
		events.Em.AccountLoginFailure(
			account.ID,
			errmsg.AccountLoginWrongPassword.Message,
		)
		return utils.StatusError(c,
			errmsg.AccountLoginWrongPassword,
		)
	}

	token := account.GenToken()

	events.Em.AccountLoginSuccess(
		account.ID,
	)

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}
