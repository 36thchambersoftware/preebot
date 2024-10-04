package preebot

import (
	"encoding/json"
	"log"
	"os"
)

type User struct {
	ID          string   `json:"id,omitempty"`
	DisplayName string   `json:"display_name,omitempty"`
	Wallets     []string `json:"wallets,omitempty"`
}

func LoadUser(userID string) User {
	filename := userID + ".json"
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("Cannot open user file: %v", err)
	}
	defer file.Close()

	userJson := []byte{}
	n, err := file.Read(userJson)
	if err != nil {
		log.Fatalf("Cannot read user file: %v", err)
	}

	UserData := User{}
	if n > 0 {
		err = json.Unmarshal(userJson, &UserData)
		if err != nil {
			log.Fatalf("Cannot unmarshal user file: %v", err)
		}
	}

	return UserData
}

func SaveUser(user User) {
	filename := user.ID + ".json"
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("Cannot open user file: %v", err)
	}

	userJson := []byte{}
	userJson, err = json.Marshal(user)
	_, err = file.Write(userJson)
	if err != nil {
		log.Fatalf("Cannot write to user file: %v", err)
	}

	// User profile created
	file.Close()
}
