package superusers

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/test/helpers"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	badgeAccountsToCreate    = 52
	badgePileSaltSettingName = "badgePileSalt"
)

var (
	badgeTestSuperUserToken string
	badgeOriginalEnvSalt    string
	badgeOriginalSetting    models.Setting
	badgeOriginalSaltExists bool

	badgeCreatedAccounts []models.Account
	badgePileSaltSetting models.Setting
)

func TestBadgePilesInitialSalt(t *testing.T) {
	require.NotNil(t, app, "test app should be initialized in TestMain")

	badgeOriginalEnvSalt = env.BADGE_PILES_SALT

	bodyBytes, statusCode := helpers.API_SuperUsersAuthLogin(
		t,
		app,
		env.SUPERUSER_USERNAME,
		env.SUPERUSER_PASSWORD,
	)
	require.Equal(t, http.StatusOK, statusCode)

	var loginResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &loginResp))
	badgeTestSuperUserToken = loginResp.Token

	// Fetch existing badge pile salt setting from database
	badgePileSaltSetting.Name = badgePileSaltSettingName
	err := badgePileSaltSetting.Get()
	if err == errmsg.EmptyStatusError {
		badgeOriginalSaltExists = true
		badgeOriginalSetting = badgePileSaltSetting
	}

	fmt.Printf("Environment badge pile salt on startup: %q\n", badgeOriginalEnvSalt)
}

func TestBadgePilesCreateAccounts(t *testing.T) {
	require.NotEmpty(t, badgeTestSuperUserToken, "superuser token should be initialized")

	for i := range badgeAccountsToCreate {
		email := fmt.Sprintf("badge_pile_test_%d@test.com", i)
		firstName := "Badge Pile Test"
		lastName := fmt.Sprintf("%d", i)

		bodyBytes, statusCode := helpers.API_SuperUsersParticipantsInitialize(
			t,
			app,
			email,
			firstName,
			lastName,
			badgeTestSuperUserToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var acc models.Account
		require.NoError(t, json.Unmarshal(bodyBytes, &acc))
		badgeCreatedAccounts = append(badgeCreatedAccounts, acc)
	}

	require.Len(t, badgeCreatedAccounts, badgeAccountsToCreate)
	fmt.Printf("Created %d accounts for badge pile evaluation\n", len(badgeCreatedAccounts))
}

func TestBadgePilesComputeSalt(t *testing.T) {
	require.NotEmpty(t, badgeTestSuperUserToken, "superuser token should be initialized")
	require.NotEmpty(t, badgeCreatedAccounts, "badge accounts should be seeded")
	require.Len(t, badgeCreatedAccounts, badgeAccountsToCreate, "should have %d created accounts before computing salt", badgeAccountsToCreate)

	trials := 10_000
	start := time.Now()
	bodyBytes, statusCode := helpers.API_SuperUsersBadgesCompute(
		t,
		app,
		&trials,
		badgeTestSuperUserToken,
	)
	duration := time.Since(start)
	require.Equal(t, http.StatusOK, statusCode)

	var computeResp struct {
		Salt   uint32 `json:"salt"`
		Counts []int  `json:"counts"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &computeResp))

	fmt.Printf("Computed badge pile salt: %d\n", computeResp.Salt)
	fmt.Printf("Reported pile counts from compute endpoint: %v\n", computeResp.Counts)
	fmt.Printf("Environment badge pile salt after compute: %q\n", env.BADGE_PILES_SALT)
	fmt.Printf("Computing badge piles took: %s\n", duration)
}

func TestBadgePilesFetchPiles(t *testing.T) {
	require.NotEmpty(t, badgeTestSuperUserToken, "superuser token should be initialized")

	bodyBytes, statusCode := helpers.API_SuperUsersBadgesGet(
		t,
		app,
		badgeTestSuperUserToken,
	)
	require.Equal(t, http.StatusOK, statusCode)

	var piles [][]models.Account
	require.NoError(t, json.Unmarshal(bodyBytes, &piles))

	totalCounts := make([]int, len(piles))
	newAccountSet := make(map[string]struct{}, len(badgeCreatedAccounts))
	for _, acc := range badgeCreatedAccounts {
		newAccountSet[acc.ID] = struct{}{}
	}

	newAccountCounts := make([]int, len(piles))
	totalAccountsInPiles := 0

	for idx, pile := range piles {
		totalCounts[idx] = len(pile)
		totalAccountsInPiles += len(pile)
		for _, acc := range pile {
			if _, ok := newAccountSet[acc.ID]; ok {
				newAccountCounts[idx]++
			}
		}
	}

	require.Equal(t, badgeAccountsToCreate, len(badgeCreatedAccounts), "created accounts should equal constant")

	totalNewAccountsInPiles := 0
	for _, count := range newAccountCounts {
		totalNewAccountsInPiles += count
	}
	require.Equal(t, badgeAccountsToCreate, totalNewAccountsInPiles, "sum of newly created accounts across piles should equal total created")

	fmt.Printf("Total accounts per pile: %v\n", totalCounts)
	fmt.Printf("Newly created accounts per pile: %v\n", newAccountCounts)
	fmt.Printf("Total newly created accounts in piles: %d (expected: %d)\n", totalNewAccountsInPiles, badgeAccountsToCreate)
	fmt.Printf("Environment badge pile salt during fetch: %q\n", env.BADGE_PILES_SALT)
}

func TestBadgePilesCleanup(t *testing.T) {
	for _, acc := range badgeCreatedAccounts {
		_, _ = db.Accounts.DeleteOne(db.Ctx, bson.M{"id": acc.ID})
	}
	badgeCreatedAccounts = nil

	if badgeOriginalSaltExists {
		fmt.Printf("Original badge pile salt before test: %s\n", badgeOriginalSetting.Value)
		// Restore original setting
		_ = badgeOriginalSetting.Update()
	} else {
		t.Log("No badge pile salt existed prior to test; keeping newly generated value")
	}

	fmt.Printf("Badge pile salt retained in settings; environment currently: %q\n", env.BADGE_PILES_SALT)
}
