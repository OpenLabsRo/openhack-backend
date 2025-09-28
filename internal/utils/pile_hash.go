package utils

import (
	"backend/internal/env"
	"strconv"
)

func BadgePileSalt() uint32 {
	if env.BADGE_PILES_SALT == "" {
		return 1
	}

	salt, err := strconv.ParseUint(env.BADGE_PILES_SALT, 10, 32)
	if err != nil {
		return 1
	}

	return uint32(salt)
}

func PileForAccount(id string, salt uint32) int {
	if env.BADGE_PILES <= 0 {
		return 0
	}

	h := Hash32(id) ^ salt
	return int(h % uint32(env.BADGE_PILES))
}
