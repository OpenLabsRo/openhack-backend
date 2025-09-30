package errmsg

import "net/http"

var (
	AccountAlreadyInitialized = NewStatusError(
		http.StatusConflict,
		"account already initialized",
	)
	SuperUserNotExists = NewStatusError(
		http.StatusNotFound,
		"superuser does not exist",
	)
	SuperUserNoToken = NewStatusError(
		http.StatusUnauthorized,
		"no token has been provided",
	)
	TagIncomplete = NewStatusError(
		http.StatusConflict,
		"tag data is incomplete",
	)
	TagNotFound = NewStatusError(
		http.StatusNotFound,
		"tag not found",
	)
	PileNotFound = NewStatusError(
		http.StatusNotFound,
		"badge pile not found",
	)
)

type _AccountAlreadyInitialized struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"account already initialized"`
}

type _SuperUserNotExists struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"superuser does not exist"`
}

type _SuperUserNoToken struct {
	StatusCode int    `json:"statusCode" example:"401"`
	Message    string `json:"message" example:"no token has been provided"`
}

type _TagIncomplete struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"tag data is incomplete"`
}

type _TagNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"tag not found"`
}

type _PileNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"badge pile not found"`
}
