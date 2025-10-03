package utils

import (
	"backend/internal/env"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"time"
)

// BadgePileSalt returns the current salt value sourced from configuration.
// It falls back to 1 when the salt is unset or cannot be parsed.
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

// PileForAccount maps an account identifier into a badge pile using the supplied salt.
// The mapping is deterministic so staff can pre-stage badge pickup.
func PileForAccount(id string, salt uint32) int {
	if env.BADGE_PILES <= 0 {
		return 0
	}

	h := Hash32(id) ^ salt
	return int(h % uint32(env.BADGE_PILES))
}

func Hash32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// PileWithSalt applies the hash/salt pairing for a custom pile count.
func PileWithSalt(id string, salt uint32, n int) int {
	// Mix ID hash with salt deterministically
	h := Hash32(id) ^ salt
	return int(h % uint32(n))
}

// BalanceScore measures how evenly accounts are distributed across piles.
// A lower value indicates a more balanced split.
func BalanceScore(counts []int) float64 {
	total := 0
	for _, c := range counts {
		total += c
	}
	exp := float64(total) / float64(len(counts))
	chi := 0.0
	for _, c := range counts {
		diff := float64(c) - exp
		chi += (diff * diff) / exp
	}
	// Tie-break with max−min to avoid weird shapes
	min, max := math.MaxInt32, math.MinInt32
	for _, c := range counts {
		if c < min {
			min = c
		}
		if c > max {
			max = c
		}
	}
	return chi + float64(max-min)*0.01
}

// ChooseBestSalt scans salts deterministically, returning the best distribution found
// within the supplied evaluation/time bounds.
func ChooseBestSalt(ids []string, n, maxEvaluations int, maxDuration time.Duration) (bestSalt uint32, bestCounts []int) {
	if n <= 0 || len(ids) == 0 {
		return 0, make([]int, max(n, 0))
	}

	if maxDuration <= 0 {
		maxDuration = time.Second
	}

	start := time.Now()
	counts := make([]int, n)
	bestScore := math.Inf(1)
	evaluations := 0

	for salt := uint32(0); ; salt++ {
		resetCounts(counts)
		for _, id := range ids {
			counts[PileWithSalt(id, salt, n)]++
		}

		score := BalanceScore(counts)
		evaluations++

		if score < bestScore {
			bestScore = score
			bestSalt = salt
			bestCounts = append([]int(nil), counts...)

			if score == 0 {
				break
			}
		}

		if maxEvaluations > 0 && evaluations >= maxEvaluations {
			break
		}

		if time.Since(start) >= maxDuration {
			break
		}

		if salt == ^uint32(0) {
			break
		}
	}

	if bestCounts == nil {
		bestCounts = append([]int(nil), counts...)
	}

	return bestSalt, bestCounts
}

func resetCounts(counts []int) {
	for i := range counts {
		counts[i] = 0
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// FindSalt is a helper utility for manual experimentation when tuning pile distributions.
// It prints the best salt to stdout and is not used in the API path.
func FindSalt(ids []string) {
	const piles = 5
	const trials = 1_000_000 // increase for even better balance

	salt, counts := ChooseBestSalt(ids, piles, trials, time.Minute)

	fmt.Printf("Chosen salt: %08x\n", salt)
	total := 0
	for i, c := range counts {
		total += c
		fmt.Printf("Pile %d: %d (%.1f%%)\n", i, c, 100*float64(c)/float64(len(ids)))
	}
	fmt.Printf("Total: %d\n", total)

	// Example: compute one participant's pile on the fly on event day
	example := ids[0]
	p := PileWithSalt(example, salt, piles)
	fmt.Printf("Example %q -> pile %d\n", example, p)

	// If you want to persist the salt as a short string:
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, salt)
	fmt.Printf("Salt (hex): %x\n", buf)
}
