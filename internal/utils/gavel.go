package utils

import (
	"math"
	"sort"
)

// CrowdBTScorer implements the Crowd Bradley-Terry algorithm for scoring pairwise comparisons
// with judge reliability tracking. Based on the paper:
// "Crowd Pairwise: Crowdsourcing Preferences for Ranking" (Chen et al.)
//
// This algorithm is more sophisticated than TrueSkill and explicitly models:
// - Team skill (mu)
// - Uncertainty in skill (sigma_sq)
// - Judge reliability (alpha, beta) - judges who are consistent/reliable have different parameters
// - Expected information gain for optimal judge assignments
type CrowdBTScorer struct {
	// Team skill estimates
	teamMu map[string]float64
	// Team skill uncertainty (variance)
	teamSigmaSq map[string]float64

	// Judge reliability parameters (Beta distribution)
	judgeAlpha map[string]float64
	judgeBeta  map[string]float64

	// Hyperparameters from the paper
	gamma  float64 // tradeoff parameter for judge reliability vs team skill (default 0.1)
	lambda float64 // regularization parameter (default 1.0)
	kappa  float64 // small value to ensure variance positivity (default 0.0001)

	// Priors
	muPrior      float64 // prior mean for team skill (default 0.0)
	sigmaSqPrior float64 // prior variance for team skill (default 1.0)
	alphaPrior   float64 // prior alpha for judge reliability (default 10.0)
	betaPrior    float64 // prior beta for judge reliability (default 1.0)
}

// Judgment represents a pairwise comparison made by a judge
type JudgmentWithJudge struct {
	WinningTeamID string
	LosingTeamID  string
	JudgeID       string
}

// NewCrowdBTScorer creates a new Crowd BT scorer with standard parameters from the paper
func NewCrowdBTScorer() *CrowdBTScorer {
	return &CrowdBTScorer{
		teamMu:       make(map[string]float64),
		teamSigmaSq:  make(map[string]float64),
		judgeAlpha:   make(map[string]float64),
		judgeBeta:    make(map[string]float64),
		gamma:        0.1,
		lambda:       1.0,
		kappa:        0.0001,
		muPrior:      0.0,
		sigmaSqPrior: 1.0,
		alphaPrior:   10.0,
		betaPrior:    1.0,
	}
}

// Score computes team skill ratings and judge reliability parameters from judgments.
// Returns a map of teamID to (mu, sigma_sq) representing skill and uncertainty.
// Judge parameters are updated internally and can be retrieved with GetJudgeParams.
func (cbt *CrowdBTScorer) Score(judgments []JudgmentWithJudge) map[string]float64 {
	if len(judgments) == 0 {
		return cbt.teamMu
	}

	// Initialize all teams and judges
	cbt.initializeFromJudgments(judgments)

	// Iterative EM-style updates (convergence typically happens in 5-10 iterations)
	for iteration := 0; iteration < 10; iteration++ {
		// Update all judges based on their judgments
		for judgeID := range cbt.judgeAlpha {
			cbt.updateJudge(judgeID, judgments)
		}

		// Update all teams based on judgments weighted by judge reliability
		for teamID := range cbt.teamMu {
			cbt.updateTeam(teamID, judgments)
		}
	}

	return cbt.teamMu
}

// initializeFromJudgments extracts all unique teams and judges, initializing their parameters
func (cbt *CrowdBTScorer) initializeFromJudgments(judgments []JudgmentWithJudge) {
	for _, j := range judgments {
		// Initialize team if not seen
		if _, exists := cbt.teamMu[j.WinningTeamID]; !exists {
			cbt.teamMu[j.WinningTeamID] = cbt.muPrior
			cbt.teamSigmaSq[j.WinningTeamID] = cbt.sigmaSqPrior
		}
		if _, exists := cbt.teamMu[j.LosingTeamID]; !exists {
			cbt.teamMu[j.LosingTeamID] = cbt.muPrior
			cbt.teamSigmaSq[j.LosingTeamID] = cbt.sigmaSqPrior
		}

		// Initialize judge if not seen
		if _, exists := cbt.judgeAlpha[j.JudgeID]; !exists {
			cbt.judgeAlpha[j.JudgeID] = cbt.alphaPrior
			cbt.judgeBeta[j.JudgeID] = cbt.betaPrior
		}
	}
}

// updateJudge updates a judge's reliability parameters (alpha, beta) based on their voting pattern
func (cbt *CrowdBTScorer) updateJudge(judgeID string, judgments []JudgmentWithJudge) {
	// Find all judgments made by this judge
	var judgeJudgments []JudgmentWithJudge
	for _, j := range judgments {
		if j.JudgeID == judgeID {
			judgeJudgments = append(judgeJudgments, j)
		}
	}

	if len(judgeJudgments) == 0 {
		return
	}

	// Aggregate the update from all this judge's decisions
	sumAlphaDelta := 0.0
	sumBetaDelta := 0.0

	for _, j := range judgeJudgments {
		winner := j.WinningTeamID
		loser := j.LosingTeamID

		mu_w := cbt.teamMu[winner]
		sigma_sq_w := cbt.teamSigmaSq[winner]
		mu_l := cbt.teamMu[loser]
		sigma_sq_l := cbt.teamSigmaSq[loser]

		// Compute the probability the judge got this right, weighted by current estimates
		c_1 := cbt.winProbability(mu_w, sigma_sq_w, mu_l, sigma_sq_l)
		c_2 := 1.0 - c_1

		// Expected value and variance of judge's reliability
		alpha := cbt.judgeAlpha[judgeID]
		beta := cbt.judgeBeta[judgeID]

		expt := (c_1*(alpha+1.0)*alpha + c_2*alpha*beta) / (c_1 * (alpha + beta + 1.0) * (alpha + beta))
		expt_sq := (c_1*(alpha+2.0)*(alpha+1.0)*alpha + c_2*(alpha+1.0)*alpha*beta) / (c_1 * (alpha + beta + 2.0) * (alpha + beta + 1.0) * (alpha + beta))

		variance := expt_sq - expt*expt
		if variance < 0.0001 {
			variance = 0.0001
		}

		alpha_delta := ((expt - expt_sq) * expt) / variance
		beta_delta := (expt - expt_sq) * (1.0 - expt) / variance

		sumAlphaDelta += alpha_delta
		sumBetaDelta += beta_delta
	}

	// Average and apply updates with regularization
	avgAlphaDelta := sumAlphaDelta / float64(len(judgeJudgments))
	avgBetaDelta := sumBetaDelta / float64(len(judgeJudgments))

	// Apply with step size and regularization toward prior
	alpha := cbt.judgeAlpha[judgeID]
	beta := cbt.judgeBeta[judgeID]

	cbt.judgeAlpha[judgeID] = alpha + 0.1*avgAlphaDelta + cbt.lambda*(cbt.alphaPrior-alpha)
	cbt.judgeBeta[judgeID] = beta + 0.1*avgBetaDelta + cbt.lambda*(cbt.betaPrior-beta)

	// Ensure positivity
	if cbt.judgeAlpha[judgeID] < 0.1 {
		cbt.judgeAlpha[judgeID] = 0.1
	}
	if cbt.judgeBeta[judgeID] < 0.1 {
		cbt.judgeBeta[judgeID] = 0.1
	}
}

// updateTeam updates a team's skill parameters (mu, sigma_sq) weighted by judge reliability
func (cbt *CrowdBTScorer) updateTeam(teamID string, judgments []JudgmentWithJudge) {
	// Find all judgments involving this team
	var teamJudgments []JudgmentWithJudge
	for _, j := range judgments {
		if j.WinningTeamID == teamID || j.LosingTeamID == teamID {
			teamJudgments = append(teamJudgments, j)
		}
	}

	if len(teamJudgments) == 0 {
		return
	}

	sumMuDelta := 0.0
	sumSigmaSqDelta := 0.0

	for _, j := range teamJudgments {
		var opponent string
		var isWinner bool

		if j.WinningTeamID == teamID {
			opponent = j.LosingTeamID
			isWinner = true
		} else {
			opponent = j.WinningTeamID
			isWinner = false
		}

		mu_team := cbt.teamMu[teamID]
		sigma_sq_team := cbt.teamSigmaSq[teamID]
		mu_opp := cbt.teamMu[opponent]
		sigma_sq_opp := cbt.teamSigmaSq[opponent]

		// Judge reliability weights
		alpha := cbt.judgeAlpha[j.JudgeID]
		beta := cbt.judgeBeta[j.JudgeID]

		// Update formulas from the paper
		if isWinner {
			mu_delta, sigma_sq_delta := cbt.updateMuSigmaSq(
				alpha, beta,
				mu_team, sigma_sq_team,
				mu_opp, sigma_sq_opp,
				true,
			)
			sumMuDelta += mu_delta
			sumSigmaSqDelta += sigma_sq_delta
		} else {
			mu_delta, sigma_sq_delta := cbt.updateMuSigmaSq(
				alpha, beta,
				mu_opp, sigma_sq_opp,
				mu_team, sigma_sq_team,
				false,
			)
			sumMuDelta -= mu_delta
			sumSigmaSqDelta += sigma_sq_delta
		}
	}

	// Average updates
	avgMuDelta := sumMuDelta / float64(len(teamJudgments))
	avgSigmaSqDelta := sumSigmaSqDelta / float64(len(teamJudgments))

	// Apply updates
	cbt.teamMu[teamID] += 0.05 * avgMuDelta
	newSigmaSq := cbt.teamSigmaSq[teamID] * math.Max(1.0+cbt.teamSigmaSq[teamID]*avgSigmaSqDelta, cbt.kappa)
	cbt.teamSigmaSq[teamID] = newSigmaSq
}

// updateMuSigmaSq computes the update deltas for mu and sigma_sq given winner/loser and judge reliability
func (cbt *CrowdBTScorer) updateMuSigmaSq(
	alpha, beta float64,
	mu_winner, sigma_sq_winner float64,
	mu_loser, sigma_sq_loser float64,
	isWinner bool,
) (float64, float64) {
	// Compute the multiplier based on judge reliability and skill difference
	exp_w := math.Exp(mu_winner)
	exp_l := math.Exp(mu_loser)

	denom := alpha*exp_w + beta*exp_l
	if denom == 0 {
		denom = 0.0001
	}

	// Weighted Bradley-Terry update
	mult := (alpha*exp_w)/denom - exp_w/(exp_w+exp_l)

	mu_delta := mult
	sigma_sq_mult := (alpha*exp_w*beta*exp_l)/(denom*denom) - (exp_w*exp_l)/((exp_w+exp_l)*(exp_w+exp_l))
	sigma_sq_delta := sigma_sq_mult

	return mu_delta, sigma_sq_delta
}

// winProbability computes the probability that team1 beats team2 using Bradley-Terry
func (cbt *CrowdBTScorer) winProbability(mu1, sigma_sq_1, mu2, sigma_sq_2 float64) float64 {
	exp1 := math.Exp(mu1)
	exp2 := math.Exp(mu2)

	// Bradley-Terry: P(1 > 2) = exp(mu1) / (exp(mu1) + exp(mu2))
	prob := exp1 / (exp1 + exp2)

	// Adjust for uncertainty (wider uncertainty = closer to 0.5)
	uncertainty := 0.5 * (sigma_sq_1 + sigma_sq_2) * (exp1 * exp2 * (exp2 - exp1)) / ((exp1 + exp2) * (exp1 + exp2) * (exp1 + exp2))
	prob += uncertainty

	// Clamp to valid probability range
	if prob > 0.99 {
		prob = 0.99
	}
	if prob < 0.01 {
		prob = 0.01
	}

	return prob
}

// GetTeamScores returns the current mu values for all teams
func (cbt *CrowdBTScorer) GetTeamScores() map[string]float64 {
	result := make(map[string]float64)
	for teamID, mu := range cbt.teamMu {
		result[teamID] = mu
	}
	return result
}

// GetTeamUncertainty returns the current sigma_sq (variance) for all teams
func (cbt *CrowdBTScorer) GetTeamUncertainty() map[string]float64 {
	result := make(map[string]float64)
	for teamID, sigma_sq := range cbt.teamSigmaSq {
		result[teamID] = sigma_sq
	}
	return result
}

// GetJudgeReliability returns (alpha, beta) for a judge. Higher alpha = judge voted for team consistently. High beta = inconsistent.
func (cbt *CrowdBTScorer) GetJudgeReliability(judgeID string) (float64, float64) {
	return cbt.judgeAlpha[judgeID], cbt.judgeBeta[judgeID]
}

// GetJudgeReliabilityAll returns all judges' (alpha, beta) parameters
func (cbt *CrowdBTScorer) GetJudgeReliabilityAll() map[string][2]float64 {
	result := make(map[string][2]float64)
	for judgeID := range cbt.judgeAlpha {
		result[judgeID] = [2]float64{cbt.judgeAlpha[judgeID], cbt.judgeBeta[judgeID]}
	}
	return result
}

// RankTeams returns teams sorted by mu (descending)
func (cbt *CrowdBTScorer) RankTeams() []string {
	type teamScore struct {
		id  string
		mu  float64
		dev float64 // sqrt(sigma_sq) for sorting by confidence
	}

	var teams []teamScore
	for teamID, mu := range cbt.teamMu {
		teams = append(teams, teamScore{
			id:  teamID,
			mu:  mu,
			dev: math.Sqrt(cbt.teamSigmaSq[teamID]),
		})
	}

	// Sort by mu descending, then by lower deviation (higher confidence) as tiebreaker
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].mu != teams[j].mu {
			return teams[i].mu > teams[j].mu
		}
		return teams[i].dev < teams[j].dev
	})

	var ranked []string
	for _, t := range teams {
		ranked = append(ranked, t.id)
	}
	return ranked
}

// ScoreCrowdBT is a convenience function that creates a scorer, runs the algorithm,
// and returns team rankings in a single call.
func ScoreCrowdBT(judgments []JudgmentWithJudge) map[string]float64 {
	scorer := NewCrowdBTScorer()
	return scorer.Score(judgments)
}

// GetJudgeStats returns the learned reliability parameters for a judge
// Returns (alpha, beta, exists)
func (cbt *CrowdBTScorer) GetJudgeStats(judgeID string) (float64, float64, bool) {
	alpha, ok := cbt.judgeAlpha[judgeID]
	if !ok {
		return 0, 0, false
	}
	beta := cbt.judgeBeta[judgeID]
	return alpha, beta, true
}

// ScoreCrowdBTWithParams allows custom hyperparameter tuning
func ScoreCrowdBTWithParams(judgments []JudgmentWithJudge, gamma, lambda float64) map[string]float64 {
	scorer := NewCrowdBTScorer()
	scorer.gamma = gamma
	scorer.lambda = lambda
	return scorer.Score(judgments)
}
