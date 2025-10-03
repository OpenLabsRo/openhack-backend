package staff

import "backend/internal/models"

// StaffRegisterResponse mirrors the payload returned when registering a staff member.
type StaffRegisterResponse struct {
	Account models.Account `json:"account"`
	Pile    int            `json:"pile"`
}
