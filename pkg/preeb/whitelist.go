package preeb

import (
	"context"
	"log"
	mongo "preebot/pkg/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Whitelist struct {
	UserID string `bson:"user_id,omitempty"`
	StakeAddress StakeAddress `bson:"stake_address,omitempty"`
	Address Address `bson:"address,omitempty"`
	GuildID string `bson:"guild_id,omitempty"`
}

func CheckAddress(address string) error {
	collection := mongo.DB.Database("preebot").Collection("whitelist")
	filter := bson.D{{"address", address}}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var found Whitelist
	err := collection.FindOne(ctx, filter).Decode(&found)
	if err != nil {
		log.Printf("address not found in whitelist: %v", err)
		return err
	}

	return nil
}

func (w Whitelist) AddAddress() error {
	collection := mongo.DB.Database("preebot").Collection("whitelist")
	opts := options.Replace().SetUpsert(true)
	filter := bson.D{{Key: "user_id", Value: w.UserID}}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	_, err := collection.ReplaceOne(ctx, filter, w, opts)
	if err != nil {
		log.Printf("cannot save user: %v", err)
		return err
	}

	return nil
}