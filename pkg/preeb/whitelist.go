package preeb

import (
	"context"
	mongo "preebot/pkg/db"

	"go.mongodb.org/mongo-driver/bson"
)

type Whitelist struct {
	UserID string `json:"id,omitempty"`
	StakeAddress StakeAddress `json:"stakeAddress,omitempty"`
}

func CheckAddress(stakeAddress string) error {
	collection := mongo.DB.Database("preebot").Collection("whitelist")
	filter := bson.D{{"stakeAddress", stakeAddress}}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var found Whitelist
	err := collection.FindOne(ctx, filter).Decode(&found)
	if err != nil {
		return err
	}

	return nil
}