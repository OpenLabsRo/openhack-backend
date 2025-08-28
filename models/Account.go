package models

import (
	"backend/db"
	"backend/env"
	"backend/utils"
	"errors"
	"strings"
	"time"

	sj "github.com/brianvoe/sjwt"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID string `json:"id" bson:"id"`

	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`

	AccountFields `bson:",inline"`
}

type AccountFields struct {
	Name                string   `json:"name" bson:"name"`
	Roles               []string `json:"roles" bson:"roles"`
	Skills              string   `json:"skills" bson:"skills"`
	TshirtSize          string   `json:"tshirtSize" bson:"tshirtSize"`
	DietaryRestrictions string   `json:"dietaryRestrictions" bson:"dietaryRestrictions"`
}

func (acc Account) GenToken() string {
	claims, _ := sj.ToClaims(acc)
	claims.SetExpiresAt(time.Now().Add(365 * 24 * time.Hour))

	token := claims.Generate(env.JWT_KEY)
	return token
}

func (acc *Account) ParseToken(token string) error {
	hasVerified := sj.Verify(token, env.JWT_KEY)

	if !hasVerified {
		return nil
	}

	claims, _ := sj.Parse(token)
	err := claims.Validate()
	claims.ToStruct(&acc)

	return err
}

func AccountMiddleware(c fiber.Ctx) error {
	var token string

	authHeader := c.Get("Authorization")

	if string(authHeader) != "" && strings.HasPrefix(string(authHeader), "Bearer") {

		tokens := strings.Fields(string(authHeader))
		if len(tokens) == 2 {
			token = tokens[1]
		}
		if token == "" {
			return utils.Error(c, errors.New("no token"))
		}

		var account Account
		err := account.ParseToken(token)
		if err != nil {
			return utils.Error(c, errors.New("a aparut o eroare"))
		}

		c.Locals("id", account.ID)
		utils.SetLocals(c, "account", account)
	}

	if token == "" {
		return utils.Error(c, errors.New("no token"))
	}

	return c.Next()
}

func (acc *Account) Initialize() (err error) {
	acc.ID = utils.GenID(6)
	acc.Roles = []string{}

	_, err = db.Accounts.InsertOne(db.Ctx, acc)

	return
}

func (acc *Account) Get(id string) (err error) {
	err = db.Accounts.FindOne(db.Ctx, bson.M{
		"id": id,
	}).Decode(&acc)

	return err
}

func (acc *Account) GetByEmail(email string) (err error) {
	err = db.Accounts.FindOne(db.Ctx, bson.M{
		"email": email,
	}).Decode(&acc)

	return err
}

func (acc *Account) ExistsAndHasPassword() (exists bool, has bool) {
	db.Accounts.FindOne(db.Ctx, bson.M{
		"id": acc.ID,
	}).Decode(&acc)

	exists = false
	if acc.Name != "" {
		exists = true
	}

	has = false
	if acc.Password != "" {
		has = true
	}

	return
}

func (acc *Account) CreatePassword(password string) (err error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	exists, hasPassword := acc.ExistsAndHasPassword()

	if !exists {
		return errors.New("the link is incorrect")
	}

	if hasPassword {
		return errors.New("there is already an account registered")
	}

	_, err = db.Accounts.UpdateOne(db.Ctx, bson.M{
		"id": acc.ID,
	}, bson.M{
		"$set": bson.M{
			"password": string(hashedPassword),
		},
	})

	if err != nil {
		return
	}

	acc.Password = string(hashedPassword)

	return
}

func (acc *Account) EditFields(
	fields AccountFields) (err error) {

	_, err = db.Accounts.UpdateOne(db.Ctx, bson.M{
		"id": acc.ID,
	}, bson.M{
		"$set": bson.M{
			"name":                fields.Name,
			"roles":               fields.Roles,
			"skills":              fields.Skills,
			"tshirtSize":          fields.TshirtSize,
			"dietaryRestrictions": fields.DietaryRestrictions,
		},
	})

	if err != nil {
		return
	}

	acc.Name = fields.Name
	acc.Roles = fields.Roles
	acc.Skills = fields.Skills
	acc.TshirtSize = fields.TshirtSize
	acc.DietaryRestrictions = fields.DietaryRestrictions

	return
}
