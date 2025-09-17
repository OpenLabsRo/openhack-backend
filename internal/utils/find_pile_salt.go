package utils

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
)

func hash32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func pileWithSalt(id string, salt uint32, n int) int {
	// Mix ID hash with salt deterministically
	h := hash32(id) ^ salt
	return int(h % uint32(n))
}

func balanceScore(counts []int) float64 {
	// Use chi-square vs uniform as score (lower is better)
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

func chooseBestSalt(ids []string, n, trials int) (bestSalt uint32, bestCounts []int) {
	bestScore := math.Inf(1)
	var counts []int
	for t := 0; t < trials; t++ {
		salt := uint32(rand.Int63())
		counts = make([]int, n)
		for _, id := range ids {
			p := pileWithSalt(id, salt, n)
			counts[p]++
		}
		score := balanceScore(counts)
		if score < bestScore {
			bestScore = score
			bestSalt = salt
			bestCounts = append([]int(nil), counts...)
		}
	}
	return
}

func FindSalt(ids []string) {
	const piles = 5
	const trials = 1_000_000 // increase for even better balance

	salt, counts := chooseBestSalt(ids, piles, trials)

	fmt.Printf("Chosen salt: %08x\n", salt)
	total := 0
	for i, c := range counts {
		total += c
		fmt.Printf("Pile %d: %d (%.1f%%)\n", i, c, 100*float64(c)/float64(len(ids)))
	}
	fmt.Printf("Total: %d\n", total)

	// Example: compute one participant's pile on the fly on event day
	example := ids[0]
	p := pileWithSalt(example, salt, piles)
	fmt.Printf("Example %q -> pile %d\n", example, p)

	// If you want to persist the salt as a short string:
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, salt)
	fmt.Printf("Salt (hex): %x\n", buf)
}
