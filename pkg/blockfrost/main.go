package blockfrost

import (
	"context"
	"encoding/hex"
	"log"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"

	"preebot/pkg/preeb"

	bfg "github.com/blockfrost/blockfrost-go"
	"golang.org/x/exp/maps"
)

var (
	client         bfg.APIClient
	APIQueryParams bfg.APIQueryParams
	blockfrostProjectID string
)

const (
	LOVELACE             = 1_000_000
	ADA_HANDLE_PREFIX    = "$"
	ADA_HANDLE_POLICY_ID = "f0ff48bbb7bbe9d59a40f1ce90e9e9d0ff5002ec48f232b49ca0fb9a"
	CIP68v1_NONSENSE     = "000de140"
)

type (
	Lovelace int
	Ada      int
)

type AddressExtended struct {
	Address      string   `json:"address,omitempty"`
	Amount       []Amount `json:"amount,omitempty"`
	StakeAddress string   `json:"stake_address,omitempty"`
	Type         string   `json:"type,omitempty"`
	Script       bool     `json:"script,omitempty"`
}
type Amount struct {
	Unit                  string `json:"unit,omitempty"`
	Quantity              string `json:"quantity,omitempty"`
	Decimals              int    `json:"decimals,omitempty"`
	HasNftOnchainMetadata bool   `json:"has_nft_onchain_metadata,omitempty"`
}

func loadBlockfrostProjectID() string {
	blockfrostProjectID, ok := os.LookupEnv("BLOCKFROST_PROJECT_ID")
	if !ok {
		slog.Error("Could not get blockfrost project id")
	}

	return blockfrostProjectID
}

func init() {
	client = bfg.NewAPIClient(bfg.APIClientOptions{ProjectID: loadBlockfrostProjectID()})
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

	var account bfg.Account
	if stakeDetails.StakeAddress != nil {
		account = GetStakeInfo(ctx, *stakeDetails.StakeAddress)
	}

	return account
}

func GetStakeInfo(ctx context.Context, stakeAddress string) bfg.Account {
	stakeDetails, err := client.Account(ctx, stakeAddress)
	if err != nil {
		log.Printf("Could not get account details: \nSTAKEADDR: %v \nERROR: %v", stakeAddress, err)
	}

	return stakeDetails
}

func GetTotalStake(ctx context.Context, poolIDs preeb.PoolID, wallets preeb.Wallets) Ada {
	var totalStake int

	accounts := maps.Keys(wallets)
	for _, stakeAddress := range accounts {
		account := GetStakeInfo(ctx, string(stakeAddress))
		if account.Active && poolIDs[*account.PoolID] {
			stake, err := strconv.Atoi(account.ControlledAmount)
			if err != nil {
				log.Fatalf("Could not convert stake to int: \nSTAKE: %v \nERROR: %v", stake, err)
			}
			totalStake = totalStake + stake
		}
	}

	totalAda := totalStake / LOVELACE

	return Ada(totalAda)
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

func GetAllUserAddresses(ctx context.Context, wallets preeb.Wallets) ([]bfg.AddressExtended, error) {
	var allAddresses []bfg.AddressExtended
	for _, wallet := range wallets {
		for _, addr := range wallet {
			address, err := client.AddressExtended(ctx, string(addr))
			if err != nil {
				return nil, err
			}

			allAddresses = append(allAddresses, address)
		}
	}

	return allAddresses, nil
}

func CountUserAssetsByPolicy(ctx context.Context, policyIDs preeb.PolicyID, allAddresses []bfg.AddressExtended) int {
	totalNfts := 0

	powInt := func (decimals int) float64 {
		return math.Pow(10, float64(decimals))
	}

	for _, address := range allAddresses {
		for _, utxo := range address.Amount {
			for policyID := range policyIDs {
				if strings.HasPrefix(utxo.Unit, policyID) {
					qty, err := strconv.Atoi(utxo.Quantity)
					if err != nil {
						log.Printf("Could not get quantity from utxo: %v\n%v", utxo, err)
					}

					if utxo.Decimals != nil && *utxo.Decimals > 0 {
						qty = int(math.Floor(float64(qty) / powInt(*utxo.Decimals)))
					}

					totalNfts+= qty
				}
			}
		}
	}

	return totalNfts

	// {
	// 	"asset": "78dea0d35c9ac1f554066ab4491b0862c2482bdf617e0ba81414d51c000de140546972656c657373576f726b657230313033",
	// 	"policy_id": "78dea0d35c9ac1f554066ab4491b0862c2482bdf617e0ba81414d51c",
	// 	"asset_name": "000de140546972656c657373576f726b657230313033",
	// 	"fingerprint": "asset1w42x7zwpee4t28y8nzss0cteg6wahvkav7a8u2",
	// 	"quantity": "1",
	// 	"initial_mint_tx_hash": "a632fae151ba7c5748513f747db4f441336a5aa022d595c4ab826b1c6ed38a9c",
	// 	"mint_or_burn_count": 1,
	// 	"onchain_metadata": {
	// 	  "name": "Tireless Worker #0103",
	// 	  "mediaType": "image/jpeg",
	// 	  "image": "ipfs://bafybeib3lggegs6cpfx3hecc3l2umzewr4tgip725eu364qjbeqhihgj7y",
	// 	  "Rarity": "46436f6d6d6f6e",
	// 	  "Headgear": "494865616470686f6e65",
	// 	  "Background": "4d436f6d6d6f6e2059656c6c6f77",
	// 	  "Eyes": "47476c6173736573",
	// 	  "Chest": "4750656173616e74",
	// 	  "Tool": "465363726f6c6c",
	// 	  "Facial hair": "45506c61696e",
	// 	  "Speciality": "424f47",
	// 	  "description": "Main collection of the Necro League. "
	// 	},
	// 	"onchain_metadata_standard": "CIP68v1",
	// 	"onchain_metadata_extra": null,
	// 	"metadata": null
	//   }
}

// Convert ADA Handle address
func HandleAddress(ctx context.Context, addr string) (string, error) {
	isAdaHandle := strings.HasPrefix(addr, ADA_HANDLE_PREFIX)
	if isAdaHandle {
		hexAddr := hex.EncodeToString([]byte(addr[1:]))
		assetName := ADA_HANDLE_POLICY_ID + CIP68v1_NONSENSE + hexAddr
		addresses, err := client.AssetAddresses(ctx, assetName, APIQueryParams)
		if err != nil {
			return "", err
		}

		if len(addresses) > 0 {
			return addresses[0].Address, nil
		}

	}

	return addr, nil
}
