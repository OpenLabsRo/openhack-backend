package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
)

type Badge struct {
	ID        string `json:"id" bson:"id"`
	AccountID string `json:"accountID" bson:"accountID"`
}

func (b *Badge) Get() (serr errmsg.StatusError) {
	if b.ID == "" {
		return errmsg.BadgeIncomplete
	}

	cache, _ := db.CacheGet("badge:" + b.ID)
	if cache == "" {
		err := db.Badges.
			FindOne(db.Ctx,
				bson.M{
					"id": b.ID,
				}).
			Decode(&b)

		if err != nil {
			return errmsg.InternalServerError(err)
		}

		if b.ID == "" {
			return errmsg.BadgeNotFound
		}

		bytes, _ := json.Marshal(b)
		db.CacheSetBytes("badge:"+b.ID, bytes)

		return
	}

	err := json.Unmarshal([]byte(cache), &b)
	if err != nil {
		return errmsg.InternalServerError(err)
	}
	return errmsg.EmptyStatusError
}

func (b *Badge) Assign() (serr errmsg.StatusError) {
	if b.AccountID == "" || b.ID == "" {
		return errmsg.BadgeIncomplete
	}

	_, err := db.Badges.InsertOne(db.Ctx, b)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	bytes, err := json.Marshal(b)
	err = db.CacheSetBytes("badge:"+b.ID, bytes)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	return errmsg.EmptyStatusError
}
