package swagger

import (
	"errors"

	"backend/internal/errmsg"
)

// StatusErrorDoc re-exports the status error payload for documentation purposes.
type StatusErrorDoc = errmsg.StatusError

// StatusErrorExamples provides canonical error examples without duplicating message strings.
var StatusErrorExamples = map[string]StatusErrorDoc{
	"accountNotInitialized":    errmsg.AccountNotInitialized,
	"accountAlreadyRegistered": errmsg.AccountAlreadyRegistered,
	"accountLoginWrongPassword": errmsg.AccountLoginWrongPassword,
	"accountNoToken":           errmsg.AccountNoToken,
	"accountNotFound":          errmsg.AccountNotFound,
	"accountAlreadyInitialized": errmsg.AccountAlreadyInitialized,
	"accountAlreadyHasTeam":    errmsg.AccountAlreadyHasTeam,
	"accountHasNoTeam":         errmsg.AccountHasNoTeam,
	"teamNotFound":             errmsg.TeamNotFound,
	"teamNotEmpty":             errmsg.TeamNotEmpty,
	"teamFull":                 errmsg.TeamFull,
	"flagRequired":             errmsg.FlagRequired,
	"superUserNoToken":         errmsg.SuperUserNoToken,
	"superUserNotExists":       errmsg.SuperUserNotExists,
	"tagNotFound":              errmsg.TagNotFound,
	"tagIncomplete":            errmsg.TagIncomplete,
	"internalServerError":      errmsg.InternalServerError(errors.New("<details>")),
}
