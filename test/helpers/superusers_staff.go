package helpers

import (
	"backend/internal/models"
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_SuperUsersStaffCheckinTagsGet(
	t *testing.T,
	app *fiber.App,
	ID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/superusers/checkin/tags?id="+ID,
		[]byte{},
		&token,
	)
}

func API_SuperUsersStaffCheckinTagsAssign(
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
		"/superusers/checkin/tags",
		sendBytes,
		&token,
	)
}

func API_SuperUsersStaffCheckinScan(
	t *testing.T,
	app *fiber.App,
	accountID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/superusers/checkin/scan?id="+accountID,
		[]byte{},
		&token,
	)
}
