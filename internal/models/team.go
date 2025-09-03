package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
)

type Team struct {
	ID      string   `json:"id" bson:"id"`
	Name    string   `json:"name" bson:"name"`
	Members []string `json:"members" bson:"members"`
}

func (t *Team) Create(firstMember string) (err error) {
	t.ID = utils.GenID(6)
	t.Name = "New Team"
	t.Members = []string{
		firstMember,
	}

	_, err = db.Teams.InsertOne(db.Ctx, t)

	return err
}

func (t *Team) Get() (err error) {
	return db.Teams.FindOne(db.Ctx, bson.M{
		"id": t.ID,
	}).Decode(t)
}

func (t *Team) ChangeMembers(newMembers []string) (err error) {
	_, err = db.Teams.UpdateOne(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"members": newMembers,
		},
	})

	return
}

func (t *Team) AddMember(newMember string) (serr errmsg.StatusError) {
	if len(t.Members) == 4 {
		return errmsg.TeamFull
	}

	newMembers := append(t.Members, newMember)

	err := t.ChangeMembers(newMembers)
	if err != nil {
		return errmsg.InternalServerError
	}

	t.Members = newMembers

	return
}

func (t *Team) RemoveMember(removeMember string) (serr errmsg.StatusError) {
	newMembers := []string{}

	for _, v := range t.Members {
		if v != removeMember {
			newMembers = append(newMembers, v)
		}
	}

	err := t.ChangeMembers(newMembers)
	if err != nil {
		return errmsg.InternalServerError
	}

	t.Members = newMembers

	return
}

func (t *Team) Delete() (err error) {
	_, err = db.Teams.DeleteOne(db.Ctx, bson.M{
		"id": t.ID,
	})

	t = &Team{}

	return
}

func (t *Team) ChangeName(name string) (serr errmsg.StatusError) {
	err := db.Teams.FindOneAndUpdate(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"name": name,
		},
	}).Decode(t)

	if err != nil {
		return errmsg.InternalServerError
	}

	if t.Name == "" {
		return errmsg.TeamNotFound
	}

	t.Name = name

	return
}
