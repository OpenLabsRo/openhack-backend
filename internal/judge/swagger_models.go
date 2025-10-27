package judge

import "backend/internal/models"

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
