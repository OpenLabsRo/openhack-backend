package models

import (
	"backend/internal/db"
	"backend/internal/utils"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

type Team struct {
	ID      string   `json:"id" bson:"id"`
	Name    string   `json:"name" bson:"name"`
	Members []string `json:"members" bson:"members"`
}

func (t *Team) Create(firstMember string) (err error) {
	t.ID = utils.GenID(6)
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

func (t *Team) AddMember(newMember string) (err error) {
	if len(t.Members) == 4 {
		return errors.New("maximum size of a team reached")
	}

	exists := false
	for _, v := range t.Members {
		if v == newMember {
			exists = true
		}
	}

	if exists {
		return errors.New("cannot add to team - teammate already exists ")
	}

	newMembers := append(t.Members, newMember)

	err = t.ChangeMembers(newMembers)
	if err != nil {
		return
	}

	t.Members = newMembers

	return
}

func (t *Team) RemoveMember(removeMember string) (err error) {
	newMembers := []string{}

	for _, v := range t.Members {
		if v != removeMember {
			newMembers = append(newMembers, v)
		}
	}

	if len(t.Members) == len(newMembers) {
		return errors.New("could not remove from team - teammate does not exist ")
	}

	err = t.ChangeMembers(newMembers)
	if err != nil {
		return
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

func (t *Team) ChangeName(name string) (err error) {
	_, err = db.Teams.UpdateOne(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"name": name,
		},
	})

	if err != nil {
		return
	}

	t.Name = name

	return
}
