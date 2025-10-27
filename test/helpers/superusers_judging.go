package helpers

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_SuperUsersJudgingInit(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"POST",
		"/superusers/judging/init",
		[]byte{},
		&token,
	)
}

func API_SuperUsersJudgingCreate(
	t *testing.T,
	app *fiber.App,
	judgeID string,
	judgeName string,
	token string,
) (bodyBytes []byte, statusCode int) {
	payload := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{
		ID:   judgeID,
		Name: judgeName,
	}

	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/superusers/judging/judge",
		sendBytes,
		&token,
	)
}

func API_SuperUsersJudgingConnect(
	t *testing.T,
	app *fiber.App,
	judgeID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	payload := struct {
		ID string `json:"id"`
	}{
		ID: judgeID,
	}

	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/superusers/judging/judge/connect",
		sendBytes,
		&token,
	)
}

func API_JudgeUpgrade(
	t *testing.T,
	app *fiber.App,
	connectToken string,
) (bodyBytes []byte, statusCode int) {
	payload := struct {
		Token string `json:"token"`
	}{
		Token: connectToken,
	}

	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"POST",
		"/judge/upgrade",
		sendBytes,
		nil,
	)
}

func API_JudgeNextTeam(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"POST",
		"/judge/next-team",
		[]byte{},
		&token,
	)
}

func API_JudgeTeamInfo(
	t *testing.T,
	app *fiber.App,
	teamID string,
	token string,
) (bodyBytes []byte, statusCode int) {
	return RequestRunner(t, app,
		"GET",
		"/judge/team?id="+teamID,
		[]byte{},
		&token,
	)
}
