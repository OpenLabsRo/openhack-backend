package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
)

type Tag struct {
	ID        string `json:"id" bson:"id"`
	AccountID string `json:"accountID" bson:"accountID"`
}

func (t *Tag) Get() (serr errmsg.StatusError) {
	if t.ID == "" {
		return errmsg.TagIncomplete
	}

	cache, _ := db.CacheGet("badge:" + t.ID)
	if cache == "" {
		err := db.Tags.
			FindOne(db.Ctx,
				bson.M{
					"id": t.ID,
				}).
			Decode(&t)

		if err != nil {
			return errmsg.TagNotFound
		}

		if t.ID == "" {
			return errmsg.TagNotFound
		}

		bytes, _ := json.Marshal(t)
		db.CacheSetBytes("badge:"+t.ID, bytes)

		return
	}

	err := json.Unmarshal([]byte(cache), &t)
	if err != nil {
		return errmsg.InternalServerError(err)
	}
	return errmsg.EmptyStatusError
}

func (t *Tag) Assign() (serr errmsg.StatusError) {
	if t.AccountID == "" || t.ID == "" {
		return errmsg.TagIncomplete
	}

	_, err := db.Tags.InsertOne(db.Ctx, t)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	bytes, err := json.Marshal(t)
	err = db.CacheSetBytes("badge:"+t.ID, bytes)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	return errmsg.EmptyStatusError
}
