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
)
