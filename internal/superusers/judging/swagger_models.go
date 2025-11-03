package judging

// JudgeCreateRequest contains the details for creating a new judge.
type JudgeCreateRequest struct {
	ID   string `json:"id" example:"judge_001"`
	Name string `json:"name" example:"Judge Alice"`
}

// JudgeConnectRequest contains the judge ID to request a connect token.
type JudgeConnectRequest struct {
	ID string `json:"id" example:"judge_001"`
}

// JudgeConnectResponse returns the ephemeral 2-minute connect token.
type JudgeConnectResponse struct {
	Token string `json:"token"`
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
