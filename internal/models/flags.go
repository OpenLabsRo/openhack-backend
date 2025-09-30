package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/utils"
	"encoding/json"
	"maps"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

type Flags struct {
	Flags map[string]bool `json:"flags" bson:"flags"`
	Stage FlagStage       `json:"stage" bson:"stage"`
}

func FlagsMiddlewareBuilder(flags []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		f := Flags{}
		err := f.Get()

		if err != nil {
			return utils.StatusError(
				c, errmsg.InternalServerError(err),
			)
		}

		for _, flagName := range flags {
			if !f.Flags[flagName] {
				return utils.StatusError(c, errmsg.FlagRequired)
			}
		}

		return c.Next()
	}
}

func (f *Flags) Get() (err error) {
	if loadFlagsFromCache(f) {
		return nil
	}

	err = db.Flags.
		FindOne(db.Ctx, bson.M{}).
		Decode(&f)

	if err != nil {
		return err
	}

	if f.Flags == nil {
		f.Flags = map[string]bool{}
	}

	cacheFlags(f)

	return nil
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

	if f.Flags == nil {
		f.Flags = map[string]bool{}
	}
	f.Flags[flag] = value

	cacheFlags(f)

	return
}

func (f *Flags) SetBulk(rawFlags map[string]bool) (err error) {
	marshaledFlags := bson.M{}

	for k, v := range rawFlags {
		marshaledFlags["flags."+k] = v
	}

	_, err = db.Flags.UpdateOne(db.Ctx, bson.M{},
		bson.M{
			"$set": marshaledFlags,
		},
	)

	if err != nil {
		return err
	}

	if f.Flags == nil {
		f.Flags = map[string]bool{}
	}
	maps.Copy(f.Flags, rawFlags)

	cacheFlags(f)

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

	cacheFlags(f)

	return
}

func (f *Flags) Reset() (err error) {
	resetFlags := map[string]bool{}

	for f := range f.Flags {
		resetFlags[f] = false
	}

	err = f.SetBulk(resetFlags)
	if err != nil {
		return
	}

	f.Flags = resetFlags
	return
}

func cacheFlags(f *Flags) {
	if f == nil {
		return
	}

	bytes, err := json.Marshal(f)
	if err != nil {
		return
	}

	_ = db.CacheSetBytes(flagsCacheKey(), bytes)
}

func loadFlagsFromCache(f *Flags) bool {
	bytes, err := db.CacheGetBytes(flagsCacheKey())
	if err != nil || len(bytes) == 0 {
		return false
	}

	if err := json.Unmarshal(bytes, f); err != nil {
		_ = db.CacheDel(flagsCacheKey())
		return false
	}

	return true
}

func flagsCacheKey() string {
	return "flags"
}

// this will not update the cache,
// as it will always be executed when changing a stage,
// which automatically does a SetBulk on the flags
func (f *Flags) SetStage(flagStage FlagStage) (err error) {
	_, err = db.Flags.UpdateOne(db.Ctx, bson.M{},
		bson.M{
			"$set": bson.M{
				"stage": flagStage,
			},
		},
	)

	if err != nil {
		return err
	}

	f.Stage = flagStage

	return
}
