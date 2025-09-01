package accounts

import (
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func AccountEdit(c fiber.Ctx) error {

	var body struct {
		Name string `json:"name" bson:"name"`
	}
	json.Unmarshal(c.Body(), &body)

	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	account.EditName(body.Name)

	token := account.GenToken()

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
	})
}
