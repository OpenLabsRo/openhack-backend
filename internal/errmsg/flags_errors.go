package errmsg

import "net/http"

var (
	FlagRequired = NewStatusError(
		http.StatusUnauthorized,
		"this feature is not available right now",
	)
)
