package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/utils"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FlagStage struct {
	ID      string   `json:"id" bson:"id"`
	Name    string   `json:"name" bson:"name"`
	TurnOff []string `json:"turnoff" bson:"turnoff"`
	TurnOn  []string `json:"turnon" bson:"turnon"`
}

func GetFlagStages() (fstages []FlagStage, err error) {
	fstages = []FlagStage{}
	cache, _ := db.CacheGet("flagstages")

	if cache == "" {
		cursor := &mongo.Cursor{}
		cursor, err = db.FlagStages.Find(db.Ctx, bson.M{})
		if err != nil {
			return
		}
		defer cursor.Close(db.Ctx)

		if err = cursor.All(db.Ctx, &fstages); err != nil {
			return
		}

		bytes := []byte{}
		bytes, err = json.Marshal(fstages)
		if err != nil {
			return
		}
		err = db.CacheSetBytes("flagstages", bytes)
		if err != nil {
			return
		}

		// now setting the cache for each and every flagstage
		for _, v := range fstages {
			bytes, err = json.Marshal(v)
			if err != nil {
				return
			}
			err = db.CacheSetBytes("flagstage:"+v.ID, bytes)
			if err != nil {
				return
			}
		}

		return
	}

	err = json.Unmarshal([]byte(cache), &fstages)

	return
}

func (fstage *FlagStage) Get() (serr errmsg.StatusError) {
	cache, _ := db.CacheGet("flagstage:" + fstage.ID)

	if cache == "" {
		err := db.FlagStages.FindOne(db.Ctx, bson.M{
			"id": fstage.ID,
		}).Decode(fstage)

		if err != nil {
			return errmsg.FlagStageNotFound
		}

		bytes := []byte{}
		bytes, err = json.Marshal(fstage)
		if err != nil {
			return
		}
		err = db.CacheSetBytes("flagstage:"+fstage.ID, bytes)
		if err != nil {
			return
		}
	}

	return errmsg.EmptyStatusError
}

func (fstage *FlagStage) Create() (err error) {
	fstage.ID = utils.GenID(6)

	_, err = db.FlagStages.InsertOne(db.Ctx, fstage)

	// --- updating the cache
	// -- first the flagstage itself
	bytes, err := json.Marshal(fstage)
	if err != nil {
		return
	}
	err = db.CacheSetBytes("flagstage:"+fstage.ID, bytes)
	if err != nil {
		return
	}

	// -- and then the flagstages

	// getting the flagStages before
	flagStages := []FlagStage{}
	fstagesBytes, _ := db.CacheGetBytes("flagstages")
	err = json.Unmarshal(fstagesBytes, &flagStages)
	if err != nil {
		return
	}

	// -- changing the flagStages and marshaling
	flagStages = append(flagStages, *fstage)
	fstagesBytes, err = json.Marshal(flagStages)
	if err != nil {
		return
	}

	// -- writing them to the cache
	err = db.CacheSetBytes("flagstages", fstagesBytes)
	if err != nil {
		return
	}

	return
}

func (fstage *FlagStage) Delete() (err error) {
	_, err = db.FlagStages.DeleteOne(db.Ctx, bson.M{
		"id": fstage.ID,
	})

	// ---  updating the cache
	// -- first deleting the flagstage itself
	err = db.CacheDel("flagstage:" + fstage.ID)
	if err != nil {
		return
	}

	// -- getting the previous flagstages
	flagStages := []FlagStage{}
	fstagesBytes, _ := db.CacheGetBytes("flagstages")
	err = json.Unmarshal(fstagesBytes, &flagStages)
	if err != nil {
		return
	}

	// -- changing the flagStages and marshaling
	newFlagStages := []FlagStage{}
	for _, v := range flagStages {
		if v.ID != fstage.ID {
			newFlagStages = append(newFlagStages, v)
		}
	}
	fstagesBytes, err = json.Marshal(newFlagStages)
	if err != nil {
		return
	}

	// -- writing them to the cache
	err = db.CacheSetBytes("flagstages", fstagesBytes)
	if err != nil {
		return
	}

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
