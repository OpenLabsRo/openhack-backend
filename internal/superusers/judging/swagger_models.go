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

// JudgeInitResponse returns the result of judge initialization with two base team orders, judge order, offsets, multipliers, and base order assignments.
type JudgeInitResponse struct {
	TeamOrderA      []string `json:"teamOrderA"`
	TeamOrderB      []string `json:"teamOrderB"`
	JudgeOrder      []string `json:"judgeOrder"`
	JudgeOffset     []int    `json:"judgeOffset"`
	JudgeMultiplier []int    `json:"judgeMultiplier"`
	JudgeBaseOrder  []int    `json:"judgeBaseOrder"`
	NumTeams        int      `json:"numTeams"`
	NumJudges       int      `json:"numJudges"`
}
