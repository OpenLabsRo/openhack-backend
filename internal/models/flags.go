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

	if err != nil {
		return err
	}

	f.Flags[flag] = value

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

	if err != nil {
		return err
	}

	newFlags := map[string]bool{}
	for k, v := range f.Flags {
		if k != flag {
			newFlags[k] = v
		}
	}
	f.Flags = newFlags

	return
}
