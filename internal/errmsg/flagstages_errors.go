package errmsg

import "net/http"

var (
	FlagStageNotFound = NewStatusError(
		http.StatusNotFound,
		"flag stage not found",
	)
)

type _FlagStageNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"flag stage not found"`
}
