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

// JudgeInitResponse returns the result of judge initialization with team order, judge order, and offsets.
type JudgeInitResponse struct {
	TeamOrder   []string `json:"teamOrder"`
	JudgeOrder  []string `json:"judgeOrder"`
	JudgeOffset []int    `json:"judgeOffset"`
	NumTeams    int      `json:"numTeams"`
	NumJudges   int      `json:"numJudges"`
}
