package preeb

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

type Configs []Config

type Config struct {
	PolicyRoles    PolicyRoles    `json:"policyroles,omitempty"`
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

type PolicyRoles map[string]RoleBounds
type DelegatorRoles map[string]RoleBounds

type RoleBounds struct {
	Min Bound
	Max Bound
}

func (drb RoleBounds) IsValid() bool {
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

	if config.PolicyRoles == nil {
		config.PolicyRoles = make(PolicyRoles)
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

func LoadConfigs() Configs {
	files, err := os.ReadDir(CONFIG_FILE_PATH)
	if err != nil {
		slog.Error("LOAD CONFIGS", "could not load configs", err)
		return nil
	}

	var configs Configs
	for _, file := range files {
		configPath := filepath.Join(CONFIG_FILE_PATH, file.Name())
		if !file.IsDir() {
			contents, err := os.ReadFile(configPath)
			if err != nil {
				slog.Error("LOAD CONFIGS", "could not load config file", err)
				return nil
			}

			var config Config
			err = json.Unmarshal(contents, &config)
			if err != nil {
				slog.Error("LOAD CONFIGS", "could not unmarshal config", err)
				return nil
			}

			configs = append(configs, config)
		}
	}

	return configs
}
