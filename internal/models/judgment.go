package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Judgment struct {
	ID            string    `bson:"id" json:"id"`
	WinningTeamID string    `bson:"winningTeamID" json:"winningTeamID"`
	LosingTeamID  string    `bson:"losingTeamID" json:"losingTeamID"`
	Date          time.Time `bson:"date" json:"date"`
	JudgeID       string    `bson:"judgeID" json:"judgeID"`
}

func (j *Judgment) Create() (err error) {
	j.ID = utils.GenID(6)
	j.Date = time.Now()

	_, err = db.Judgments.InsertOne(db.Ctx, j)
	if err != nil {
		return err
	}

	return nil
}

func (j *Judgment) Get() (err error) {
	err = db.Judgments.FindOne(db.Ctx, bson.M{
		"id": j.ID,
	}).Decode(&j)

	return err
}

func GetAllJudgments() (judgments []Judgment, serr errmsg.StatusError) {
	cursor, err := db.Judgments.Find(db.Ctx, bson.M{})
	if err != nil {
		return nil, errmsg.InternalServerError(err)
	}

	if err = cursor.All(db.Ctx, &judgments); err != nil {
		return nil, errmsg.InternalServerError(err)
	}

	return judgments, errmsg.EmptyStatusError
}
