package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var SettingBadgePileSalt = "badgePileSalt"
var SettingJudgeToGroupIndex = "judgeToGroupIndex"
var SettingJudgeInitMatrix = "judgeInitMatrix"
var SettingFinalist1 = "finalist_1"
var SettingFinalist2 = "finalist_2"
var SettingFinalist3 = "finalist_3"
var SettingFinalist4 = "finalist_4"
var SettingFinalist5 = "finalist_5"
var SettingWaitMinutes = "waitMinutes"

type Setting struct {
	Name  string `json:"name" bson:"name"`
	Value any    `json:"value" bson:"value"`
}

func (s *Setting) Get() errmsg.StatusError {
	if s.Name == "" {
		return errmsg.SettingIncomplete
	}

	if loadSettingFromCache(s.Name, s) {
		return errmsg.EmptyStatusError
	}

	err := db.Settings.FindOne(
		db.Ctx,
		bson.M{"name": s.Name},
	).Decode(s)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errmsg.SettingNotFound
		}

		return errmsg.InternalServerError(err)
	}

	cacheSetting(*s)

	return errmsg.EmptyStatusError
}

func (s *Setting) Update() errmsg.StatusError {
	if s.Name == "" {
		return errmsg.SettingIncomplete
	}

	update := bson.M{
		"value": s.Value,
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := db.Settings.FindOneAndUpdate(
		db.Ctx,
		bson.M{"name": s.Name},
		bson.M{"$set": update},
		opts,
	).Decode(s)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errmsg.SettingNotFound
		}

		return errmsg.InternalServerError(err)
	}

	cacheSetting(*s)

	return errmsg.EmptyStatusError
}

func (s *Setting) Delete() errmsg.StatusError {
	if s.Name == "" {
		return errmsg.SettingIncomplete
	}

	res, err := db.Settings.DeleteOne(
		db.Ctx,
		bson.M{"name": s.Name},
	)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	if res.DeletedCount == 0 {
		return errmsg.SettingNotFound
	}

	invalidateSettingCache(s.Name)

	return errmsg.EmptyStatusError
}

func (s *Setting) Save() errmsg.StatusError {
	if s.Name == "" {
		return errmsg.SettingIncomplete
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetUpsert(true)

	err := db.Settings.FindOneAndUpdate(
		db.Ctx,
		bson.M{"name": s.Name},
		bson.M{
			"$set": bson.M{
				"value": s.Value,
			},
		},
		opts,
	).Decode(s)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	cacheSetting(*s)

	return errmsg.EmptyStatusError
}

func cacheSetting(setting Setting) {
	if setting.Name == "" {
		return
	}

	bytes, err := json.Marshal(setting)
	if err != nil {
		return
	}

	_ = db.CacheSetBytes(settingCacheKey(setting.Name), bytes)
}

func loadSettingFromCache(name string, setting *Setting) bool {
	if name == "" {
		return false
	}

	bytes, err := db.CacheGetBytes(settingCacheKey(name))
	if err != nil || len(bytes) == 0 {
		return false
	}

	if err := json.Unmarshal(bytes, setting); err != nil {
		_ = db.CacheDel(settingCacheKey(name))
		return false
	}

	return setting.Name != ""
}

func invalidateSettingCache(name string) {
	if name == "" {
		return
	}

	_ = db.CacheDel(settingCacheKey(name))
}

func settingCacheKey(name string) string {
	return "setting:" + name
}
