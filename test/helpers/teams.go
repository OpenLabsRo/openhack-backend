package helpers

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func API_TeamsGet(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {

	return RequestRunner(t, app,
		"GET",
		"/teams",
		[]byte{},
		&token,
	)
}

func API_TeamsCreate(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {

	return RequestRunner(t, app,
		"POST",
		"/teams",
		[]byte{},
		&token,
	)
}

func API_TeamsGetMembers(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {

	return RequestRunner(t, app,
		"GET",
		"/teams/members",
		[]byte{},
		&token,
	)
}

func API_TeamsChange(
	t *testing.T,
	app *fiber.App,
	name string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// building the payload
	payload := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PATCH",
		"/teams",
		sendBytes,
		&token,
	)
}

func API_TeamsDelete(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {

	return RequestRunner(t, app,
		"DELETE",
		"/teams",
		[]byte{},
		&token,
	)
}

func API_TeamsJoin(
	t *testing.T,
	app *fiber.App,
	teamID string,
	token string,
) (bodyBytes []byte, statusCode int) {

	return RequestRunner(t, app,
		"PATCH",
		"/teams/join?id="+teamID,
		[]byte{},
		&token,
	)
}

func API_TeamsLeave(
	t *testing.T,
	app *fiber.App,
	token string,
) (bodyBytes []byte, statusCode int) {

	return RequestRunner(t, app,
		"PATCH",
		"/teams/leave",
		[]byte{},
		&token,
	)
}

func API_TeamsKick(
	t *testing.T,
	app *fiber.App,
	accountID string,
	token string,
) (bodyBytes []byte, statusCode int) {

	return RequestRunner(t, app,
		"PATCH",
		"/teams/kick?id="+accountID,
		[]byte{},
		&token,
	)
}

func API_TeamsSubmissionsChangeName(
	t *testing.T,
	app *fiber.App,
	name string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// building the payload
	payload := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PATCH",
		"/teams/submissions/name",
		sendBytes,
		&token,
	)
}

func API_TeamsSubmissionsChangeDesc(
	t *testing.T,
	app *fiber.App,
	desc string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// building the payload
	payload := struct {
		Desc string `json:"desc"`
	}{
		Desc: desc,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PATCH",
		"/teams/submissions/desc",
		sendBytes,
		&token,
	)
}

func API_TeamsSubmissionsChangeRepo(
	t *testing.T,
	app *fiber.App,
	repo string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// building the payload
	payload := struct {
		Repo string `json:"repo"`
	}{
		Repo: repo,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PATCH",
		"/teams/submissions/repo",
		sendBytes,
		&token,
	)
}
func API_TeamsSubmissionsChangePres(
	t *testing.T,
	app *fiber.App,
	pres string,
	token string,
) (bodyBytes []byte, statusCode int) {
	// building the payload
	payload := struct {
		Pres string `json:"pres"`
	}{
		Pres: pres,
	}

	// marshalling the payload into JSON
	sendBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	return RequestRunner(t, app,
		"PATCH",
		"/teams/submissions/pres",
		sendBytes,
		&token,
	)
}
