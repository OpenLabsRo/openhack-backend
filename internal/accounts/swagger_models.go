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

// AccountEditRequest provides the shape for display-name updates.
type AccountEditRequest struct {
	Name string `json:"name"`
}
