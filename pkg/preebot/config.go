package preebot

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	PoolID string `json:"poolid,omitempty"`
}

const CONFIG_FILE = "config.json"

func LoadConfig() Config {
	file, err := os.OpenFile(CONFIG_FILE, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("Cannot open config file: %v", err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Cannot get stats on config file: %v", err)
	}

	configJson := make([]byte, fileInfo.Size())
	n, err := file.Read(configJson)
	if err != nil {
		log.Fatalf("Cannot read config file: %v", err)
	}

	config := Config{}
	if n > 0 {
		err = json.Unmarshal(configJson, &config)
		if err != nil {
			log.Fatalf("Cannot unmarshal config file: %v", err)
		}
	}

	return config
}

func SaveConfig(config Config) {
	file, err := os.OpenFile(CONFIG_FILE, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("Cannot open config file: %v", err)
	}
	defer file.Close()

	configJson, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Cannot marshal config: %v", err)
	}

	_, err = file.Write(configJson)
	if err != nil {
		log.Fatalf("Cannot write to config file: %v", err)
	}
}
