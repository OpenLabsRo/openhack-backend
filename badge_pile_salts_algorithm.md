# Badge Pile Salt Algorithm Analysis & Improvements

## Current State

### What the Algorithm Does
- **Location**: `internal/utils/badge_pile.go` - `ChooseBestSalt()` function
- **Strategy**: Sequential salt search from 0, 1, 2, 3, ... up to either:
  - `maxEvaluations` (default 10,000)
  - `maxDuration` (default 1 second)
  - Or uint32 overflow
- **Balance Metric**: Chi-squared statistic + small max-min range penalty:
  ```
  score = chi + 0.01 * (max - min)
  ```
- **Entry Point**: Called by `internal/superusers/badges/service.go` - `computeAndPersistBadgePileSalt()`

### Current Results (Test Case: 52 participants, 4 piles)
- Distribution: [11, 15, 13, 13]
- Expected per pile: ~13
- Max-min range: 4
- Evaluation time: ~1.1 seconds with 10,000 trials

### Observed Problem
With ~52 participants across N piles, expecting ~52/N per pile, but results are "wildly varying" rather than equilibrated.
- Example: [11, 15, 13, 13] shows uneven distribution (11 vs 15 is a 4-account swing)
- Goal: Achieve balanced distributions like [13, 13, 13, 13] or [12, 13, 13, 14]

---

## Root Causes

1. **Linear Sequential Bias**: Searching salt 0, 1, 2, ... assumes lower salts are better. No theoretical reason this is true; you just get lucky if salt=0 happens to work.

2. **Search Space Waste**: With 10,000 evaluations, only checking salts 0-10k out of 4 billion possible. Never explores the broader space.

3. **Misaligned Metric**: Chi-squared measures variance, but the real goal is **balanced piles** (minimize `max - min`). The tiny `0.01` weight on range means chi-squared dominates, not range.

4. **Non-Exploratory**: No mechanism to escape local minima or explore distant regions of the search space.

---

## Concrete Solutions (Ranked by Impact & Effort)

### Option 1: Improve Balance Metric (HIGH IMPACT, EASY)
**Effort**: Change 5 lines in `BalanceScore()` function

**Current Code**:
```go
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
    min, max := math.MaxInt32, math.MinInt32
    for _, c := range counts {
        if c < min {
            min = c
        }
        if c > max {
            max = c
        }
    }
    return chi + 0.01 * (max - min)  // <-- Problem: 0.01 weight is too small
}
```

**Proposed Change**:
```go
// Replace final return with:
return chi*0.2 + float64(max-min)*10.0
```

**Rationale**: Directly prioritize balanced piles (minimizing max-min range). Chi-squared still provides fine-tuning but doesn't dominate.

**Expected Impact**: May achieve [13, 13, 13, 13] or [12, 13, 13, 14] within same time budget.

---

### Option 2: Random Sampling (HIGH IMPACT, MODERATE EFFORT)
**Effort**: Replace sequential loop with random sampling in `ChooseBestSalt()`

**Current Code**:
```go
for salt := uint32(0); ; salt++ {
    resetCounts(counts)
    for _, id := range ids {
        counts[PileWithSalt(id, salt, n)]++
    }
    score := BalanceScore(counts)
    // ...
}
```

**Proposed Change**:
```go
rand.Seed(time.Now().UnixNano())
for i := 0; i < maxSalts; i++ {
    if time.Since(start) >= maxDuration {
        break
    }
    
    salt := rand.Uint32()  // Random salt from full uint32 space
    resetCounts(counts)
    for _, id := range ids {
        counts[PileWithSalt(id, salt, n)]++
    }
    score := BalanceScore(counts)
    // ... same best-tracking logic
}
```

**Rationale**: 
- Explores full uint32 space, not biased toward 0
- 100k random samples still evaluates in ~1 second
- With better metric (Option 1), much more likely to find truly balanced salts

**Expected Impact**: Explores 10x more of the search space.

---

### Option 3: Increase Search Budget (LOWEST EFFORT)
**Effort**: Change call site in `internal/superusers/badges/service.go`

**Current Code**:
```go
salt, counts := utils.ChooseBestSalt(ids, env.BADGE_PILES, 0, time.Second)
```

**Proposed Change**:
```go
salt, counts := utils.ChooseBestSalt(ids, env.BADGE_PILES, 100_000, 3*time.Second)
```

**Rationale**: Explore 10x more salts in 3x more time. Still completes quickly.

**Expected Impact**: Better coverage of salt space with existing sequential algorithm.

---

### Option 4: Random Sampling + Improved Metric (RECOMMENDED)
**Effort**: Combine Options 1 + 2 (two function changes)

This is the **sweet spot**:
- Explores full uint32 space (not biased toward low values)
- Directly optimizes for balanced piles (metric improvement)
- Scales efficiently (100k samples in ~1 second)
- No determinism requirement (salt persisted in MongoDB anyway)

**Implementation Steps**:
1. Update `BalanceScore()` in `badge_pile.go` (change metric weights)
2. Update `ChooseBestSalt()` in `badge_pile.go` (use random sampling)
3. Optionally update call site (increase trials if needed)

**Expected Result**: For 52 accounts across 4 piles, expect distributions like [12, 13, 13, 14] instead of [11, 15, 13, 13].

---

### Option 5: Local Refinement (Hill Climbing) (ADVANCED)
**Effort**: Rewrite main loop in `ChooseBestSalt()`, add neighborhood exploration

**Approach**:
1. Start with a random salt
2. Try nearby salts (e.g., ±100 range)
3. Move to best neighbor, repeat
4. When stuck, jump to new random starting point
5. Continue until evaluation budget exhausted

**Rationale**: Efficient exploration of promising regions. Can find very good local optima.

**Pros**: Sophisticated, efficient
**Cons**: More complex, risk of local minima (though restarts mitigate)

---

## Recommendations for Tomorrow

### Priority 1: Just Fix the Metric (Quick Validation)
Try **Option 1 alone** first:
- Change `BalanceScore()` return from `chi + 0.01 * (max-min)` to `chi*0.2 + float64(max-min)*10.0`
- Re-run test: `go test ./test/superusers -run "TestBadgePiles" -v`
- Check if [13, 13, 13, 13] or [12, 13, 13, 14] achievable

**Takes 5 minutes. May solve the problem immediately.**

### Priority 2: Add Random Sampling (If Priority 1 Insufficient)
Implement **Option 2**:
- Replace sequential loop with `rand.Uint32()` sampling
- Adjust `maxSalts` to 50,000-100,000 in call site
- Re-test

**Takes 15 minutes. Explores much more of search space.**

### Priority 3: Fine-Tune Together
Combine **Options 1 + 2 + 3**:
- Better metric + random sampling + higher budget
- Test different budget levels (50k, 100k, 200k trials)
- Find the sweet spot for your dataset size and time tolerance

---

## Key Metrics to Track

After each change, log and compare:
```
- Computed badge pile salt: <salt value>
- Reported pile counts from compute endpoint: <distribution>
- Computing badge piles took: <time>
- Max-min range: <max - min>
- Chi-squared score: <chi value>
```

Current test already outputs this—rerun after changes to see improvements.

---

## Files to Modify

1. **`internal/utils/badge_pile.go`**
   - `BalanceScore()` — update metric weights
   - `ChooseBestSalt()` — replace sequential loop with random sampling (if Option 2+)

2. **`internal/superusers/badges/service.go`**
   - `computeAndPersistBadgePileSalt()` — call `ChooseBestSalt()` with higher budget (if Option 3+)

3. **Test file** (no changes needed, but monitor output)
   - `test/superusers/badge_piles_test.go` — already logs distribution and salt value

---

## Notes

- **Determinism not required**: Salt is persisted in MongoDB, so finding different salt on re-run is fine
- **Participants are static**: Once a set of 52 accounts is frozen, the salt is frozen with it
- **Data is unbiased**: No assumptions about ID structure; algorithm must explore broadly
- **Time budget**: 1-3 seconds is acceptable for salt computation (one-time or rare operation)

---

## Success Criteria

**Before**: [11, 15, 13, 13] with range=4
**Target**: [12, 13, 13, 14] or [13, 13, 13, 13] with range≤1

Test by running:
```bash
go test ./test/superusers -run "TestBadgePiles" -v
```

Look for "Newly created accounts per pile" output.