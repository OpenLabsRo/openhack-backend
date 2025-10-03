package models

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/utils"
	"encoding/json"
	"strings"
	"time"

	sj "github.com/brianvoe/sjwt"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type Consumables struct {
	Water      int  `json:"water" bson:"water" example:"0"`
	Pizza      bool `json:"pizza" bson:"pizza" example:"false"`
	Coffee     bool `json:"coffee" bson:"coffee" example:"false"`
	Jerky      bool `json:"jerky" bson:"jerky" example:"false"`
	Sandwiches int  `json:"sandwiches" bson:"sandwiches" example:"0"`
}

type Account struct {
	ID string `json:"id" bson:"id"`

	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`

	Name string `json:"name" bson:"name"`

	// extra information about the user
	MedicalConditions string      `json:"medicalConditions" bson:"medicalConditions"`
	FoodRestrictions  string      `json:"foodRestrictions" bson:"foodRestrictions"`
	University        string      `json:"university" bson:"university"`
	DOB               string      `json:"dob" bson:"dob"`
	PhoneNumber       string      `json:"phoneNumber" bson:"phoneNumber"`
	CheckedIn         bool        `json:"checkedIn" bson:"checkedIn"`
	Consumables       Consumables `json:"consumables" bson:"consumables"`

	Present bool `json:"present" bson:"present"`

	TeamID string `json:"teamID" bson:"teamID"`
}

func (acc Account) GenToken() string {
	claims, _ := sj.ToClaims(acc)
	claims.SetExpiresAt(time.Now().Add(365 * 24 * time.Hour))

	token := claims.Generate(env.JWT_SECRET)
	return token
}

func (acc *Account) ParseToken(token string) error {
	hasVerified := sj.Verify(token, env.JWT_SECRET)

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

	if string(authHeader) != "" &&
		strings.HasPrefix(string(authHeader), "Bearer") {

		tokens := strings.Fields(string(authHeader))
		if len(tokens) == 2 {
			token = tokens[1]
		}

		if token == "" {
			return utils.StatusError(c,
				errmsg.AccountNoToken,
			)
		}

		var account Account
		err := account.ParseToken(token)
		if err != nil {
			return utils.StatusError(
				c, errmsg.AccountNoToken,
			)
		}

		if account.ID == "" {
			return utils.StatusError(
				c, errmsg.AccountNoToken,
			)
		}

		c.Locals("id", account.ID)
		utils.SetLocals(c, "account", account)
	}

	if token == "" {
		return utils.StatusError(c,
			errmsg.AccountNoToken,
		)
	}

	return c.Next()
}

func (acc *Account) Initialize() (serr errmsg.StatusError) {
	_ = acc.GetByEmail(acc.Email)

	if acc.ID != "" {
		return errmsg.AccountAlreadyInitialized
	}

	acc.ID = utils.GenID(6)

	_, err := db.Accounts.InsertOne(db.Ctx, acc)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	cacheAccount(acc)

	return
}

func (acc *Account) Delete() (err error) {
	var cachedEmail string
	if acc.Email != "" {
		cachedEmail = acc.Email
	} else {
		if data, cacheErr := db.CacheGetBytes(accountCacheKey(acc.ID)); cacheErr != nil {
			var cached Account
			if jsonErr := json.Unmarshal(data, &cached); jsonErr == nil {
				cachedEmail = cached.Email
			}
		}
	}

	_, err = db.Accounts.DeleteOne(db.Ctx, bson.M{
		"id": acc.ID,
	})

	if err == nil {
		invalidateAccountCache(acc.ID, cachedEmail)
	}

	return
}

func (acc *Account) Get() (err error) {
	if loadAccountFromCache(accountCacheKey(acc.ID), acc) {
		return nil
	}

	err = db.Accounts.FindOne(db.Ctx, bson.M{
		"id": acc.ID,
	}).Decode(&acc)

	if err == nil {
		cacheAccount(acc)
	}

	return
}

func (acc *Account) GetByEmail(email string) (serr errmsg.StatusError) {
	if loadAccountFromCache(accountEmailCacheKey(email), acc) {
		return
	}

	err := db.Accounts.FindOne(db.Ctx, bson.M{
		"email": email,
	}).Decode(&acc)

	if err != nil {
		return errmsg.AccountNotInitialized
	}

	if acc.ID == "" {
		return errmsg.AccountNotInitialized
	}

	cacheAccount(acc)

	return
}

func (acc *Account) ExistsAndHasPassword() (exists bool, has bool) {
	if err := acc.Get(); err != nil {
		return false, false
	}

	exists = acc.Name != ""
	has = acc.Password != ""

	return exists, has
}

func (acc *Account) CreatePassword(password string) (serr errmsg.StatusError) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	_, err := db.Accounts.UpdateOne(db.Ctx, bson.M{
		"id": acc.ID,
	}, bson.M{
		"$set": bson.M{
			"password": string(hashedPassword),
		},
	})

	if err != nil {
		return errmsg.InternalServerError(err)
	}

	acc.Password = string(hashedPassword)

	cacheAccount(acc)

	return
}

func (acc *Account) EditName(name string) (err error) {

	_, err = db.Accounts.UpdateOne(db.Ctx, bson.M{
		"id": acc.ID,
	}, bson.M{
		"$set": bson.M{
			"name": name},
	})

	if err != nil {
		return
	}

	acc.Name = name

	cacheAccount(acc)

	return
}

func (acc *Account) UpdateConsumables(consumables Consumables) (err error) {
	_, err = db.Accounts.UpdateOne(db.Ctx, bson.M{
		"id": acc.ID,
	}, bson.M{
		"$set": bson.M{
			"consumables": consumables,
		},
	})

	if err != nil {
		return
	}

	acc.Consumables = consumables

	cacheAccount(acc)

	return
}

func (acc *Account) UpdatePresent(pres bool) (err error) {
	_, err = db.Accounts.UpdateOne(db.Ctx, bson.M{
		"id": acc.ID,
	}, bson.M{
		"$set": bson.M{
			"present": pres,
		},
	})

	if err != nil {
		return
	}

	acc.Present = pres

	cacheAccount(acc)

	return
}

func (acc *Account) AddToTeam(teamID string) (err error) {
	_, err = db.Accounts.UpdateOne(db.Ctx, bson.M{
		"id": acc.ID,
	}, bson.M{
		"$set": bson.M{
			"teamID": teamID,
		},
	})

	if err != nil {
		return
	}

	acc.TeamID = teamID

	cacheAccount(acc)

	return
}

func (acc *Account) RemoveFromTeam(teamID string) (err error) {
	_, err = db.Accounts.UpdateOne(db.Ctx, bson.M{
		"id": acc.ID,
	}, bson.M{
		"$set": bson.M{
			"teamID": "",
		},
	})

	if err != nil {
		return
	}

	acc.TeamID = ""

	cacheAccount(acc)

	return
}

func cacheAccount(acc *Account) {
	if acc == nil || acc.ID == "" {
		return
	}

	bytes, err := json.Marshal(acc)
	if err != nil {
		return
	}

	_ = db.CacheSetBytes(accountCacheKey(acc.ID), bytes)

	if acc.Email != "" {
		_ = db.CacheSetBytes(accountEmailCacheKey(acc.Email), bytes)
	}
}

func loadAccountFromCache(key string, acc *Account) bool {
	if key == "" {
		return false
	}

	bytes, err := db.CacheGetBytes(key)
	if err != nil || len(bytes) == 0 {
		return false
	}

	if jsonErr := json.Unmarshal(bytes, acc); jsonErr != nil {
		_ = db.CacheDel(key)
		return false
	}

	return acc.ID != ""
}

func invalidateAccountCache(id string, email string) {
	if id != "" {
		_ = db.CacheDel(accountCacheKey(id))
	}
	if email != "" {
		_ = db.CacheDel(accountEmailCacheKey(email))
	}
}

func accountCacheKey(id string) string {
	if id == "" {
		return ""
	}
	return "account:" + id
}

func accountEmailCacheKey(email string) string {
	if email == "" {
		return ""
	}
	return "account:email:" + email
}
