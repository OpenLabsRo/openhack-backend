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

	if loadTagFromCache(t.ID, t) {
		return errmsg.EmptyStatusError
	}

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

	cacheTag(*t)

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

	cacheTag(*t)

	return errmsg.EmptyStatusError
}

func cacheTag(tag Tag) {
	if tag.ID == "" {
		return
	}

	bytes, err := json.Marshal(tag)
	if err != nil {
		return
	}

	_ = db.CacheSetBytes(tagCacheKey(tag.ID), bytes)
}

func loadTagFromCache(id string, tag *Tag) bool {
	if id == "" {
		return false
	}

	bytes, err := db.CacheGetBytes(tagCacheKey(id))
	if err != nil || len(bytes) == 0 {
		return false
	}

	if err := json.Unmarshal(bytes, tag); err != nil {
		_ = db.CacheDel(tagCacheKey(id))
		return false
	}

	return tag.ID != ""
}

func tagCacheKey(id string) string {
	return "badge:" + id
}
