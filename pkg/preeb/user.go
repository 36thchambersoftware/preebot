package preeb

import (
	"context"
	"log"
	mongo "preebot/pkg/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Users []User

type User struct {
	ID          string  `json:"id,omitempty"`
	DisplayName string  `json:"display_name,omitempty"`
	Wallets     Wallets `json:"wallets,omitempty"`
}

type (
	Wallets      map[StakeAddress][]Address
	StakeAddress string
	Address      string
)

func (a Address) String() string {
	return string(a)
}

const USER_FILE_PATH = "data"

func LoadUser(userID string) User {
	collection := mongo.DB.Database("preebot").Collection("user")
	filter := bson.D{{"id", userID}}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var user User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Fatalf("cannot find user: %v", err)
	}

	return user
}

func (u User) Save() interface{} {
	collection := mongo.DB.Database("preebot").Collection("user")
	opts := options.Replace().SetUpsert(true)
	filter := bson.D{{"id", u.ID}}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	result, err := collection.ReplaceOne(ctx, filter, u, opts)
	if err != nil {
		log.Fatalf("cannot save user: %v", err)
	}

	return result.UpsertedID
}

func LoadUsers() Users {
	if mongo.DB == nil {
		log.Println("Waiting for DB...")
		return nil
	}
	collection := mongo.DB.Database("preebot").Collection("user")
	filter := bson.D{}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var users Users
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatalf("cannot find users: %v", err)
	}

	for {
		if cursor.TryNext(context.TODO()) {
			var user User
			if err := cursor.Decode(&user); err != nil {
				log.Fatal(err)
			}

			users = append(users, user)

			continue
		}
		if err := cursor.Err(); err != nil {
			log.Fatal(err)
		}
		if cursor.ID() == 0 {
			break
		}
	}

	return users
}