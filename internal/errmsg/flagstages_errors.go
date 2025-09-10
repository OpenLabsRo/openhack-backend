package errmsg

import "net/http"

var (
	FlagStageNotFound = NewStatusError(http.StatusNotFound, "flag stage not found")
)
