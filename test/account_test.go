package test

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AccountCleanup() {
	err := accounts[0].Delete()
	if err != nil {
		fmt.Printf("failed to delete account: %v", err)
	}

	accounts = []models.Account{}
	passwords = []string{}
	tokens = []string{}
}

func TestAccountPing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/accounts/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "PONG")
}

func TestAccountInitialize(t *testing.T) {
	// request payload
	payload := map[string]string{
		"email": "testaccount1@openhack.ro",
		"name":  "Test Initialize",
	}

	// marshal to JSON
	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/accounts/initialize", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// send request to the shared app
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// status code
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// decode response
	var acc models.Account
	err = json.NewDecoder(resp.Body).Decode(&acc)
	require.NoError(t, err)

	// assertions
	require.NotEmpty(t, acc.ID, "expected ID to be set")
	require.Equal(t, payload["email"], acc.Email, "email should match")
	require.Equal(t, payload["name"], acc.Name, "name should match")

	accounts = append(accounts, acc)
	marshaled, err := json.Marshal(accounts[0])

	fmt.Printf("Account: %s", string(marshaled))
}

func TestAccountInitializeDuplicateEmail(t *testing.T) {
	payload := map[string]string{
		"email": accounts[0].Email,
		"name":  "Duplicate User",
	}

	// marshal to JSON
	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/accounts/initialize", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// send request to the shared app
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody struct {
		Message string `json:"message"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	require.Equal(t, errmsg.AccountInitializeAlreadyExists, respBody.Message, "message should match")
}

func TestAccountRegister(t *testing.T) {
	// request payload
	payload := map[string]string{
		"password": "testingpassword",
	}
	passwords = append(passwords, payload["password"])

	// marshal to JSON
	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("/accounts/register?id=%v", accounts[0].ID),
		bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// send request to the shared app
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// status code
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// decode response
	var respBody struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	// assertions
	require.NotEmpty(t, respBody.Account.ID, "expected ID to be set")
	require.NotEmpty(t, respBody.Token, "expected token to be set")

	tokens = append(tokens, respBody.Token)
	fmt.Println("Token:", respBody.Token)
}

func TestAccountRegisterNotExist(t *testing.T) {
	// request payload
	payload := map[string]string{
		"password": "testingpassword",
	}
	passwords = append(passwords, payload["password"])

	// marshal to JSON
	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST",
		"/accounts/register?id=123",
		bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// send request to the shared app
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// status code
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// decode response
	var respBody struct {
		Error string `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	// assertions
	require.Equal(t, errmsg.AccountRegisterNotExist, respBody.Error)
}

func TestAccountLogin(t *testing.T) {
	// request payload
	payload := map[string]string{
		"email":    accounts[0].Email,
		"password": passwords[0],
	}

	// marshal to JSON
	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST",
		"/accounts/login",
		bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// send request to the shared app
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// status code
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// decode response
	var respBody struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	// assertions
	require.NotEmpty(t, respBody.Account.ID, "expected ID to be set")
	require.NotEmpty(t, respBody.Token, "expected token to be set")

	tokens[0] = respBody.Token
	fmt.Println("Token:", tokens[0])

	accounts[0] = respBody.Account
	marshaled, err := json.Marshal(accounts[0])
	fmt.Printf("Account: %s \n", string(marshaled))
}

func TestAccountEdit(t *testing.T) {
	// request payload
	payload := struct {
		Name string `json:"name"`
	}{
		Name: "Updated Name",
	}

	// marshal to JSON
	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest("PATCH",
		"/accounts",
		bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", tokens[0]))

	// send request to the shared app
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// status code
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// decode response
	var respBody struct {
		Account models.Account `json:"account"`
		Token   string         `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	// assertions
	require.NotEmpty(t, respBody.Account.ID, "expected ID to be set")
	require.NotEmpty(t, respBody.Token, "expected token to be set")
	require.Equal(t, respBody.Account.Name, payload.Name, "expected name to be equal to payload")

	// updating the account
	tokens[0] = respBody.Token
	fmt.Println("Token:", tokens[0])

	accounts[0] = respBody.Account
	marshaled, err := json.Marshal(accounts[0])
	fmt.Printf("Account: %s \n", string(marshaled))

	t.Cleanup(func() {
		AccountCleanup()
	})
}
