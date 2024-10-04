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

func init() {
	blockfrostProjectID, ok := os.LookupEnv("BLOCKFROST_PROJECT_ID")
	if !ok {
		slog.Error("Could not get blockfrost project id")
	}

	Client = bfg.NewAPIClient(bfg.APIClientOptions{ProjectID: blockfrostProjectID})
}

func GetLastTransaction(ctx context.Context, address string) bfg.TransactionUTXOs {
	slog.Info("GetLastTransaction", "address", address)
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
		slog.Info("GetLastTransaction", "has", hash)
	}

	txDetails, err := Client.TransactionUTXOs(ctx, hash)
	if err != nil {
		log.Fatalf("Could not get tx details: \nHASH: %v \nERROR: %v", hash, err)
	}

	log.Printf("txdetails: %+v", txDetails)

	return txDetails
}
