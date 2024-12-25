package preeb

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"preebot/pkg/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Configs []Config

type Config struct {
	PolicyRoles    PolicyRoles    `bson:"policy_roles,omitempty"`
	DelegatorRoles DelegatorRoles `bson:"delegator_roles,omitempty"`
	PoolIDs        PoolID         `bson:"pool_ids,omitempty"`
	PolicyIDs      PolicyID       `bson:"policy_ids,omitempty"`
	EngageRole     string         `bson:"engage_role,omitempty"`
	GuildID        string         `bson:"guild_id,omitempty"`
	Custodians     []Custodian    `bson:"custodians,omitempty"`
}

type (
	PoolID   map[string]bool
	PolicyID map[string]bool
)

type PolicyRoles map[string]RoleBounds
type DelegatorRoles map[string]RoleBounds

type RoleBounds struct {
	Min Bound `bson:"min,omitempty"`
	Max Bound `bson:"max,omitempty"`
}

func (drb RoleBounds) IsValid() bool {
	return drb.Max > drb.Min
}

type Bound int64

type Custodian struct {
	Url url.URL `bson:"url,omitempty"`
	UserAddress string `bson:"user_address,omitempty"`
	CustodianAddress string `bson:"custodian_address,omitempty"`
}

const (
	CONFIG_FILE_SUFFIX = "-config.json"
	CONFIG_FILE_PATH   = "config"
)

// func LoadConfig(gID string) Config {
// 	filename := filepath.Join(CONFIG_FILE_PATH, gID+CONFIG_FILE_SUFFIX)
// 	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
// 	if err != nil {
// 		log.Fatalf("Cannot open config file: %v", err)
// 	}
// 	defer file.Close()
// 	fileInfo, err := file.Stat()
// 	if err != nil {
// 		log.Fatalf("Cannot get stats on config file: %v", err)
// 	}

// 	configJson := make([]byte, fileInfo.Size())
// 	n, err := file.Read(configJson)
// 	if err != nil {
// 		log.Fatalf("Cannot read config file: %v", err)
// 	}

// 	var config Config
// 	if n > 0 {
// 		err = json.Unmarshal(configJson, &config)
// 		if err != nil {
// 			log.Fatalf("Cannot unmarshal config file: %v", err)
// 		}
// 	}

// 	if config.GuildID == "" {
// 		config.GuildID = gID
// 	}

// 	if config.DelegatorRoles == nil {
// 		config.DelegatorRoles = make(DelegatorRoles)
// 	}

// 	if config.PoolIDs == nil {
// 		config.PoolIDs = make(PoolID)
// 	}

// 	if config.PolicyRoles == nil {
// 		config.PolicyRoles = make(PolicyRoles)
// 	}

// 	if config.PolicyIDs == nil {
// 		config.PolicyIDs = make(PolicyID)
// 	}

// 	return config
// }

// func SaveConfig(config Config) {
// 	filename := filepath.Join(CONFIG_FILE_PATH, config.GuildID+CONFIG_FILE_SUFFIX)
// 	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o644)
// 	if err != nil {
// 		log.Fatalf("Cannot open config file: %v", err)
// 	}
// 	defer file.Close()

// 	configJson, err := json.Marshal(config)
// 	if err != nil {
// 		log.Fatalf("Cannot marshal config: %v", err)
// 	}

// 	_, err = file.Write(configJson)
// 	if err != nil {
// 		log.Fatalf("Cannot write to config file: %v", err)
// 	}
// }

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

func (c Config) Save() interface{} {
	collection := mongo.DB.Database("preebot").Collection("config")
	opts := options.Replace().SetUpsert(true)
	filter := bson.D{{"guild_id", c.GuildID}}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	result, err := collection.ReplaceOne(ctx, filter, c, opts)
	if err != nil {
		log.Fatalf("cannot save config: %v", err)
	}

	return result.UpsertedID
}

func LoadConfig(guild_id string) Config {
	collection := mongo.DB.Database("preebot").Collection("config")
	filter := bson.D{{"guild_id", guild_id}}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var config Config
	err := collection.FindOne(ctx, filter).Decode(&config)
	if err != nil {
		log.Fatalf("cannot find config: %v", err)
	}

	return config
}

// func LoadConfigs() Configs {
// 	collection := mongo.DB.Database("preebot").Collection("config")
// 	filter := bson.D{}
// 	ctx := context.Background()
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	var configs Configs
// 	cursor, err := collection.Find(ctx, filter)
// 	if err != nil {
// 		log.Fatalf("cannot find configs: %v", err)
// 	}

// 	for {
// 		if cursor.TryNext(context.TODO()) {
// 			var config Config
// 			if err := cursor.Decode(&config); err != nil {
// 				log.Fatal(err)
// 			}

// 			configs = append(configs, config)

// 			continue
// 		}
// 		if err := cursor.Err(); err != nil {
// 			log.Fatal(err)
// 		}
// 		if cursor.ID() == 0 {
// 			break
// 		}
// 	}

// 	return configs
// }