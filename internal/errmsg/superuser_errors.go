package errmsg

import "net/http"

var (
	AccountAlreadyInitialized = NewStatusError(http.StatusConflict, "account already initialized")
	SuperUserNotExists        = NewStatusError(http.StatusNotFound, "superuser does not exist")
)
