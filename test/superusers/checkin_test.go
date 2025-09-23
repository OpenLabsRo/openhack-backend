package superusers

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/models"
	"backend/internal/utils"
	"backend/test/helpers"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	checkinSetupOnce      sync.Once
	checkinCleanupOnce    sync.Once
	checkinSetupErr       error
	checkinSuperUserToken string
	checkinStaffUser      models.SuperUser
	checkinStaffToken     string
	checkinAccounts       []models.Account
	checkinAccountCount   = 60
)

func ensureCheckinFixtures(t *testing.T) {
	t.Helper()

	checkinSetupOnce.Do(func() {
		if err := bootstrapCheckinFixtures(t); err != nil {
			checkinSetupErr = err
			return
		}
	})

	if checkinSetupErr != nil {
		t.Fatalf("check-in fixture setup failed: %v", checkinSetupErr)
	}

	t.Cleanup(func() {
		checkinCleanupOnce.Do(func() {
			teardownCheckinFixtures()
		})
	})
}

func bootstrapCheckinFixtures(t *testing.T) error {
	// Authenticate as the default superuser.
	bodyBytes, status := helpers.API_SuperUsersLogin(t, app, env.SUPERUSER_USERNAME, env.SUPERUSER_PASSWORD)
	if status != http.StatusOK {
		return fmt.Errorf("superuser login failed: status %d", status)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		return fmt.Errorf("decode superuser login: %w", err)
	}
	checkinSuperUserToken = loginResp.Token

	// Create a staff user that can perform check-ins.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("staffpassword"), 12)
	if err != nil {
		return fmt.Errorf("hash staff password: %w", err)
	}

	staffUser := models.SuperUser{
		Username:    "staffuser",
		Password:    string(hashedPassword),
		Permissions: []string{"staff.checkin"},
	}

	if _, err := db.SuperUsers.InsertOne(context.Background(), staffUser); err != nil {
		return fmt.Errorf("insert staff user: %w", err)
	}
	checkinStaffUser = staffUser

	// Obtain staff token.
	bodyBytes, status = helpers.API_SuperUsersLogin(t, app, staffUser.Username, "staffpassword")
	if status != http.StatusOK {
		return fmt.Errorf("staff login failed: status %d", status)
	}

	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		return fmt.Errorf("decode staff login: %w", err)
	}
	checkinStaffToken = loginResp.Token

	// Seed a batch of attendee accounts that can be checked in.
	for i := 0; i < checkinAccountCount; i++ {
		email := fmt.Sprintf("checkin_test_%d@test.com", i)
		display := fmt.Sprintf("Checkin Test %d", i)

		bodyBytes, status = helpers.API_SuperUsersAccountsInitialize(t, app, email, display, checkinSuperUserToken)
		if status != http.StatusOK {
			return fmt.Errorf("initialize account %s: status %d", email, status)
		}

		var acc models.Account
		if err := json.Unmarshal(bodyBytes, &acc); err != nil {
			return fmt.Errorf("decode account %s: %w", email, err)
		}
		checkinAccounts = append(checkinAccounts, acc)
	}

	return nil
}

func teardownCheckinFixtures() {
	for _, acc := range checkinAccounts {
		_, _ = db.Accounts.DeleteOne(context.Background(), bson.M{"id": acc.ID})
	}

	if checkinStaffUser.Username != "" {
		_, _ = db.SuperUsers.DeleteOne(context.Background(), bson.M{"username": checkinStaffUser.Username})
	}

	_, _ = db.Tags.DeleteMany(context.Background(), bson.M{})

	checkinAccounts = nil
	checkinStaffUser = models.SuperUser{}
	checkinSuperUserToken = ""
	checkinStaffToken = ""
}

func TestCheckinProcess(t *testing.T) {
	ensureCheckinFixtures(t)

	cursor, err := db.Accounts.Find(db.Ctx, bson.M{"id": bson.M{"$in": getCheckinAccountIDs()}})
	require.NoError(t, err)

	var accounts []models.Account
	require.NoError(t, cursor.All(db.Ctx, &accounts))

	salt, _ := utils.ChooseBestSalt(collectAccountIDs(accounts), env.BADGE_PILES, 1000)
	env.BADGE_PILES_SALT = strconv.FormatUint(uint64(salt), 10)

	bodyBytes, statusCode := helpers.API_SuperUsersStaffCheckinBadgesGet(t, app, checkinSuperUserToken)
	require.Equal(t, http.StatusOK, statusCode)

	var piles [][]models.Account
	require.NoError(t, json.Unmarshal(bodyBytes, &piles))

	total := 0
	for _, pile := range piles {
		total += len(pile)
	}
	require.Equal(t, len(accounts), total)

	for _, acc := range accounts {
		// Scan account badge
		bodyBytes, statusCode = helpers.API_SuperUsersStaffCheckinScan(t, app, acc.ID, checkinStaffToken)
		require.Equal(t, http.StatusOK, statusCode)

		var scanResp struct {
			Account models.Account `json:"account"`
		}
		require.NoError(t, json.Unmarshal(bodyBytes, &scanResp))
		require.Equal(t, acc.ID, scanResp.Account.ID)

		// Assign physical tag
		tagID := fmt.Sprintf("tag_%s", acc.ID)
		_, statusCode = helpers.API_SuperUsersStaffCheckinTagsAssign(t, app, tagID, acc.ID, checkinStaffToken)
		require.Equal(t, http.StatusOK, statusCode)
	}

	cursor, err = db.Tags.Find(db.Ctx, bson.M{})
	require.NoError(t, err)

	var tags []models.Tag
	require.NoError(t, cursor.All(db.Ctx, &tags))
	require.Len(t, tags, checkinAccountCount)
}

func getCheckinAccountIDs() []string {
	ids := make([]string, 0, len(checkinAccounts))
	for _, acc := range checkinAccounts {
		ids = append(ids, acc.ID)
	}
	return ids
}

func collectAccountIDs(accounts []models.Account) []string {
	ids := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		ids = append(ids, acc.ID)
	}
	return ids
}
