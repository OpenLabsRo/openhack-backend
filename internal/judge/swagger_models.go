package judge

import (
	"backend/internal/models"
	"time"
)

// JudgeUpgradeRequest exchanges a connect token for a full 24-hour token.
type JudgeUpgradeRequest struct {
	Token string `json:"token"`
}

// JudgeUpgradeResponse returns the full judge token and judge data.
type JudgeUpgradeResponse struct {
	Token string       `json:"token"`
	Judge models.Judge `json:"judge"`
}

// NextTeamResponse returns the team ID for the judge to evaluate next.
type NextTeamResponse struct {
	TeamID string `json:"teamID" example:"team_001"`
}

// CreateJudgmentRequest contains the teams being compared in a judgment.
type CreateJudgmentRequest struct {
	WinningTeamID string `json:"winningTeamID" example:"team_001"`
	LosingTeamID  string `json:"losingTeamID" example:"team_002"`
}

// JudgeInfoResponse returns the judge's current progress and timing information.
type JudgeInfoResponse struct {
	CurrentTeam  int       `json:"currentTeam" example:"0"`
	NextTeamTime time.Time `json:"nextTeamTime" example:"2024-01-01T12:05:00Z"`
}
