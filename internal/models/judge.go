package models

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/utils"
	"encoding/json"
	"strings"
	"time"

	sj "github.com/brianvoe/sjwt"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

type Judge struct {
	ID          string `bson:"id" json:"id"`
	Name        string `bson:"name" json:"name"`
	CurrentTeam int    `bson:"currentTeam" json:"currentTeam"`
	Pair        string `bson:"pair" json:"pair"`
}

func (j *Judge) IssueJudgeConnectToken() (token string) {
	claims, _ := sj.ToClaims(j)
	claims.SetExpiresAt(time.Now().Add(2 * time.Minute))

	token = claims.Generate(env.JWT_SECRET)
	return token
}

func (j *Judge) GenToken() string {
	claims, _ := sj.ToClaims(j)
	claims.SetExpiresAt(time.Now().Add(24 * time.Hour))

	token := claims.Generate(env.JWT_SECRET)
	return token
}

func (j *Judge) ParseToken(token string) (err error) {
	hasVerified := sj.Verify(token, env.JWT_SECRET)

	if !hasVerified {
		return nil
	}

	claims, _ := sj.Parse(token)
	err = claims.Validate()
	claims.ToStruct(&j)

	return
}

func (j *Judge) Initialize() (serr errmsg.StatusError) {
	err := j.GetByName()

	if err == nil {
		return errmsg.JudgeAlreadyExists
	}

	_, err = db.Judges.InsertOne(db.Ctx, j)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	return
}

func (j *Judge) Get() (err error) {
	err = db.Judges.FindOne(db.Ctx, bson.M{
		"id": j.ID,
	}).Decode(&j)

	return
}

func (j *Judge) GetByName() (err error) {
	err = db.Judges.FindOne(db.Ctx, bson.M{
		"name": j.Name,
	}).Decode(&j)

	return
}

func (j *Judge) Delete() (err error) {
	_, err = db.Judges.DeleteOne(db.Ctx, bson.M{
		"id": j.ID,
	})

	return
}

func (j *Judge) GetNextTeam() (teamID string, serr errmsg.StatusError) {
	// Fetch the initialization matrix
	matrixSetting := &Setting{Name: SettingJudgeInitMatrix}
	if err := matrixSetting.Get(); err != errmsg.EmptyStatusError {
		return "", err
	}

	var matrixObj struct {
		Steps  int        `json:"steps"`
		Groups int        `json:"groups"`
		Matrix [][]string `json:"matrix"`
	}
	if err := json.Unmarshal([]byte(matrixSetting.Value.(string)), &matrixObj); err != nil {
		return "", errmsg.InternalServerError(err)
	}

	numSteps := matrixObj.Steps
	numGroups := matrixObj.Groups
	matrix := matrixObj.Matrix

	if numSteps == 0 {
		return "", errmsg.InternalServerError(&errorMessage{message: "no steps available"})
	}

	// Get judge to group index mapping
	judgeToGroupIndexSetting := &Setting{Name: SettingJudgeToGroupIndex}
	if err := judgeToGroupIndexSetting.Get(); err != errmsg.EmptyStatusError {
		return "", err
	}

	var judgeToGroupIdx map[string]int
	if err := json.Unmarshal([]byte(judgeToGroupIndexSetting.Value.(string)), &judgeToGroupIdx); err != nil {
		return "", errmsg.InternalServerError(err)
	}

	// Get this judge's pair group index
	groupIdx, exists := judgeToGroupIdx[j.ID]
	if !exists {
		return "", errmsg.InternalServerError(&errorMessage{message: "judge not found in pair group mapping"})
	}

	if groupIdx < 0 || groupIdx >= numGroups {
		return "", errmsg.InternalServerError(&errorMessage{message: "invalid pair group index"})
	}

	// Initialize on first call: -1 becomes 0
	currentStep := j.CurrentTeam
	if currentStep == -1 {
		currentStep = 0
		j.CurrentTeam = 0
	}

	// Check if we've exhausted all steps
	if currentStep >= numSteps {
		return "", errmsg.JudgingFinished
	}

	// Read the assignment for this step from the matrix
	// If blank, return empty string (rest), but don't skip forward
	assignedTeamID := ""
	if len(matrix) > currentStep && len(matrix[currentStep]) > groupIdx {
		assignedTeamID = matrix[currentStep][groupIdx]
	}

	// Increment step for next call (whether this step was blank or not)
	j.CurrentTeam = currentStep + 1

	// Update judge's CurrentTeam in database
	_, err := db.Judges.UpdateOne(db.Ctx, bson.M{
		"id": j.ID,
	}, bson.M{
		"$set": bson.M{
			"currentTeam": j.CurrentTeam,
		},
	})
	if err != nil {
		return "", errmsg.InternalServerError(err)
	}

	return assignedTeamID, errmsg.EmptyStatusError
}

// errorMessage is a simple error wrapper for the InternalServerError function
type errorMessage struct {
	message string
}

func (e *errorMessage) Error() string {
	return e.message
}

func JudgeMiddleware(c fiber.Ctx) error {
	var token string

	authHeader := c.Get("Authorization")

	if string(authHeader) != "" &&
		strings.HasPrefix(string(authHeader), "Bearer") {

		tokens := strings.Fields(string(authHeader))
		if len(tokens) == 2 {
			token = tokens[1]
		}

		if token == "" {
			return utils.StatusError(c,
				errmsg.AccountNoToken,
			)
		}

		var judge Judge
		err := judge.ParseToken(token)
		if err != nil {
			return utils.StatusError(
				c, errmsg.AccountNoToken,
			)
		}

		if judge.ID == "" {
			return utils.StatusError(
				c, errmsg.AccountNoToken,
			)
		}

		// Fetch current judge state from database to get updated CurrentTeam
		err = judge.Get()
		if err != nil {
			return utils.StatusError(
				c, errmsg.AccountNoToken,
			)
		}

		c.Locals("id", judge.ID)
		utils.SetLocals(c, "judge", judge)
	}

	if token == "" {
		return utils.StatusError(c,
			errmsg.AccountNoToken,
		)
	}

	return c.Next()
}
