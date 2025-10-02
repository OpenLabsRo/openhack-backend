package staff

import "backend/internal/models"

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
