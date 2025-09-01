package models

import (
	"backend/internal/db"

	"go.mongodb.org/mongo-driver/bson"
)

type Flags struct {
	Flags map[string]bool `json:"flags" bson:"flags"`
}

func (f *Flags) Get() (err error) {
	return db.Flags.FindOne(db.Ctx, bson.M{}).Decode(&f)
}

func (f *Flags) Set(flag string, value bool) (err error) {
	_, err = db.Flags.UpdateOne(db.Ctx, bson.M{},
		bson.M{
			"$set": bson.M{
				"flags." + flag: value,
			},
		},
	)

	return
}

func (f *Flags) Unset(flag string) (err error) {
	_, err = db.Flags.UpdateOne(db.Ctx, bson.M{},
		bson.M{
			"$unset": bson.M{
				"flags." + flag: "",
			},
		},
	)

	return
}
