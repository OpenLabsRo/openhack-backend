package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func AccountInitialize(c fiber.Ctx) error {
	var account models.Account
	json.Unmarshal(c.Body(), &account)

	err := account.Initialize()
	if err != nil {
		return utils.Error(c, http.StatusConflict, err)
	}

	return c.JSON(account)
}

func AccountRegister(c fiber.Ctx) error {
	id := c.Query("id")

	var body struct {
		Password string `json:"password" bson:"password"`
	}
	json.Unmarshal(c.Body(), &body)

	account := models.Account{
		ID: id,
	}

	err := account.CreatePassword(body.Password)
	if err != nil {
		return utils.Error(c, http.StatusNotFound, errors.New(errmsg.AccountRegisterNotExist))
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
	account.GetByEmail(body.Email)

	if bcrypt.CompareHashAndPassword(
		[]byte(account.Password),
		[]byte(body.Password),
	) != nil {
		return utils.Error(c, http.StatusUnauthorized, errors.New(errmsg.AccountLoginWrongPassword))
	}

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}
