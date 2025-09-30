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
	if loadFlagStagesFromCache(&fstages) {
		return fstages, nil
	}

	return refreshFlagStagesCache()
}

func (fstage *FlagStage) Get() (serr errmsg.StatusError) {
	if loadFlagStageFromCache(fstage.ID, fstage) {
		return errmsg.EmptyStatusError
	}

	err := db.FlagStages.FindOne(db.Ctx, bson.M{
		"id": fstage.ID,
	}).Decode(fstage)

	if err != nil {
		return errmsg.FlagStageNotFound
	}

	cacheFlagStage(*fstage)

	return errmsg.EmptyStatusError
}

func (fstage *FlagStage) Create() (err error) {
	fstage.ID = utils.GenID(6)

	_, err = db.FlagStages.InsertOne(db.Ctx, fstage)
	if err != nil {
		return err
	}

	_, err = refreshFlagStagesCache()
	return err
}

func (fstage *FlagStage) Delete() (err error) {
	_, err = db.FlagStages.DeleteOne(db.Ctx, bson.M{
		"id": fstage.ID,
	})
	if err != nil {
		return
	}

	invalidateFlagStageCache(fstage.ID)
	_, err = refreshFlagStagesCache()
	if err != nil {
		return
	}

	*fstage = FlagStage{}

	return nil
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

func cacheFlagStage(stage FlagStage) {
	if stage.ID == "" {
		return
	}

	bytes, err := json.Marshal(stage)
	if err != nil {
		return
	}

	_ = db.CacheSetBytes(flagStageCacheKey(stage.ID), bytes)
}

func cacheFlagStageList(stages []FlagStage) {
	bytes, err := json.Marshal(stages)
	if err != nil {
		return
	}

	_ = db.CacheSetBytes(flagStageListCacheKey(), bytes)

	for _, stage := range stages {
		cacheFlagStage(stage)
	}
}

func loadFlagStagesFromCache(stages *[]FlagStage) bool {
	bytes, err := db.CacheGetBytes(flagStageListCacheKey())
	if err != nil || len(bytes) == 0 {
		return false
	}

	if err := json.Unmarshal(bytes, stages); err != nil {
		_ = db.CacheDel(flagStageListCacheKey())
		return false
	}

	return true
}

func loadFlagStageFromCache(id string, stage *FlagStage) bool {
	if id == "" {
		return false
	}

	bytes, err := db.CacheGetBytes(flagStageCacheKey(id))
	if err != nil || len(bytes) == 0 {
		return false
	}

	if err := json.Unmarshal(bytes, stage); err != nil {
		_ = db.CacheDel(flagStageCacheKey(id))
		return false
	}

	return stage.ID != ""
}

func refreshFlagStagesCache() (stages []FlagStage, err error) {
	stages = []FlagStage{}
	cursor := &mongo.Cursor{}
	cursor, err = db.FlagStages.Find(db.Ctx, bson.M{})
	if err != nil {
		return stages, err
	}
	defer cursor.Close(db.Ctx)

	if err = cursor.All(db.Ctx, &stages); err != nil {
		return stages, err
	}

	cacheFlagStageList(stages)

	return stages, nil
}

func invalidateFlagStageCache(id string) {
	if id != "" {
		_ = db.CacheDel(flagStageCacheKey(id))
	}
}

func flagStageCacheKey(id string) string {
	return "flagstage:" + id
}

func flagStageListCacheKey() string {
	return "flagstages"
}
