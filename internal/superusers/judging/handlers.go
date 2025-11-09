package judging

import (
	"backend/internal/db"
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
// @Router /superusers/judging/judge/connect [post]
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
// @Description Creates a new judge with an auto-generated ID and the given name. Optionally accepts a pair attribute for grouping judges.
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
		Name string `json:"name"`
		Pair string `json:"pair"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	judge := models.Judge{
		ID:          utils.GenID(6),
		Name:        body.Name,
		Pair:        body.Pair,
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

	// Save finalists to settings (up to 5)
	if len(rankedTeams) >= 1 {
		finalist1 := &models.Setting{Name: models.SettingFinalist1, Value: rankedTeams[0]}
		if serr := finalist1.Save(); serr != errmsg.EmptyStatusError {
			return utils.StatusError(c, serr)
		}
	}
	if len(rankedTeams) >= 2 {
		finalist2 := &models.Setting{Name: models.SettingFinalist2, Value: rankedTeams[1]}
		if serr := finalist2.Save(); serr != errmsg.EmptyStatusError {
			return utils.StatusError(c, serr)
		}
	}
	if len(rankedTeams) >= 3 {
		finalist3 := &models.Setting{Name: models.SettingFinalist3, Value: rankedTeams[2]}
		if serr := finalist3.Save(); serr != errmsg.EmptyStatusError {
			return utils.StatusError(c, serr)
		}
	}
	if len(rankedTeams) >= 4 {
		finalist4 := &models.Setting{Name: models.SettingFinalist4, Value: rankedTeams[3]}
		if serr := finalist4.Save(); serr != errmsg.EmptyStatusError {
			return utils.StatusError(c, serr)
		}
	}
	if len(rankedTeams) >= 5 {
		finalist5 := &models.Setting{Name: models.SettingFinalist5, Value: rankedTeams[4]}
		if serr := finalist5.Save(); serr != errmsg.EmptyStatusError {
			return utils.StatusError(c, serr)
		}
	}

	// Fetch full team details for each ranked team
	teamDetails := make(map[string]models.Team)
	for _, teamID := range rankedTeams {
		team := models.Team{ID: teamID}
		if err := team.Get(); err == nil {
			teamDetails[teamID] = team
		}
	}

	// Build finalists list in order from settings
	var finalists []models.Team
	finalistSettings := []string{
		models.SettingFinalist1,
		models.SettingFinalist2,
		models.SettingFinalist3,
		models.SettingFinalist4,
		models.SettingFinalist5,
	}

	for _, settingName := range finalistSettings {
		setting := &models.Setting{Name: settingName}
		if serr := setting.Get(); serr == errmsg.EmptyStatusError {
			if teamID, ok := setting.Value.(string); ok && teamID != "" {
				if team, exists := teamDetails[teamID]; exists {
					finalists = append(finalists, team)
				}
			}
		}
	}

	response := map[string]interface{}{
		"rankedTeams":      rankedTeams,
		"teamScores":       teamScores,
		"teamUncertainty":  teamUncertainty,
		"judgeReliability": judgeStats,
		"judgmentCount":    len(judgments),
		"teamDetails":      teamDetails,
		"finalists":        finalists,
	}

	return c.JSON(response)
}

// getFinalistsHandler retrieves the current finalists.
// @Summary Get current finalists
// @Description Returns the list of current finalist teams that have been saved after ranking computation.
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Produce json
// @Success 200 {object} GetFinalistsResponse
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judging/finalists [get]
func getFinalistsHandler(c fiber.Ctx) error {
	finalistSettings := []string{
		models.SettingFinalist1,
		models.SettingFinalist2,
		models.SettingFinalist3,
		models.SettingFinalist4,
		models.SettingFinalist5,
	}

	finalists := make([]models.Team, 0)

	for _, settingName := range finalistSettings {
		setting := &models.Setting{Name: settingName}
		if err := setting.Get(); err == errmsg.EmptyStatusError {
			if teamID, ok := setting.Value.(string); ok && teamID != "" {
				team := models.Team{ID: teamID}
				if err := team.Get(); err == nil {
					finalists = append(finalists, team)
				}
			}
		}
	}

	return c.JSON(bson.M{
		"finalists": finalists,
	})
}

// getAllJudgesHandler retrieves all judges with their current progress information.
// @Summary Get all judges with progress
// @Description Returns a list of all judges including their current team step and next available time.
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Produce json
// @Success 200 {array} JudgeProgressResponse
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judging/judges [get]
func getAllJudgesHandler(c fiber.Ctx) error {
	cursor, err := db.Judges.Find(db.Ctx, bson.M{})
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	defer cursor.Close(db.Ctx)

	var judges []models.Judge
	if err = cursor.All(db.Ctx, &judges); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	return c.JSON(judges)
}

// getVotingResultsHandler retrieves voting results with team details and vote counts.
// @Summary Get voting results
// @Description Returns all votes grouped by finalist team with full team details and vote counts.
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Produce json
// @Success 200 {array} VotingResultItem
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judging/voting-results [get]
func getVotingResultsHandler(c fiber.Ctx) error {
	// Get all votes from database
	allVotes, err := models.GetAllVotes()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	// Count votes by team
	voteCounts := make(map[string]int64)
	for _, vote := range allVotes {
		voteCounts[vote.Choice]++
	}

	// Build results with team details
	results := make([]map[string]interface{}, 0)

	for teamID, count := range voteCounts {
		team := models.Team{ID: teamID}
		if err := team.Get(); err != nil {
			// Skip teams that don't exist
			continue
		}

		result := map[string]interface{}{
			"team":  team,
			"count": count,
		}
		results = append(results, result)
	}

	return c.JSON(results)
}

// deleteJudgeHandler deletes a judge by ID.
// @Summary Delete a judge
// @Description Removes a judge from the system by their ID.
// @Tags Superusers Judging
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body JudgeDeleteRequest true "Judge ID to delete"
// @Success 200 {object} map[string]string
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 404 {object} errmsg._JudgeNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/judging/judge [delete]
func deleteJudgeHandler(c fiber.Ctx) error {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	if body.ID == "" {
		return utils.StatusError(c, errmsg.JudgeNotFound)
	}

	judge := models.Judge{ID: body.ID}
	err := judge.Delete()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	superuser := models.SuperUser{}
	utils.GetLocals(c, "superuser", &superuser)

	events.Em.JudgeDeleted(superuser.Username, judge.ID)

	return c.JSON(bson.M{
		"message": "judge deleted successfully",
		"id":      body.ID,
	})
}
