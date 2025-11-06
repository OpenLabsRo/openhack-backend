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

	AccountNotFound = NewStatusError(
		http.StatusNotFound,
		"account not found",
	)

	VoucherNoPromoCode = NewStatusError(
		http.StatusForbidden,
		"no promotional code for this voucher type",
	)

	VoucherInvalidIndex = NewStatusError(
		http.StatusBadRequest,
		"invalid voucher index",
	)

	VoucherNotFound = NewStatusError(
		http.StatusNotFound,
		"voucher not found",
	)

	VoucherAccessDenied = NewStatusError(
		http.StatusForbidden,
		"access denied",
	)
)

type _AccountNotInitialized struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"account not initialized - talk to the administrator"`
}

type _AccountAlreadyRegistered struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"account already registered"`
}

type _AccountLoginWrongPassword struct {
	StatusCode int    `json:"statusCode" example:"401"`
	Message    string `json:"message" example:"wrong password"`
}

type _AccountNoToken struct {
	StatusCode int    `json:"statusCode" example:"401"`
	Message    string `json:"message" example:"you are not logged in"`
}

type _AccountNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"account not found"`
}

type _VoucherNoPromoCode struct {
	StatusCode int    `json:"statusCode" example:"403"`
	Message    string `json:"message" example:"no promotional code for this voucher type"`
}

type _VoucherInvalidIndex struct {
	StatusCode int    `json:"statusCode" example:"400"`
	Message    string `json:"message" example:"invalid voucher index"`
}

type _VoucherNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"voucher not found"`
}

type _VoucherAccessDenied struct {
	StatusCode int    `json:"statusCode" example:"403"`
	Message    string `json:"message" example:"access denied"`
}
