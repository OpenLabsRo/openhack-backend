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

// AccountInitializeRequest captures the minimal data required to seed an account.
type AccountInitializeRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// FlagSetRequest toggles a single feature flag.
type FlagSetRequest struct {
	Flag  string `json:"flag"`
	Value bool   `json:"value"`
}

// FlagAssignments represents a collection of flag states keyed by flag name.
type FlagAssignments map[string]bool

// FlagUnsetRequest identifies the flag to remove.
type FlagUnsetRequest struct {
	Flag string `json:"flag"`
}

// FlagStageCreateRequest documents the payload to create a new stage.
type FlagStageCreateRequest struct {
	Name    string   `json:"name"`
	TurnOn  []string `json:"turnon"`
	TurnOff []string `json:"turnoff"`
}

// FlagStageDeleteRequest identifies which stage should be removed.
type FlagStageDeleteRequest struct {
	ID string `json:"id"`
}

// BadgePilesResponse is a slice of badge piles, each containing a slice of accounts.
type BadgePilesResponse [][]models.Account

// TagAssignRequest represents the payload for assigning a check-in tag.
type TagAssignRequest struct {
	AccountID string `json:"accountID"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

// StaffRegisterResponse mirrors the payload returned when registering a staff member.
type StaffRegisterResponse struct {
	Account models.Account `json:"account"`
	Pile    int            `json:"pile"`
}
