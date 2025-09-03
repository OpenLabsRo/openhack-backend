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

func AccountRegister(c fiber.Ctx) error {
	id := c.Query("id")

	var body struct {
		Password string `json:"password" bson:"password"`
	}
	json.Unmarshal(c.Body(), &body)

	account := models.Account{
		ID: id,
	}

	serr := account.CreatePassword(body.Password)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}

func AccountLogin(c fiber.Ctx) error {
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
