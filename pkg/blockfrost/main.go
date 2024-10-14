package blockfrost

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strconv"

	"preebot/pkg/preebot"

	bfg "github.com/blockfrost/blockfrost-go"
	"golang.org/x/exp/maps"
)

var (
	client         bfg.APIClient
	APIQueryParams bfg.APIQueryParams
)

const (
	LOVELACE = 1_000_000
)

func init() {
	blockfrostProjectID, ok := os.LookupEnv("BLOCKFROST_PROJECT_ID")
	if !ok {
		slog.Error("Could not get blockfrost project id")
	}

	client = bfg.NewAPIClient(bfg.APIClientOptions{ProjectID: blockfrostProjectID})
}

func GetLastTransaction(ctx context.Context, address string) (bfg.TransactionUTXOs, error) {
	APIQueryParams.Order = "desc"
	txs, err := client.AddressTransactions(ctx, address, APIQueryParams)
	if err != nil {
		log.Printf("Could not get txs for address: \nADDRESS: %v \nERROR: %v", address, err)
		return bfg.TransactionUTXOs{}, err
	}

	var hash string
	if len(txs) > 0 {
		hash = txs[0].TxHash
	}

	txDetails, err := client.TransactionUTXOs(ctx, hash)
	if err != nil {
		log.Printf("Could not get tx details: \nHASH: %v \nERROR: %v", hash, err)
	}

	return txDetails, nil
}

func GetAccountByAddress(ctx context.Context, address string) bfg.Account {
	stakeDetails, err := client.Address(ctx, address)
	if err != nil {
		log.Printf("Could not get account details: \aADDRESS: %v \nERROR: %v", address, err)
	}

	account := GetStakeInfo(ctx, *stakeDetails.StakeAddress)

	return account
}

func GetStakeInfo(ctx context.Context, stakeAddress string) bfg.Account {
	stakeDetails, err := client.Account(ctx, stakeAddress)
	if err != nil {
		log.Printf("Could not get account details: \nSTAKEADDR: %v \nERROR: %v", stakeAddress, err)
	}

	return stakeDetails
}

func GetTotalStake(ctx context.Context, wallets preebot.Wallets) int {
	config := preebot.LoadConfig()
	var totalStake int

	accounts := maps.Keys(wallets)
	for _, stakeAddress := range accounts {
		account := GetStakeInfo(ctx, string(stakeAddress))
		if account.Active && config.PoolID[*account.PoolID] {
			stake, err := strconv.Atoi(account.ControlledAmount)
			if err != nil {
				log.Fatalf("Could not convert stake to int: \nSTAKE: %v \nERROR: %v", stake, err)
			}
			totalStake = totalStake + stake
		}
	}

	return totalStake
}

func GetPoolMetaData(ctx context.Context, poolID string) (bfg.PoolMetadata, error) {
	metaData, err := client.PoolMetadata(ctx, poolID)
	if err != nil {
		return bfg.PoolMetadata{}, err
	}

	return metaData, nil
}

func GetPolicyAssets(ctx context.Context, policyID string) ([]bfg.AssetByPolicy, error) {
	assets, err := client.AssetsByPolicy(ctx, policyID)
	if err != nil {
		return []bfg.AssetByPolicy{}, err
	}

	return assets, nil
}

func CountUserAssetsByPolicy(ctx context.Context, policyIDs preebot.PolicyID, wallets []string) (int, error) {
	var totalNfts int
	var allAddresses []bfg.Address
	for _, wallet := range wallets {
		address, err := client.Address(ctx, wallet)
		if err != nil {
			return 0, err
		}

		allAddresses = append(allAddresses, address)
	}

	return totalNfts, nil
}
