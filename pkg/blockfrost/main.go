package blockfrost

import (
	"context"
	"log"
	"log/slog"
	"os"

	bfg "github.com/blockfrost/blockfrost-go"
)

var (
	Client         bfg.APIClient
	APIQueryParams bfg.APIQueryParams
)

const (
	LOVELACE      = 1_000_000
	PREEB_POOL_ID = "pool19peeq2czwunkwe3s70yuvwpsrqcyndlqnxvt67usz98px57z7fk"
)

func init() {
	blockfrostProjectID, ok := os.LookupEnv("BLOCKFROST_PROJECT_ID")
	if !ok {
		slog.Error("Could not get blockfrost project id")
	}

	Client = bfg.NewAPIClient(bfg.APIClientOptions{ProjectID: blockfrostProjectID})
}

func GetLastTransaction(ctx context.Context, address string) bfg.TransactionUTXOs {
	APIQueryParams.Order = "desc"
	txs, err := Client.AddressTransactions(ctx, address, APIQueryParams)
	if err != nil {
		if err != nil {
			log.Fatalf("Could not get txs for address: \nADDRESS: %v \nERROR: %v", address, err)
		}
	}

	var hash string
	if len(txs) > 0 {
		hash = txs[0].TxHash
	}

	txDetails, err := Client.TransactionUTXOs(ctx, hash)
	if err != nil {
		log.Fatalf("Could not get tx details: \nHASH: %v \nERROR: %v", hash, err)
	}

	return txDetails
}

func GetStakeInfo(ctx context.Context, address string) bfg.Account {
	addressDetails, err := Client.Address(ctx, address)
	if err != nil {
		log.Fatalf("Could not get address details: \nHASH: %v \nERROR: %v", address, err)
	}

	stakeDetails, err := Client.Account(ctx, *addressDetails.StakeAddress)
	if err != nil {
		log.Fatalf("Could not get account details: \nHASH: %v \nERROR: %v", address, err)
	}

	return stakeDetails
}
