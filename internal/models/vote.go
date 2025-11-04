package models

import (
	"backend/internal/db"
	"backend/internal/utils"
	"crypto/rand"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Vote struct {
	ID           string    `bson:"id" json:"id"`
	Choice       string    `bson:"choice" json:"choice"`
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
	BucketedTime time.Time `bson:"bucketedTime" json:"bucketedTime"`
	Nonce        string    `bson:"nonce" json:"nonce"`
}

func (v *Vote) Create() error {
	v.ID = utils.GenID(6)
	v.CreatedAt = time.Now()
	v.BucketedTime = v.CreatedAt.Truncate(5 * time.Minute)

	// Generate random nonce
	nonceBuf := make([]byte, 8)
	if _, err := rand.Read(nonceBuf); err != nil {
		return err
	}
	v.Nonce = hex.EncodeToString(nonceBuf)

	_, err := db.Votes.InsertOne(db.Ctx, v)
	return err
}

func GetVoteCount(teamID string) (int64, error) {
	count, err := db.Votes.CountDocuments(db.Ctx, bson.M{"choice": teamID})
	return count, err
}

func GetAllVotes() ([]Vote, error) {
	cursor, err := db.Votes.Find(db.Ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(db.Ctx)

	var votes []Vote
	if err = cursor.All(db.Ctx, &votes); err != nil {
		return nil, err
	}

	return votes, nil
}

func GetVoteResults(finalists []string) (map[string]int64, error) {
	results := make(map[string]int64)

	for _, teamID := range finalists {
		count, err := GetVoteCount(teamID)
		if err != nil {
			return nil, err
		}
		results[teamID] = count
	}

	return results, nil
}
