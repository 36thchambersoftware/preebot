package preeb

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"path/filepath"
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

// func LoadUser(userID string) User {
// 	collection := mongo.DB.Database("preebot").Collection("user")
// 	filter := bson.D{{"id", userID}}
// 	ctx := context.Background()
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	var user User
// 	err := collection.FindOne(ctx, filter).Decode(&user)
// 	if err != nil {
// 		log.Fatalf("cannot find user: %v", err)
// 	}

// 	return user
// }

func LoadUser(userID string) User {
	filename := filepath.Join(USER_FILE_PATH, userID+".json")
	// filename := "data/" + userID + ".json"
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("Cannot open user file: %v", err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Cannot get stats on user file: %v", err)
	}

	userJson := make([]byte, fileInfo.Size())
	n, err := file.Read(userJson)
	if err != nil {
		log.Fatalf("Cannot read user file: %v", err)
	}

	var userData User
	if n > 0 {
		err = json.Unmarshal(userJson, &userData)
		if err != nil {
			log.Fatalf("Cannot unmarshal user file: %v", err)
		}
	}

	if userData.Wallets == nil {
		userData.Wallets = make(Wallets)
	}

	return userData
}

// func (u User) Save() interface{} {
// 	collection := mongo.DB.Database("preebot").Collection("user")
// 	opts := options.Replace().SetUpsert(true)
// 	filter := bson.D{{"id", u.ID}}
// 	ctx := context.Background()
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	result, err := collection.ReplaceOne(ctx, filter, u, opts)
// 	if err != nil {
// 		log.Fatalf("cannot save user: %v", err)
// 	}

// 	return result.UpsertedID
// }

func SaveUser(user User) {
	filename := filepath.Join("data", user.ID+".json")
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("Cannot open user file: %v", err)
	}
	defer file.Close()

	userJson, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("Cannot marshal user: %v", err)
	}

	_, err = file.Write(userJson)
	if err != nil {
		log.Fatalf("Cannot write to user file: %v", err)
	}
}

// func LoadUsers() Users {
// 	if mongo.DB == nil {
// 		log.Println("Waiting for DB...")
// 		return nil
// 	}
// 	collection := mongo.DB.Database("preebot").Collection("user")
// 	filter := bson.D{}
// 	ctx := context.Background()
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	var users Users
// 	cursor, err := collection.Find(ctx, filter)
// 	if err != nil {
// 		log.Fatalf("cannot find users: %v", err)
// 	}

// 	for {
// 		if cursor.TryNext(context.TODO()) {
// 			var user User
// 			if err := cursor.Decode(&user); err != nil {
// 				log.Fatal(err)
// 			}

// 			users = append(users, user)

// 			continue
// 		}
// 		if err := cursor.Err(); err != nil {
// 			log.Fatal(err)
// 		}
// 		if cursor.ID() == 0 {
// 			break
// 		}
// 	}

// 	return users
// }

func LoadUsers() []User {
	files, err := os.ReadDir(USER_FILE_PATH)
	if err != nil {
		slog.Error("LOAD USERS", "could not load users", err)
		return nil
	}

	var users Users
	for _, file := range files {
		userPath := filepath.Join(USER_FILE_PATH, file.Name())
		fi, err := os.Stat(userPath)

		if err != nil || fi.Size() == 0 {
			slog.Error("LOAD USERS", "file info", "could not load user file, or file is empty", "file", file.Name(), "size", fi.Size(), "error", err)
			continue
		}

		if !file.IsDir() {
			contents, err := os.ReadFile(userPath)
			if err != nil {
				slog.Error("LOAD USERS", "could not load user file", err)
				return nil
			}

			var user User
			err = json.Unmarshal(contents, &user)
			if err != nil {
				slog.Error("LOAD USERS", "could not unmarshal user", err)
				return nil
			}

			users = append(users, user)
		}
	}

	return users
}
