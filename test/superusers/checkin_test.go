package superusers

import (
	"backend/internal"
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/models"
	"backend/internal/utils"
	"backend/test/helpers"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	checkinApp               *fiber.App
	checkinTestSuperUserToken string
	checkinTestStaffUser      models.SuperUser
	checkinTestStaffUserToken string
	checkinTestAccounts       []models.Account
	numCheckinTestAccounts    = 60
)

func TestMain(m *testing.M) {
	// Setup
	checkinApp = internal.SetupApp("test")

	// login superuser
	bodyBytes, _ := helpers.API_SuperUsersLogin(
		&testing.T{},
		checkinApp,
		env.SUPERUSER_USERNAME,
		env.SUPERUSER_PASSWORD,
	)

	var body struct {
		Token string `json:"token"`
	}

	json.Unmarshal(bodyBytes, &body)
	checkinTestSuperUserToken = body.Token

	// create staff user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("staffpassword"), 12)
	staffUser := models.SuperUser{
		Username:    "staffuser",
		Password:    string(hashedPassword),
		Permissions: []string{"staff.checkin"},
	}
	db.SuperUsers.InsertOne(context.Background(), staffUser)
	checkinTestStaffUser = staffUser

	bodyBytes, _ = helpers.API_SuperUsersLogin(
		&testing.T{},
		checkinApp,
		staffUser.Username,
		"staffpassword",
	)

	json.Unmarshal(bodyBytes, &body)
	checkinTestStaffUserToken = body.Token

	// create accounts
	for i := 0; i < numCheckinTestAccounts; i++ {
		bodyBytes, _ = helpers.API_SuperUsersAccountsInitialize(
			&testing.T{},
			checkinApp,
			fmt.Sprintf("checkin_test_%v@test.com", i),
			fmt.Sprintf("Checkin Test %v", i),
			checkinTestSuperUserToken,
		)

		var acc models.Account
		json.Unmarshal(bodyBytes, &acc)
		checkinTestAccounts = append(checkinTestAccounts, acc)
	}

	// Run tests
	code := m.Run()

	// Teardown
	for _, acc := range checkinTestAccounts {
		acc.Delete()
	}
	db.SuperUsers.DeleteOne(context.Background(), bson.M{"username": checkinTestStaffUser.Username})
	db.Tags.DeleteMany(context.Background(), bson.M{})

	os.Exit(code)
}

func TestCheckinProcess(t *testing.T) {
	// get all accounts
	var accounts []models.Account
	cursor, _ := db.Accounts.Find(db.Ctx, bson.M{"id": bson.M{"$in": getCheckinTestAccountIDs()}})
	cursor.All(db.Ctx, &accounts)

	var accountIDs []string
	for _, acc := range accounts {
		accountIDs = append(accountIDs, acc.ID)
	}

	// find best salt
	salt, _ := utils.ChooseBestSalt(accountIDs, env.BADGE_PILES, 1000)

	// set salt
	env.BADGE_PILES_SALT = strconv.FormatUint(uint64(salt), 10)

	// get badges
	bodyBytes, statusCode := helpers.API_SuperUsersStaffCheckinBadgesGet(t, checkinApp, checkinTestSuperUserToken)
	require.Equal(t, http.StatusOK, statusCode)

	var piles [][]models.Account
	json.Unmarshal(bodyBytes, &piles)

	// check that all accounts are in the piles
	count := 0
	for _, pile := range piles {
		count += len(pile)
	}
	require.Equal(t, len(accounts), count)

	// check-in each account
	for _, acc := range accounts {
		// scan account
		bodyBytes, statusCode := helpers.API_SuperUsersStaffCheckinScan(t, checkinApp, acc.ID, checkinTestStaffUserToken)
		require.Equal(t, http.StatusOK, statusCode)

		var body struct {
			Account models.Account `json:"account"`
		}
		json.Unmarshal(bodyBytes, &body)
		require.Equal(t, acc.ID, body.Account.ID)

		// assign tag
		tagID := fmt.Sprintf("tag_%v", acc.ID)
		_, statusCode = helpers.API_SuperUsersStaffCheckinTagsAssign(t, checkinApp, tagID, acc.ID, checkinTestStaffUserToken)
		require.Equal(t, http.StatusOK, statusCode)
	}

	// verify tags
	var tags []models.Tag
	cursor, _ = db.Tags.Find(db.Ctx, bson.M{})
	cursor.All(db.Ctx, &tags)

	require.Len(t, tags, numCheckinTestAccounts)
}

func getCheckinTestAccountIDs() []string {
	var ids []string
	for _, acc := range checkinTestAccounts {
		ids = append(ids, acc.ID)
	}
	return ids
}
