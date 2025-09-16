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

	// caching the members
	t.GetMembers()

	// caching the tam
	tBytes, err := json.Marshal(t)
	return db.CacheSetBytes("team"+t.ID, tBytes)
}

func (t *Team) Get() (err error) {
	cache, _ := db.CacheGet("team:" + t.ID)
	if cache == "" {
		err = db.Teams.FindOne(db.Ctx, bson.M{
			"id": t.ID,
		}).Decode(t)
		if err != nil {
			return
		}

		tBytes := []byte{}
		tBytes, err = json.Marshal(t)
		if err != nil {
			return
		}

		err = db.CacheSetBytes("team:"+t.ID, tBytes)
		if err != nil {
			return
		}

		return
	}

	err = json.Unmarshal([]byte(cache), t)

	return
}

func (t *Team) GetMembers() (members []Account, err error) {
	members = []Account{}

	cache, _ := db.CacheGet("members:" + t.ID)
	if cache == "" {
		cursor := &mongo.Cursor{}
		cursor, err = db.Accounts.Find(db.Ctx, bson.M{
			"id": bson.M{
				"$in": t.Members,
			},
		})
		if err != nil {
			return
		}

		if err = cursor.All(db.Ctx, &members); err != nil {
			return
		}

		membersBytes := []byte{}
		membersBytes, err = json.Marshal(members)
		if err != nil {
			return
		}

		err = db.CacheSetBytes("members:"+t.ID, membersBytes)
		if err != nil {
			return
		}

		return
	}

	err = json.Unmarshal([]byte(cache), &members)
	if err != nil {
		return
	}

	return
}

func (t *Team) ChangeMembers(newMembers []string) (err error) {
	_, err = db.Teams.UpdateOne(db.Ctx, bson.M{
		"id": t.ID,
	}, bson.M{
		"$set": bson.M{
			"members": newMembers,
		},
	})

	t.Members = newMembers

	// --- updating the team cache
	tBytes, err := json.Marshal(t)
	if err != nil {
		return
	}
	err = db.CacheSetBytes("team:"+t.ID, tBytes)
	if err != nil {
		return
	}

	return
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

	// --- updating the cache
	// -- get the old members
	fullMembers, err := t.GetMembers()
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	// -- add the new member to the old fullMembers and marshal it
	newFullMember.TeamID = t.ID
	fullMembers = append(fullMembers, newFullMember)
	fullMembersBytes, err := json.Marshal(fullMembers)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	// -- changing the cache
	err = db.CacheSetBytes("members:"+t.ID, fullMembersBytes)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

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
		return errmsg.InternalServerError(err)
	}

	// --- updating the cache
	// -- get the old members
	fullMembers, err := t.GetMembers()
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	// -- add the new member to the old fullMembers and marshal it
	newFullMembers := []Account{}
	for _, v := range fullMembers {
		if v.ID != removeMember {
			newFullMembers = append(newFullMembers, v)
		}
	}
	fullMembersBytes, err := json.Marshal(newFullMembers)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	// -- changing the cache
	err = db.CacheSetBytes("members:"+t.ID, fullMembersBytes)
	if err != nil {
		return errmsg.InternalServerError(err)
	}

	return
}

func (t *Team) Delete() (oldID string, err error) {
	_, err = db.Teams.UpdateOne(db.Ctx,
		bson.M{
			"id": t.ID,
		},
		bson.M{
			"deleted": true,
		},
	)

	err = db.CacheDel("members:" + t.ID)
	if err != nil {
		return
	}
	err = db.CacheDel("team:" + t.ID)
	if err != nil {
		return
	}

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

	tBytes, err := json.Marshal(t)
	if err != nil {
		return "", errmsg.InternalServerError(err)
	}
	err = db.CacheSetBytes("team:"+t.ID, tBytes)
	if err != nil {
		return "", errmsg.InternalServerError(err)
	}

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

	tBytes, err := json.Marshal(t)
	if err != nil {
		return oldName, errmsg.InternalServerError(err)
	}
	err = db.CacheSetBytes("team:"+t.ID, tBytes)
	if err != nil {
		return oldName, errmsg.InternalServerError(err)
	}

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

	tBytes, err := json.Marshal(t)
	if err != nil {
		return oldDesc, errmsg.InternalServerError(err)
	}
	err = db.CacheSetBytes("team:"+t.ID, tBytes)
	if err != nil {
		return oldDesc, errmsg.InternalServerError(err)
	}

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

	tBytes, err := json.Marshal(t)
	if err != nil {
		return oldRepo, errmsg.InternalServerError(err)
	}
	err = db.CacheSetBytes("team:"+t.ID, tBytes)
	if err != nil {
		return oldRepo, errmsg.InternalServerError(err)
	}

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

	tBytes, err := json.Marshal(t)
	if err != nil {
		return oldPres, errmsg.InternalServerError(err)
	}
	err = db.CacheSetBytes("team:"+t.ID, tBytes)
	if err != nil {
		return oldPres, errmsg.InternalServerError(err)
	}

	return
}
