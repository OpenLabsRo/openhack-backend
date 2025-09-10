package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
)

type FlagStage struct {
	ID      string   `json:"id" bson:"id"`
	Name    string   `json:"name" bson:"name"`
	TurnOff []string `json:"turnoff" bson:"turnoff"`
	TurnOn  []string `json:"turnon" bson:"turnon"`
}

func (fstage *FlagStage) Get() (serr errmsg.StatusError) {
	err := db.FlagStages.FindOne(db.Ctx, bson.M{
		"id": fstage.ID,
	}).Decode(fstage)

	if err != nil {
		return errmsg.FlagStageNotFound
	}

	return errmsg.EmptyStatusError
}

func (fstage *FlagStage) Create() (err error) {
	fstage.ID = utils.GenID(6)

	_, err = db.FlagStages.InsertOne(db.Ctx, fstage)

	return
}

func (fstage *FlagStage) Delete() (err error) {
	_, err = db.FlagStages.DeleteOne(db.Ctx, bson.M{
		"id": fstage.ID,
	})

	*fstage = FlagStage{}

	return
}

func (fstage *FlagStage) Execute() (err error) {
	// combine the two lists
	instructions := map[string]bool{}
	for _, v := range fstage.TurnOff {
		instructions[v] = false
	}

	for _, v := range fstage.TurnOn {
		instructions[v] = true
	}

	flags := Flags{}
	err = flags.Get()
	if err != nil {
		return
	}

	err = flags.SetStage(*fstage)
	if err != nil {
		return
	}

	err = flags.SetBulk(instructions)
	if err != nil {
		return
	}

	return
}

func GetFlagStages() (fstages []FlagStage, err error) {
	cursor, err := db.FlagStages.Find(db.Ctx, bson.M{})
	if err != nil {
		return
	}
	defer cursor.Close(db.Ctx)

	if err = cursor.All(db.Ctx, &fstages); err != nil {
		return
	}

	return
}
