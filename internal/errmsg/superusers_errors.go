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
