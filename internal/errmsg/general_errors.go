package errmsg

import "net/http"

var InternalServerError = NewStatusError(http.StatusInternalServerError, "internal server error")
