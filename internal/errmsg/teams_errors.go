package errmsg

import "net/http"

var (
	TeamNotFound = NewStatusError(
		http.StatusNotFound,
		"team  not found",
	)

	TeamNotEmpty = NewStatusError(
		http.StatusConflict,
		"team is not empty",
	)

	TeamFull = NewStatusError(
		http.StatusConflict,
		"team is full",
	)

	AccountAlreadyHasTeam = NewStatusError(
		http.StatusConflict,
		"account already has a team",
	)

	AccountHasNoTeam = NewStatusError(
		http.StatusConflict,
		"account does not belong to a team",
	)

	TeamSubmissionNotFound = NewStatusError(
		http.StatusNotFound,
		"submission not found",
	)
)

type _TeamNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"team  not found"`
}

type _TeamNotEmpty struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"team is not empty"`
}

type _TeamFull struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"team is full"`
}

type _AccountAlreadyHasTeam struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"account already has a team"`
}

type _AccountHasNoTeam struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"account does not belong to a team"`
}

type _TeamSubmissionNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"submission not found"`
}
