package models

import (
	"backend/internal/db"
	"backend/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
)

type FlagStage struct {
	ID      string   `json:"id" bson:"id"`
	Name    string   `json:"name" bson:"name"`
	TurnOff []string `json:"turnoff" bson:"turnoff"`
	TurnOn  []string `json:"turnon" bson:"turnon"`
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

func ExecuteFlagStage() (err error) {
	return nil
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
