package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

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
		return utils.StatusError(c, errmsg.AccountAlreadyRegistered)
	}

	serr = account.CreatePassword(body.Password)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

func AccountLoginHandler(c fiber.Ctx) error {
	var body struct {
		Email    string `json:"email" bson:"email"`
		Password string `json:"password" bson:"password"`
	}
	json.Unmarshal(c.Body(), &body)

	account := models.Account{}
	serr := account.GetByEmail(body.Email)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(account.Password),
		[]byte(body.Password),
	) != nil {
		return utils.StatusError(c,
			errmsg.AccountLoginWrongPassword,
		)
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}
