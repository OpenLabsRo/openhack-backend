package errmsg

import (
	"net/http"
)

var (
	AccountNotInitialized = NewStatusError(
		http.StatusNotFound,
		"account not initialized - talk to the administrator",
	)

	AccountAlreadyRegistered = NewStatusError(
		http.StatusConflict,
		"account already registered",
	)

	AccountLoginWrongPassword = NewStatusError(
		http.StatusUnauthorized,
		"wrong password",
	)

	AccountNoToken = NewStatusError(
		http.StatusUnauthorized,
		"you are not logged in",
	)

	AccountNotFound = NewStatusError(
		http.StatusNotFound,
		"account not found",
	)
)

type _AccountNotInitialized struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"account not initialized - talk to the administrator"`
}

type _AccountAlreadyRegistered struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"account already registered"`
}

type _AccountLoginWrongPassword struct {
	StatusCode int    `json:"statusCode" example:"401"`
	Message    string `json:"message" example:"wrong password"`
}

type _AccountNoToken struct {
	StatusCode int    `json:"statusCode" example:"401"`
	Message    string `json:"message" example:"you are not logged in"`
}

type _AccountNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"account not found"`
}
