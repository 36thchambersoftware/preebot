package preeb

import (
	"context"
	"log"
	"net/url"
	mongo "preebot/pkg/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Configs []Config

type (
	Config struct {
		DelegatorRoles  DelegatorRoles `bson:"delegator_roles,omitempty"`
		PoolIDs         PoolID         `bson:"pool_ids,omitempty"`
		PoolChannelID   string         `bson:"pool_channel_id,omitempty"`
		PolicyIDs       PolicyIDs      `bson:"policy_ids,omitempty"`
		EngageRole      string         `bson:"engage_role,omitempty"`
		GuildID         string         `bson:"guild_id,omitempty"`
		Custodians      []Custodian    `bson:"custodians,omitempty"`
	}

	PoolID   map[string]bool
	DelegatorRoles map[string]RoleBounds
)

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
		log.Printf("cannot find config: %v", err)
	}

	if config.GuildID == "" {
		config.GuildID = guild_id
	}

	if config.PoolIDs == nil {
		config.PoolIDs = make(PoolID)
	}

	if config.PolicyIDs == nil {
		config.PolicyIDs = make(PolicyIDs)
	}

	if config.DelegatorRoles == nil {
		config.DelegatorRoles = make(DelegatorRoles)
	}

	return config
}

func LoadConfigs() Configs {
	if mongo.DB == nil {
		log.Println("Waiting for DB...")
		return nil
	}
	collection := mongo.DB.Database("preebot").Collection("config")
	filter := bson.D{}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var configs Configs
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatalf("cannot find configs: %v", err)
	}

	for {
		if cursor.TryNext(context.TODO()) {
			var config Config
			if err := cursor.Decode(&config); err != nil {
				log.Fatal(err)
			}

			configs = append(configs, config)

			continue
		}
		if err := cursor.Err(); err != nil {
			log.Fatal(err)
		}
		if cursor.ID() == 0 {
			break
		}
	}

	return configs
}