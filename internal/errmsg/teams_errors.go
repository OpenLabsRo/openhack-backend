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
)
