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
	// Get team order setting (JSON string containing two base orders)
	teamOrderSetting := &Setting{Name: SettingTeamOrder}
	if err := teamOrderSetting.Get(); err != errmsg.EmptyStatusError {
		return "", err
	}

	var baseOrders [][]string
	if err := json.Unmarshal([]byte(teamOrderSetting.Value.(string)), &baseOrders); err != nil {
		return "", errmsg.InternalServerError(err)
	}

	if len(baseOrders) < 2 {
		return "", errmsg.InternalServerError(&errorMessage{message: "team order setting must contain two base orders"})
	}

	// Get judge offset setting (JSON string)
	judgeOffsetSetting := &Setting{Name: SettingJudgeOffset}
	if err := judgeOffsetSetting.Get(); err != errmsg.EmptyStatusError {
		return "", err
	}

	var judgeOffsets []int
	if err := json.Unmarshal([]byte(judgeOffsetSetting.Value.(string)), &judgeOffsets); err != nil {
		return "", errmsg.InternalServerError(err)
	}

	// Get judge multiplier setting (JSON string)
	judgeMultiplierSetting := &Setting{Name: SettingJudgeMultiplier}
	if err := judgeMultiplierSetting.Get(); err != errmsg.EmptyStatusError {
		return "", err
	}

	var judgeMultipliers []int
	if err := json.Unmarshal([]byte(judgeMultiplierSetting.Value.(string)), &judgeMultipliers); err != nil {
		return "", errmsg.InternalServerError(err)
	}

	// Get judge base order assignment setting (JSON string)
	judgeBaseOrderSetting := &Setting{Name: SettingJudgeBaseOrder}
	if err := judgeBaseOrderSetting.Get(); err != errmsg.EmptyStatusError {
		return "", err
	}

	var judgeBaseOrders []int
	if err := json.Unmarshal([]byte(judgeBaseOrderSetting.Value.(string)), &judgeBaseOrders); err != nil {
		return "", errmsg.InternalServerError(err)
	}

	// Get judge order setting to find this judge's index (JSON string)
	judgeOrderSetting := &Setting{Name: SettingJudgeOrder}
	if err := judgeOrderSetting.Get(); err != errmsg.EmptyStatusError {
		return "", err
	}

	var judgeOrder []string
	if err := json.Unmarshal([]byte(judgeOrderSetting.Value.(string)), &judgeOrder); err != nil {
		return "", errmsg.InternalServerError(err)
	}

	// Find this judge's index in judgeOrder
	judgeIndex := -1
	for i, jID := range judgeOrder {
		if jID == j.ID {
			judgeIndex = i
			break
		}
	}

	if judgeIndex == -1 {
		return "", errmsg.InternalServerError(&errorMessage{message: "judge not found in judge order"})
	}

	// Select which base order this judge uses (0 or 1)
	baseOrderIndex := judgeBaseOrders[judgeIndex]
	teamOrder := baseOrders[baseOrderIndex]

	numTeams := len(teamOrder)
	offset := judgeOffsets[judgeIndex]
	multiplier := judgeMultipliers[judgeIndex]

	// If CurrentTeam is -1, this is the first request - start at step 0
	if j.CurrentTeam == -1 {
		j.CurrentTeam = 0
	} else {
		// Check if we've completed all teams
		if j.CurrentTeam >= numTeams-1 {
			return "", errmsg.JudgingFinished
		}
		// Move to next step
		j.CurrentTeam++
	}

	// Calculate team index using coprime multiplier formula:
	// index = (offset + step * multiplier) % numTeams
	teamIndex := (offset + j.CurrentTeam*multiplier) % numTeams

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

	// Return the team ID at calculated position
	currentTeamID := teamOrder[teamIndex]

	return currentTeamID, errmsg.EmptyStatusError
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
