package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func loginHandler(c fiber.Ctx) error {
	var body struct {
		Username string `json:"username" bson:"username"`
		Password string `json:"password" bson:"password"`
	}
	json.Unmarshal(c.Body(), &body)

	su := models.SuperUser{}
	serr := su.Get(body.Username)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(su.Password),
		[]byte(body.Password),
	) != nil {
		return utils.StatusError(c,
			errmsg.AccountLoginWrongPassword,
		)
	}

	token := su.GenToken()

	return c.JSON(bson.M{
		"token":     token,
		"superuser": su,
	})
}
