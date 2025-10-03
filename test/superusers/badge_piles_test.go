package superusers

import (
	"backend/internal"
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/models"
	"backend/test/helpers"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	badgeApp *fiber.App

	badgeTestSuperUserToken string
	badgeOriginalEnvSalt    string
	badgeOriginalSetting    models.Setting
	badgeOriginalSaltExists bool

	badgeCreatedAccounts []models.Account
)

func init() {
	badgeApp = internal.SetupApp("test", "", "")
	helpers.ResetTestCache()
	clearBadgeEvents()
}

func clearBadgeEvents() {
	if db.Events == nil {
		return
	}
	_, _ = db.Events.DeleteMany(db.Ctx, bson.M{})
}

func TestBadgePilesInitialSalt(t *testing.T) {
	badgeOriginalEnvSalt = env.BADGE_PILES_SALT

	bodyBytes, statusCode := helpers.API_SuperUsersAuthLogin(
		t,
		badgeApp,
		env.SUPERUSER_USERNAME,
		env.SUPERUSER_PASSWORD,
	)
	require.Equal(t, http.StatusOK, statusCode)

	var loginResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &loginResp))
	badgeTestSuperUserToken = loginResp.Token

	t.Logf("Environment badge pile salt on startup: %q", badgeOriginalEnvSalt)
}

func TestBadgePilesCreateAccounts(t *testing.T) {
	require.NotEmpty(t, badgeTestSuperUserToken, "superuser token should be initialized")

	const totalNewAccounts = 60
	baseSuffix := time.Now().UnixNano()

	for i := 0; i < totalNewAccounts; i++ {
		email := fmt.Sprintf("badge_pile_test_%d_%d@test.com", baseSuffix, i)
		name := fmt.Sprintf("Badge Pile Test %d", i)

		bodyBytes, statusCode := helpers.API_SuperUsersParticipantsInitialize(
			t,
			badgeApp,
			email,
			name,
			badgeTestSuperUserToken,
		)
		require.Equal(t, http.StatusOK, statusCode)

		var acc models.Account
		require.NoError(t, json.Unmarshal(bodyBytes, &acc))
		badgeCreatedAccounts = append(badgeCreatedAccounts, acc)
	}

	t.Logf("Created %d accounts for badge pile evaluation", len(badgeCreatedAccounts))
}

func TestBadgePilesComputeSalt(t *testing.T) {
	require.NotEmpty(t, badgeTestSuperUserToken, "superuser token should be initialized")
	require.NotEmpty(t, badgeCreatedAccounts, "badge accounts should be seeded")

	trials := 10_000
	start := time.Now()
	bodyBytes, statusCode := helpers.API_SuperUsersBadgesCompute(
		t,
		badgeApp,
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

	t.Logf("Computed badge pile salt: %d", computeResp.Salt)
	t.Logf("Reported pile counts from compute endpoint: %v", computeResp.Counts)
	t.Logf("Environment badge pile salt after compute: %q", env.BADGE_PILES_SALT)
	t.Logf("Computing badge piles took: %s", duration)
}

func TestBadgePilesFetchPiles(t *testing.T) {
	require.NotEmpty(t, badgeTestSuperUserToken, "superuser token should be initialized")

	bodyBytes, statusCode := helpers.API_SuperUsersBadgesGet(
		t,
		badgeApp,
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

	for idx, pile := range piles {
		totalCounts[idx] = len(pile)
		for _, acc := range pile {
			if _, ok := newAccountSet[acc.ID]; ok {
				newAccountCounts[idx]++
			}
		}
	}

	t.Logf("Total accounts per pile: %v", totalCounts)
	t.Logf("Newly created accounts per pile: %v", newAccountCounts)
	t.Logf("Environment badge pile salt during fetch: %q", env.BADGE_PILES_SALT)
}

func TestBadgePilesCleanup(t *testing.T) {
	for _, acc := range badgeCreatedAccounts {
		_, _ = db.Accounts.DeleteOne(db.Ctx, bson.M{"id": acc.ID})
	}
	badgeCreatedAccounts = nil

	if badgeOriginalSaltExists {
		t.Logf("Original badge pile salt before test: %s", badgeOriginalSetting.Value)
	} else {
		t.Log("No badge pile salt existed prior to test; keeping newly generated value")
	}

	t.Logf("Badge pile salt retained in settings; environment currently: %q", env.BADGE_PILES_SALT)
}
