package badges

import "backend/internal/models"

// PilesResponse is a slice of badge piles, each containing accounts assigned to that pile.
type PilesResponse [][]models.Account
