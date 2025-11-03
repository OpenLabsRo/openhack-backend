# Crowd Bradley-Terry (Crowd BT) Scoring Algorithm

## Overview

The Crowd BT algorithm is a sophisticated pairwise comparison scoring system that accounts for **judge reliability**. Unlike simple win/loss counting, Crowd BT learns which judges are trustworthy and weights their votes accordingly.

**Paper**: "Crowd Pairwise: Crowdsourcing Preferences for Ranking" (Chen et al.)
**Reference**: http://people.stern.nyu.edu/xchen3/images/crowd_pairwise.pdf

## Why Use Crowd BT?

In hackathon judging, different judges have different standards and biases:
- Some judges might consistently underrate certain types of projects
- Some judges might be more reliable/consistent than others
- Some judges might have strong but legitimate preferences (e.g., healthcare vs. web3)

**TrueSkill assumption**: All votes are equally valid
**Crowd BT reality**: Some judges are more reliable than others

### Example

Say Project A gets 8 wins, Project B gets 2 wins from only 2 judges:
- **Judge 1** (medical expert): Always picks A (8-0)
- **Judge 2** (blockchain expert): Always picks B (0-2)

**TrueSkill**: "A is clearly better" (8:2 ratio)

**Crowd BT**: "Both judges have strong preferences but seem biased. Their votes are less predictive. Let's be more conservative."

## Data Model

### Team/Item State
```go
Mu       float64  // Skill rating (higher = better)
SigmaSq  float64  // Uncertainty/variance (higher = less confident)
```

### Judge State
```go
Alpha    float64  // Reliability parameter for Beta distribution
Beta     float64  // Reliability parameter for Beta distribution
```

**Interpretation**:
- Judges who make consistent decisions have parameters that don't change much
- Judges who make contradictory decisions have parameters that update significantly
- Over time, reliable judges' parameters stabilize; unreliable judges' parameters diverge

## Usage

### Basic Usage

```go
package main

import (
    "backend/internal/utils"
)

func scoreJudgments() {
    // Create judgment records from your database
    judgments := []utils.JudgmentWithJudge{
        {
            WinningTeamID: "TEAM01",
            LosingTeamID:  "TEAM02",
            JudgeID:       "JUDGE_A",
        },
        {
            WinningTeamID: "TEAM03",
            LosingTeamID:  "TEAM01",
            JudgeID:       "JUDGE_B",
        },
        // ... more judgments
    }

    // Create scorer
    scorer := utils.NewCrowdBTScorer()

    // Run the algorithm (returns map of teamID -> mu)
    scores := scorer.Score(judgments)

    // Get rankings
    ranking := scorer.RankTeams()  // Returns []string in order of best -> worst

    for i, teamID := range ranking {
        println(i+1, teamID, scores[teamID])
    }
}
```

### Getting Judge Reliability

```go
scorer := utils.NewCrowdBTScorer()
scores := scorer.Score(judgments)

// Get specific judge's reliability
alpha, beta := scorer.GetJudgeReliability("JUDGE_A")
println("Judge A - Alpha:", alpha, "Beta:", beta)

// Get all judges
allJudgeStats := scorer.GetJudgeReliabilityAll()
for judgeID, params := range allJudgeStats {
    println(judgeID, "- Alpha:", params[0], "Beta:", params[1])
}
```

### Understanding Judge Parameters

Higher `alpha` relative to `beta` means:
- Judge makes "obvious" picks (high-skill vs low-skill)
- Judge is more reliable

Higher `beta` relative to `alpha` means:
- Judge makes unconventional picks (upsets)
- Judge may be noisy or have different criteria

Equal `alpha` and `beta` means:
- Judge's decisions are random/uninformative

### Custom Hyperparameters

```go
scorer := utils.NewCrowdBTScorer()

// Adjust learning rates
scorer.gamma = 0.2   // Higher = trust judge updates more (default 0.1)
scorer.lambda = 0.5  // Higher = regularize toward priors more (default 1.0)

scores := scorer.Score(judgments)
```

### Getting Team Uncertainty

```go
scorer := utils.NewCrowdBTScorer()
scores := scorer.Score(judgments)

// Get uncertainty estimates
uncertainty := scorer.GetTeamUncertainty()
for teamID, sigma_sq := range uncertainty {
    confidence := 1.0 / math.Sqrt(sigma_sq)  // Lower sigma = higher confidence
    println(teamID, "confidence:", confidence)
}
```

## Algorithm Details

### Initialization

For each team and judge, parameters are initialized with priors:
- Teams: `mu = 0.0, sigma_sq = 1.0`
- Judges: `alpha = 10.0, beta = 1.0`

### Bradley-Terry Model

The core probabilistic model uses the Bradley-Terry model for pairwise comparisons:

```
P(Team A beats Team B) = exp(mu_A) / (exp(mu_A) + exp(mu_B))
```

This is different from logistic (which uses `mu` directly) and is mathematically tailored for pairwise preferences.

### EM-Style Iterations

The algorithm iterates, alternating between:

1. **Update Judges**: For each judge, compute how reliable they are by seeing if their votes matched expected outcomes
   - If a judge votes for the expected winner: alpha increases
   - If a judge votes for an upset: beta increases

2. **Update Teams**: For each team, recompute skill based on weighted judgments
   - Votes from reliable judges (high alpha) are weighted more
   - Votes from unreliable judges (high beta) are weighted less

Convergence typically happens in 10 iterations.

### Math Behind Judge Updates

For each judgment a judge makes:
```
C = P(their pick was correct | current skill estimates)
```

If `C` is high (they picked the obvious winner):
- Their `alpha` parameter increases (they're reliable)

If `C` is low (they picked an upset):
- Their `beta` parameter increases (they're noisy/have different taste)

Over many judgments, the algorithm learns which judges to trust.

## Hyperparameters

### GAMMA (default 0.1)
- Controls how much to update judge reliability
- Higher values: judge parameters change more rapidly
- Lower values: judge parameters are stickier
- Use **higher** if judges are truly variable in quality
- Use **lower** if you want to see more dramatic changes in team scores

### LAMBDA (default 1.0)
- Regularization parameter, pulls parameters toward priors
- Higher values: parameters regularize more (smoother, more conservative)
- Lower values: parameters fit the data more tightly
- Use **higher** for noisy data
- Use **lower** if you trust your judgment data

### KAPPA (default 0.0001)
- Minimum variance to prevent numerical issues
- Leave as-is in most cases

### Priors
- `MU_PRIOR`: 0.0 (teams start at neutral skill)
- `SIGMA_SQ_PRIOR`: 1.0 (high initial uncertainty)
- `ALPHA_PRIOR`: 10.0 (assume judges are somewhat reliable initially)
- `BETA_PRIOR`: 1.0 (assume judges are not very noisy initially)

## Integration with Your System

To use with your existing models:

```go
// In your judging/ranking handler:

import (
    "backend/internal/models"
    "backend/internal/utils"
)

func rankTeams() {
    // Fetch all judgments from database
    allJudgments, _ := models.GetAllJudgments()

    // Convert to scorer format
    judgmentData := make([]utils.JudgmentWithJudge, len(allJudgments))
    for i, j := range allJudgments {
        judgmentData[i] = utils.JudgmentWithJudge{
            WinningTeamID: j.WinningTeamID,
            LosingTeamID:  j.LosingTeamID,
            JudgeID:       j.JudgeID,
        }
    }

    // Score
    scorer := utils.NewCrowdBTScorer()
    scores := scorer.Score(judgmentData)
    ranking := scorer.RankTeams()

    // Use ranking...
    for i, teamID := range ranking {
        // Update team with final rank
        team := models.Team{ID: teamID}
        team.Get()
        team.Rank = i + 1
        team.Score = scores[teamID]
        team.Update()
    }
}
```

## Interpreting Results

### Mu (Skill Rating)
- **Range**: Typically -5 to 5 in practice
- **Meaning**: Relative ranking. Difference of 1.0 in mu means exponentially more likely to win
- **Example**: Team A (mu=1.0) vs Team B (mu=0.0) → ~73% win probability for Team A

### Sigma-Sq (Uncertainty)
- **Higher value**: Less confident in that team's ranking
  - Teams with few judgments have high sigma_sq
  - Teams everyone agrees about have low sigma_sq
- **Use for**: Confidence intervals, identifying close rankings

### Judge Alpha/Beta
- **Alpha >> Beta**: Judge is reliable and votes for obvious winners
- **Beta >> Alpha**: Judge is noisy or has unconventional taste
- **Alpha ≈ Beta**: Judge is random/uninformative
- **Both low**: Judge hasn't provided enough information yet

## Common Issues

### Issue: All teams have similar scores

**Cause**: Not enough judgments or judges are too disagreeable

**Solutions**:
- Lower `gamma` so judge parameters stabilize more
- Higher `lambda` for more regularization
- Ensure judges are comparing comparable projects

### Issue: One judge's votes dominate everything

**Cause**: That judge made way more judgments than others

**Solutions**:
- This is actually correct behavior - more data should influence more
- If judge is unreliable, algorithm will downweight them via alpha/beta
- If judge is reliable, they deserve more influence

### Issue: Scores barely change between iterations

**Cause**: Convergence achieved (normal)

**Solutions**:
- No action needed - algorithm is working
- This is good! Means solution is stable

### Issue: Judge reliability parameters blow up (very high alpha/beta)

**Cause**: Judge is highly consistent (either very reliable or very noisy)

**Solutions**:
- Lower `gamma` to dampen updates
- Check if judge is actually making sense (or if data has issues)

## Performance Notes

- Time complexity: O(iterations × judgments × judges)
- With 10 iterations, 1000 judgments, 50 judges: ~500k operations (milliseconds)
- Memory: O(teams + judges) for storing parameters

## References

- Original paper: Chen et al., "Crowd Pairwise: Crowdsourcing Preferences for Ranking"
- Bradley-Terry model: Bradley & Terry (1952)
- Gavel implementation: https://github.com/anishathalye/gavel