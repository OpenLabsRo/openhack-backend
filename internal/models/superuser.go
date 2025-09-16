package models

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/utils"
	"slices"
	"strings"
	"time"

	sj "github.com/brianvoe/sjwt"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

type SuperUser struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`

	Permissions []string `json:"permissions" bson:"permissions"`
}

func (su SuperUser) GenToken() string {
	claims, _ := sj.ToClaims(su)
	claims.SetExpiresAt(time.Now().Add(365 * 24 * time.Hour))

	token := claims.Generate(env.JWT_SECRET)
	return token
}

func (su *SuperUser) ParseToken(token string) error {
	hasVerified := sj.Verify(token, env.JWT_SECRET)

	if !hasVerified {
		return nil
	}

	claims, _ := sj.Parse(token)
	err := claims.Validate()
	claims.ToStruct(&su)

	return err
}

func (su *SuperUser) HasAllRoles(required []string) bool {
	if su == nil {
		return false
	}

	if slices.Contains(su.Permissions, "admin") {
		return true
	}

	for _, want := range required {
		if !slices.Contains(su.Permissions, want) {
			return false
		}
	}

	return true
}

func SuperUserMiddlewareBuilder(required []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		var token string

		authHeader := c.Get("Authorization")

		if string(authHeader) != "" && strings.HasPrefix(string(authHeader), "Bearer") {

			tokens := strings.Fields(string(authHeader))
			if len(tokens) == 2 {
				token = tokens[1]
			}
			if token == "" {
				return utils.StatusError(c,
					errmsg.SuperUserNoToken,
				)
			}

			var su SuperUser
			err := su.ParseToken(token)
			if err != nil {
				return utils.StatusError(c,
					errmsg.SuperUserNoToken,
				)
			}
			if allowed := su.HasAllRoles(required); !allowed {
				return utils.StatusError(c,
					errmsg.SuperUserNoToken,
				)
			}

			utils.SetLocals(c, "superuser", su)
		}

		if token == "" {
			return utils.StatusError(c,
				errmsg.SuperUserNoToken,
			)
		}

		return c.Next()
	}
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
