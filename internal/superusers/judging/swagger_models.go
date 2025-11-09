package judging

import "time"

// JudgeCreateRequest contains the details for creating a new judge.
type JudgeCreateRequest struct {
	Name string `json:"name" example:"Judge Alice"`
	Pair string `json:"pair" example:"pair_group_1"`
}

// JudgeDeleteRequest contains the judge ID to delete.
type JudgeDeleteRequest struct {
	ID string `json:"id" example:"abc123"`
}

// JudgeConnectRequest contains the judge ID to request a connect token.
type JudgeConnectRequest struct {
	ID string `json:"id" example:"judge_001"`
}

// JudgeConnectResponse returns the ephemeral 2-minute connect token.
type JudgeConnectResponse struct {
	Token string `json:"token"`
}

// JudgeProgressResponse represents a judge with their current progress information.
type JudgeProgressResponse struct {
	ID           string    `json:"id" example:"abc123"`
	Name         string    `json:"name" example:"Judge Alice"`
	Pair         string    `json:"pair" example:"pair_1"`
	CurrentTeam  int       `json:"currentTeam" example:"5"`
	NextTeamTime time.Time `json:"nextTeamTime" example:"2024-01-15T14:30:00Z"`
}

// GetFinalistsResponse returns the current finalists with full team details.
type GetFinalistsResponse struct {
	Finalists []interface{} `json:"finalists" description:"Array of full team objects"`
}

// VotingResultItem represents a single finalist with its vote count.
type VotingResultItem struct {
	Team  interface{} `json:"team" description:"Full team object with all details"`
	Count int64       `json:"count" example:"42"`
}

// GetVotingResultsResponse contains voting results for all finalists.
type GetVotingResultsResponse struct {
	Results []VotingResultItem `json:"results" description:"Array of voting results sorted by count descending"`
}

// JudgeInitResponse returns the result of judge initialization with judge pair groups.
type JudgeInitResponse struct {
	Message       string        `json:"message"`
	NumTeams      int           `json:"numTeams"`
	NumJudges     int           `json:"numJudges"`
	NumPairGroups int           `json:"numPairGroups"`
	NumSteps      int           `json:"numSteps"`
	Collisions    int           `json:"collisions"`
	PairGroups    []interface{} `json:"pairGroups"`
	TeamOrderA    []string      `json:"teamOrderA"`
	TeamOrderB    []string      `json:"teamOrderB"`
}

// JudgePairGroup represents a group of judges that are paired together.
type JudgePairGroup struct {
	GroupID   int      `json:"groupId"`
	JudgeIDs  []string `json:"judgeIds"`
	PairAttr  string   `json:"pairAttr"`
	NumJudges int      `json:"numJudges"`
}

// ComputeRankingsResponse returns the computed team rankings using CrowdBT algorithm.
type ComputeRankingsResponse struct {
	RankedTeams      []string                      `json:"rankedTeams" description:"Team IDs sorted by skill (mu) descending"`
	TeamScores       map[string]float64            `json:"teamScores" description:"Team skill estimates (mu)"`
	TeamUncertainty  map[string]float64            `json:"teamUncertainty" description:"Team skill uncertainty (sigma_sq)"`
	JudgeReliability map[string]map[string]float64 `json:"judgeReliability" description:"Judge reliability parameters (alpha, beta)"`
	JudgmentCount    int                           `json:"judgmentCount" description:"Total number of judgments processed"`
}
