package preebot

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

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

func LoadUser(userID string) User {
	filename := filepath.Join("data", userID+".json")
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
