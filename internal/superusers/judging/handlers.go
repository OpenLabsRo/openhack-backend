package judging

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// judgeConnectHandler generates an ephemeral 2-minute connect token for a judge.
// @Summary Generate judge connect token
// @Description Creates a short-lived 2-minute token that is displayed via QR code for judges to scan and exchange for a full session token.
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body JudgeConnectRequest true "Judge ID"
// @Success 200 {object} JudgeConnectResponse
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 404 {object} errmsg._JudgeNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judges/connect [post]
func judgeConnectHandler(c fiber.Ctx) error {
	var body struct {
		ID string `json:"id"`
	}
	json.Unmarshal(c.Body(), &body)

	judge := models.Judge{ID: body.ID}
	err := judge.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.JudgeNotFound)
	}

	superuser := models.SuperUser{}
	utils.GetLocals(c, "superuser", &superuser)

	token := judge.IssueJudgeConnectToken()

	events.Em.JudgeConnectTokenIssued(
		superuser.Username,
		judge.ID,
	)

	return c.JSON(bson.M{
		"token": token,
	})
}

// judgeCreateHandler creates a new judge.
// @Summary Create a new judge
// @Description Creates a new judge with the given ID and name.
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body JudgeCreateRequest true "Judge details"
// @Success 200 {object} models.Judge
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 409 {object} errmsg._JudgeAlreadyExists
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judging/judge [post]
func judgeCreateHandler(c fiber.Ctx) error {
	var body struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	judge := models.Judge{
		ID:          body.ID,
		Name:        body.Name,
		CurrentTeam: -1,
	}

	serr := judge.Initialize()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	superuser := models.SuperUser{}
	utils.GetLocals(c, "superuser", &superuser)

	events.Em.JudgeCreated(superuser.Username, judge.ID, judge.Name)

	return c.JSON(judge)
}

// computeRankingsHandler computes team rankings using the CrowdBT algorithm
// from all judgments currently in the database.
// @Summary Compute team rankings from judgments
// @Description Uses the Crowd Bradley-Terry algorithm to compute team rankings and judge reliability scores from all judgments in the database.
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Success 200 {object} ComputeRankingsResponse
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judging/compute-rankings [post]
func computeRankingsHandler(c fiber.Ctx) error {
	// Fetch all judgments from the database
	judgments, serr := models.GetAllJudgments()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	// Convert Judgment models to JudgmentWithJudge format for the scorer
	var scorerJudgments []utils.JudgmentWithJudge
	for _, judgment := range judgments {
		scorerJudgments = append(scorerJudgments, utils.JudgmentWithJudge{
			WinningTeamID: judgment.WinningTeamID,
			LosingTeamID:  judgment.LosingTeamID,
			JudgeID:       judgment.JudgeID,
		})
	}

	// Run the CrowdBT scorer
	scorer := utils.NewCrowdBTScorer()
	scorer.Score(scorerJudgments)

	// Build the response
	rankedTeams := scorer.RankTeams()
	teamScores := scorer.GetTeamScores()
	teamUncertainty := scorer.GetTeamUncertainty()
	judgeReliability := scorer.GetJudgeReliabilityAll()

	// Convert judge reliability maps to a cleaner format
	judgeStats := make(map[string]map[string]float64)
	for judgeID, params := range judgeReliability {
		judgeStats[judgeID] = map[string]float64{
			"alpha": params[0],
			"beta":  params[1],
		}
	}

	// Save top 3 finalists to settings
	if len(rankedTeams) >= 3 {
		finalist1 := &models.Setting{Name: models.SettingFinalist1, Value: rankedTeams[0]}
		finalist1.Save()
		finalist2 := &models.Setting{Name: models.SettingFinalist2, Value: rankedTeams[1]}
		finalist2.Save()
		finalist3 := &models.Setting{Name: models.SettingFinalist3, Value: rankedTeams[2]}
		finalist3.Save()
	}

	response := map[string]interface{}{
		"rankedTeams":      rankedTeams,
		"teamScores":       teamScores,
		"teamUncertainty":  teamUncertainty,
		"judgeReliability": judgeStats,
		"judgmentCount":    len(judgments),
	}

	return c.JSON(response)
}
