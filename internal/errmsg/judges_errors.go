package errmsg

import (
	"net/http"
)

var (
	JudgeNotFound = NewStatusError(
		http.StatusNotFound,
		"judge not found",
	)

	JudgeAlreadyExists = NewStatusError(
		http.StatusConflict,
		"judge already exists",
	)

	JudgingFinished = NewStatusError(
		http.StatusOK,
		"judging finished",
	)
)

type _JudgeNotFound struct {
	StatusCode int    `json:"statusCode" example:"404"`
	Message    string `json:"message" example:"judge not found"`
}

type _JudgeAlreadyExists struct {
	StatusCode int    `json:"statusCode" example:"409"`
	Message    string `json:"message" example:"judge already exists"`
}

type _JudgingFinished struct {
	StatusCode int    `json:"statusCode" example:"200"`
	Message    string `json:"message" example:"judging finished"`
}
