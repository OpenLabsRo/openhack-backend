package superusers

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/models"
	"backend/test/helpers"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	checkinSetupOnce      sync.Once
	checkinCleanupOnce    sync.Once
	checkinSetupErr       error
	checkinSuperUserToken string
	checkinStaffToken     string
	checkinAccounts       []models.Account
	checkinAccountCount   = 1
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
	bodyBytes, status := helpers.API_SuperUsersAuthLogin(t, app, env.SUPERUSER_USERNAME, env.SUPERUSER_PASSWORD)
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

	if env.STAFF_SUPERUSER_USERNAME == "" || env.STAFF_SUPERUSER_PASSWORD == "" {
		return fmt.Errorf("staff superuser credentials not configured")
	}

	bodyBytes, status = helpers.API_SuperUsersAuthLogin(t, app, env.STAFF_SUPERUSER_USERNAME, env.STAFF_SUPERUSER_PASSWORD)
	if status != http.StatusOK {
		return fmt.Errorf("staff login failed: status %d", status)
	}

	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		return fmt.Errorf("decode staff login: %w", err)
	}
	checkinStaffToken = loginResp.Token

	for i := 0; i < checkinAccountCount; i++ {
		email := fmt.Sprintf("checkin_test_%d@test.com", i)
		firstName := "Checkin Test"
		lastName := fmt.Sprintf("%d", i)

		bodyBytes, status = helpers.API_SuperUsersParticipantsInitialize(t, app, email, firstName, lastName, checkinSuperUserToken)
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

	_, _ = db.Tags.DeleteMany(context.Background(), bson.M{})

	checkinAccounts = nil
	checkinSuperUserToken = ""
	checkinStaffToken = ""
}

func TestCheckinProcess(t *testing.T) {
	ensureCheckinFixtures(t)

	cursor, err := db.Accounts.Find(db.Ctx, bson.M{"id": bson.M{"$in": getCheckinAccountIDs()}})
	require.NoError(t, err)

	var accounts []models.Account
	require.NoError(t, cursor.All(db.Ctx, &accounts))

	var (
		bodyBytes  []byte
		statusCode int
	)

	for i, acc := range accounts {
		// Scan account badge
		bodyBytes, statusCode = helpers.API_SuperUsersStaffRegister(t, app, acc.ID, checkinStaffToken)
		require.Equal(t, http.StatusOK, statusCode)

		var scanResp struct {
			Account models.Account `json:"account"`
		}
		require.NoError(t, json.Unmarshal(bodyBytes, &scanResp))
		require.Equal(t, acc.ID, scanResp.Account.ID)

		// Assign physical tag
		tagID := fmt.Sprintf("tag_%s", acc.ID)
		_, statusCode = helpers.API_SuperUsersStaffTagsAssign(t, app, tagID, acc.ID, checkinStaffToken)
		require.Equal(t, http.StatusOK, statusCode)

		bodyBytes, statusCode = helpers.API_SuperUsersStaffTagsGet(t, app, tagID, checkinStaffToken)
		require.Equal(t, http.StatusOK, statusCode)

		var taggedAccount models.Account
		require.NoError(t, json.Unmarshal(bodyBytes, &taggedAccount))
		require.Equal(t, acc.ID, taggedAccount.ID)

		if i == 0 {
			consumables := models.Consumables{Water: 3, Pizza: true, Coffee: true}
			_, statusCode = helpers.API_SuperUsersStaffConsumablesUpdate(t, app, acc.ID, consumables, checkinStaffToken)
			require.Equal(t, http.StatusOK, statusCode)

			_, statusCode = helpers.API_SuperUsersStaffPresenceIn(t, app, acc.ID, checkinStaffToken)
			require.Equal(t, http.StatusOK, statusCode)

			bodyBytes, statusCode = helpers.API_SuperUsersStaffAccountGet(t, app, acc.ID, checkinStaffToken)
			require.Equal(t, http.StatusOK, statusCode)

			var staffAccount models.Account
			require.NoError(t, json.Unmarshal(bodyBytes, &staffAccount))
			require.True(t, staffAccount.Present)
			require.Equal(t, consumables.Water, staffAccount.Consumables.Water)

			_, statusCode = helpers.API_SuperUsersStaffPresenceOut(t, app, acc.ID, checkinStaffToken)
			require.Equal(t, http.StatusOK, statusCode)
		}
	}

	bodyBytes, statusCode = helpers.API_SuperUsersStaffAccountGet(t, app, accounts[0].ID, checkinStaffToken)
	require.Equal(t, http.StatusOK, statusCode)

	var finalAccount models.Account
	require.NoError(t, json.Unmarshal(bodyBytes, &finalAccount))
	require.False(t, finalAccount.Present)

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
