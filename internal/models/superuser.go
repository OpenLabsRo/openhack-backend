package models

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/utils"
	"strings"
	"time"

	sj "github.com/brianvoe/sjwt"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

type SuperUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (su SuperUser) GenToken() string {
	claims, _ := sj.ToClaims(su)
	claims.SetExpiresAt(time.Now().Add(365 * 24 * time.Hour))

	token := claims.Generate(env.JWT_KEY)
	return token
}

func (su *SuperUser) ParseToken(token string) error {
	hasVerified := sj.Verify(token, env.JWT_KEY)

	if !hasVerified {
		return nil
	}

	claims, _ := sj.Parse(token)
	err := claims.Validate()
	claims.ToStruct(&su)

	return err
}

func SuperUserMiddleware(c fiber.Ctx) error {
	var token string

	authHeader := c.Get("Authorization")

	if string(authHeader) != "" && strings.HasPrefix(string(authHeader), "Bearer") {

		tokens := strings.Fields(string(authHeader))
		if len(tokens) == 2 {
			token = tokens[1]
		}
		if token == "" {
			return utils.StatusError(c,
				errmsg.AccountNoToken,
			)
		}

		var su SuperUser
		err := su.ParseToken(token)
		if err != nil {

		}

		utils.SetLocals(c, "superuser", su)
	}

	if token == "" {
		return utils.StatusError(c,
			errmsg.AccountNoToken,
		)
	}

	return c.Next()
}

func (su *SuperUser) Get(username string) (serr errmsg.StatusError) {
	db.SuperUsers.FindOne(db.Ctx, bson.M{
		"username": username,
	}).Decode(&su)

	if su.Password == "" {
		return errmsg.SuperUserNotExists
	}

	return
}
