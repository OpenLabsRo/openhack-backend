package participants

// InitializeRequest captures the minimal data required to seed an account.
type InitializeRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
