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
	BadgeIncomplete = NewStatusError(
		http.StatusConflict,
		"badge data is incomplete",
	)
	BadgeNotFound = NewStatusError(
		http.StatusNotFound,
		"badge not found",
	)
)
