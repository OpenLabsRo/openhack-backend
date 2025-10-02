package superusers

import "backend/internal/models"

// SuperUserLoginRequest documents the credentials payload for superuser login.
type SuperUserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SuperUserLoginResponse represents the login token and principal context.
type SuperUserLoginResponse struct {
	Token     string           `json:"token"`
	Superuser models.SuperUser `json:"superuser"`
}

// BadgePilesResponse is a slice of badge piles, each containing a slice of accounts.
type BadgePilesResponse [][]models.Account
