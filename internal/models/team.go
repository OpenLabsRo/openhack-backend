package models

import (
	"backend/internal/db"
	"backend/internal/errmsg"
	"backend/internal/utils"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Team struct {
	ID      string   `json:"id" bson:"id"`
	Name    string   `json:"name" bson:"name"`
	Members []string `json:"members" bson:"members"`

	Submission struct {
		Name string `json:"name" bson:"name"`
		Desc string `json:"desc" bson:"desc"`
		Repo string `json:"repo" bson:"repo"`
		Pres string `json:"pres" bson:"pres"`
	} `json:"submission" bson:"submission"`

	Table string `json:"table" bson:"table"`

	Deleted bool `json:"deleted" bson:"deleted"`
}

func (t *Team) Create(firstMember string) (err error) {
	t.ID = utils.GenTeamID()
	t.Name = "New Team"
	t.Members = []string{
		firstMember,
	}
	t.Deleted = false

	_, err = db.Teams.InsertOne(db.Ctx, t)
	if err != nil {
		return
	}

	cacheTeam(t)
	invalidateTeamMembersCache(t.ID)

	return nil
}

func (t *Team) Get() (err error) {
	if loadTeamFromCache(t.ID, t) {
		return nil
	}

	err = db.Teams.FindOne(db.Ctx, bson.M{
		"id": t.ID,
	}).Decode(t)
	if err != nil {
		return err
	}

	cacheTeam(t)

	return nil
}

func (t *Team) GetMembers() (members []Account, err error) {
	members = []Account{}
	if loadTeamMembersFromCache(t.ID, &members) {
		return members, nil
	}

	cursor := &mongo.Cursor{}
	cursor, err = db.Accounts.Find(db.Ctx, bson.M{
		"id": bson.M{
			"$in": t.Members,
		},
	})
	if err != nil {
		return members, err
	}

	if err = cursor.All(db.Ctx, &members); err != nil {
		return members, err
	}

	cacheTeamMembers(t.ID, members)

	return members, nil
}

func (t *Team) ChangeMembers(newMembers []string) (err error) {
	_, err = db.Teams.UpdateOne(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"members": newMembers,
		},
	})
	if err != nil {
		return
	}

	t.Members = newMembers

	cacheTeam(t)
	invalidateTeamMembersCache(t.ID)

	return nil
}

func (t *Team) AddMember(newMember string, newFullMember Account) (serr errmsg.StatusError) {
	if len(t.Members) == 4 {
		return errmsg.TeamFull
	}

	newMembers := append(t.Members, newMember)

	err := t.ChangeMembers(newMembers)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	newFullMember.TeamID = t.ID

	return errmsg.EmptyStatusError
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
		return errmsg.InternalServerError(err)
	}

	return errmsg.EmptyStatusError
}

func (t *Team) Delete() (oldID string, err error) {
	_, err = db.Teams.UpdateOne(db.Ctx,
		bson.M{
			"id": t.ID,
		},
		bson.M{
			"$set": bson.M{
				"deleted": true,
			},
		},
	)
	if err != nil {
		return
	}

	invalidateTeamMembersCache(t.ID)
	invalidateTeamCache(t.ID)

	oldID = t.ID
	t = &Team{}

	return
}

func (t *Team) ChangeName(name string) (oldName string, serr errmsg.StatusError) {
	err := db.Teams.FindOneAndUpdate(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"name": name,
		},
	}).Decode(t)

	if err != nil {
		return "", errmsg.InternalServerError(err)
	}

	if t.Name == "" {
		return "", errmsg.TeamNotFound
	}

	oldName = t.Name

	t.Name = name

	cacheTeam(t)

	return
}

func (t *Team) ChangeTable(table string) (oldTable string, serr errmsg.StatusError) {
	err := db.Teams.FindOneAndUpdate(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"table": table,
		},
	}).Decode(t)

	if err != nil {
		return "", errmsg.InternalServerError(err)
	}

	if t.Name == "" {
		return "", errmsg.TeamNotFound
	}

	oldTable = t.Table

	t.Table = table

	cacheTeam(t)

	return
}

func (t *Team) ChangeSubmissionName(name string) (oldName string, serr errmsg.StatusError) {
	err := db.Teams.FindOneAndUpdate(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"submission.name": name,
		},
	}).Decode(t)

	if err != nil {
		return oldName, errmsg.InternalServerError(err)
	}

	oldName = t.Submission.Name
	t.Submission.Name = name

	cacheTeam(t)

	return
}

func (t *Team) ChangeSubmissionDesc(desc string) (oldDesc string, serr errmsg.StatusError) {
	err := db.Teams.FindOneAndUpdate(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"submission.desc": desc,
		},
	}).Decode(t)

	if err != nil {
		return oldDesc, errmsg.InternalServerError(err)
	}

	oldDesc = t.Submission.Desc
	t.Submission.Desc = desc

	cacheTeam(t)

	return
}

func (t *Team) ChangeSubmissionRepo(repo string) (oldRepo string, serr errmsg.StatusError) {
	err := db.Teams.FindOneAndUpdate(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"submission.repo": repo,
		},
	}).Decode(t)

	if err != nil {
		return oldRepo, errmsg.InternalServerError(err)
	}

	oldRepo = t.Submission.Repo
	t.Submission.Repo = repo

	cacheTeam(t)

	return
}
func (t *Team) ChangeSubmissionPres(pres string) (oldPres string, serr errmsg.StatusError) {
	err := db.Teams.FindOneAndUpdate(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"submission.pres": pres,
		},
	}).Decode(t)

	if err != nil {
		return oldPres, errmsg.InternalServerError(err)
	}

	oldPres = t.Submission.Pres
	t.Submission.Pres = pres

	cacheTeam(t)

	return
}

func cacheTeam(t *Team) {
	if t == nil || t.ID == "" {
		return
	}

	bytes, err := json.Marshal(t)
	if err != nil {
		return
	}

	_ = db.CacheSetBytes(teamCacheKey(t.ID), bytes)
}

func loadTeamFromCache(id string, team *Team) bool {
	if id == "" {
		return false
	}

	bytes, err := db.CacheGetBytes(teamCacheKey(id))
	if err != nil || len(bytes) == 0 {
		return false
	}

	if err := json.Unmarshal(bytes, team); err != nil {
		_ = db.CacheDel(teamCacheKey(id))
		return false
	}

	return team.ID != ""
}

func invalidateTeamCache(id string) {
	if id != "" {
		_ = db.CacheDel(teamCacheKey(id))
	}
}

func cacheTeamMembers(teamID string, members []Account) {
	if teamID == "" {
		return
	}

	bytes, err := json.Marshal(members)
	if err != nil {
		return
	}

	_ = db.CacheSetBytes(teamMembersCacheKey(teamID), bytes)
}

func loadTeamMembersFromCache(teamID string, members *[]Account) bool {
	if teamID == "" {
		return false
	}

	bytes, err := db.CacheGetBytes(teamMembersCacheKey(teamID))
	if err != nil || len(bytes) == 0 {
		return false
	}

	if err := json.Unmarshal(bytes, members); err != nil {
		_ = db.CacheDel(teamMembersCacheKey(teamID))
		return false
	}

	return true
}

func invalidateTeamMembersCache(teamID string) {
	if teamID != "" {
		_ = db.CacheDel(teamMembersCacheKey(teamID))
	}
}

func teamCacheKey(id string) string {
	return "team:" + id
}

func teamMembersCacheKey(teamID string) string {
	return "members:" + teamID
}
