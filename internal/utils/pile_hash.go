package utils

import (
	"backend/internal/env"
)

func PileForAccount(id string, salt uint32) int {
	h := Hash32(id) ^ salt
	return int(h % uint32(env.BADGE_PILES))
}
