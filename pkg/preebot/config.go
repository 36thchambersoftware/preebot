package preebot

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	DelegatorRoles DelegatorRoles `json:"delegatorroles,omitempty"`
	PoolIDs        PoolID         `json:"poolids,omitempty"`
	PolicyIDs      PolicyID       `json:"policyids,omitempty"`
	EngageRole     string         `json:"engagerole,omitempty"`
	GuildID        string         `json:"guildid,omitempty"`
}

type (
	PoolID   map[string]bool
	PolicyID map[string]bool
)

type DelegatorRoles map[string]DelegatorRoleBounds

type DelegatorRoleBounds struct {
	Min Bound
	Max Bound
}

func (drb DelegatorRoleBounds) IsValid() bool {
	return drb.Max > drb.Min
}

type Bound int64

const (
	CONFIG_FILE_SUFFIX = "-config.json"
	CONFIG_FILE_PATH   = "config"
)

func LoadConfig(gID string) Config {
	filename := filepath.Join(CONFIG_FILE_PATH, gID+CONFIG_FILE_SUFFIX)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
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

	var config Config
	if n > 0 {
		err = json.Unmarshal(configJson, &config)
		if err != nil {
			log.Fatalf("Cannot unmarshal config file: %v", err)
		}
	}

	if config.GuildID == "" {
		config.GuildID = gID
	}

	if config.DelegatorRoles == nil {
		config.DelegatorRoles = make(DelegatorRoles)
	}

	if config.PoolIDs == nil {
		config.PoolIDs = make(PoolID)
	}

	if config.PolicyIDs == nil {
		config.PolicyIDs = make(PolicyID)
	}

	return config
}

func SaveConfig(config Config) {
	filename := filepath.Join(CONFIG_FILE_PATH, config.GuildID+CONFIG_FILE_SUFFIX)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
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
