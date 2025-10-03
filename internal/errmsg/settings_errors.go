package errmsg

import "net/http"

var (
	SettingIncomplete = NewStatusError(
		http.StatusBadRequest,
		"setting data is incomplete",
	)
	SettingNotFound = NewStatusError(
		http.StatusNotFound,
		"setting not found",
	)
)

type _SettingIncomplete struct {
	StatusCode int    `json:"statusCode" example:"400"`
	Message    string `json:"message" example:"setting data is incomplete"`
}

type _SettingNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"setting not found"`
}
