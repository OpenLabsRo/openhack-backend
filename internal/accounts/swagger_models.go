package accounts

import "backend/internal/models"

// AccountCheckRequest mirrors the payload for /accounts/check.
type AccountCheckRequest struct {
	Email string `json:"email"`
}

// AccountCheckResponse describes the registration state payload from /accounts/check.
type AccountCheckResponse struct {
	Registered bool `json:"registered"`
}

// CredentialRequest captures login and registration credentials.
type CredentialRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AccountTokenResponse embeds the refreshed token and account snapshot returned by several account endpoints.
type AccountTokenResponse struct {
	Token   string         `json:"token"`
	Account models.Account `json:"account"`
}

// AccountEditRequest provides the shape for name updates.
type AccountEditRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// VotingStatusResponse returns the current voting status for a participant.
type VotingStatusResponse struct {
	VotingOpen bool     `json:"votingOpen"`
	HasVoted   bool     `json:"hasVoted"`
	Finalists  []string `json:"finalists"`
}

// VotingFinalistsResponse returns the list of finalist teams.
type VotingFinalistsResponse struct {
	Finalists []map[string]interface{} `json:"finalists"`
}

// VotingCastRequest contains the team ID to vote for.
type VotingCastRequest struct {
	TeamID string `json:"teamID"`
}

// VotingCastResponse returns confirmation of vote submission.
type VotingCastResponse struct {
	Message string `json:"message"`
}
