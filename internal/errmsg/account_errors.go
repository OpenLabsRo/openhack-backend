package errmsg

import (
	"net/http"
)

var AccountExists = NewStatusError(
	http.StatusConflict,
	"account already exists",
)

var AccountNotExists = NewStatusError(
	http.StatusNotFound,
	"account does not exist",
)

var AccountLoginWrongPassword = NewStatusError(
	http.StatusUnauthorized,
	"wrong password",
)

var AccountNoToken = NewStatusError(
	http.StatusUnauthorized,
	"you are not logged in",
)
