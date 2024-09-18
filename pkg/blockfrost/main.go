package blockfrost

import (
	"log/slog"
	"os"

	bfg "github.com/blockfrost/blockfrost-go"
)

var bfc bfg.APIClient

func init() {
	blockfrostProjectID, ok := os.LookupEnv("BLOCKFROST_PROJECT_ID")
	if !ok {
		slog.Error("Could not get blockfrost project id")
	}

	bfc = bfg.NewAPIClient(bfg.APIClientOptions{ProjectID: blockfrostProjectID})
}
