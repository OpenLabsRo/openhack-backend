package errmsg

import "net/http"

var (
	FlagRequired = NewStatusError(
		http.StatusUnauthorized,
		"this feature is not available right now",
	)
)

type _FlagRequired struct {
	StatusCode int    `json:"statusCode" example:"401"`
	Message    string `json:"message" example:"this feature is not available right now"`
}
