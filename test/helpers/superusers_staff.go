package helpers

import (
	"backend/internal/models"
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_SuperUsersBadgesGet(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/superusers/badges",
		[]byte{},
		&token,
	)
}

func API_SuperUsersStaffTagsGet(
	t *testing.T,
	app *fiber.App,
	ID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/superusers/staff/tags?id="+ID,
		[]byte{},
		&token,
	)
}

func API_SuperUsersStaffTagsAssign(
	t *testing.T,
	app *fiber.App,
	tagID string,
	accountID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	payload := models.Tag{
		ID:        tagID,
		AccountID: accountID,
	}

	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/superusers/staff/tags",
		sendBytes,
		&token,
	)
}

func API_SuperUsersStaffRegister(
	t *testing.T,
	app *fiber.App,
	accountID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"POST",
		"/superusers/staff/register?id="+accountID,
		[]byte{},
		&token,
	)
}

func API_SuperUsersStaffAccountGet(
	t *testing.T,
	app *fiber.App,
	accountID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/superusers/staff/account?accountID="+accountID,
		[]byte{},
		&token,
	)
}

func API_SuperUsersStaffConsumablesUpdate(
	t *testing.T,
	app *fiber.App,
	accountID string,
	consumables models.Consumables,
	token string,
) (bodyBytes []byte, statusCode int) {
	sendBytes, err := json.Marshal(consumables)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PUT",
		"/superusers/staff/consumables?accountID="+accountID,
		sendBytes,
		&token,
	)
}

func API_SuperUsersStaffPresenceIn(
	t *testing.T,
	app *fiber.App,
	accountID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"PATCH",
		"/superusers/staff/in?accountID="+accountID,
		[]byte{},
		&token,
	)
}

func API_SuperUsersStaffPresenceOut(
	t *testing.T,
	app *fiber.App,
	accountID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"PATCH",
		"/superusers/staff/out?accountID="+accountID,
		[]byte{},
		&token,
	)
}
