package blockfrost

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"preebot/pkg/logger"
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

func GetAddressTransactions(ctx context.Context, address string) ([]bfg.AddressTransactions, error) {
	APIQueryParams.Order = "desc"
	txs, err := client.AddressTransactions(ctx, address, APIQueryParams)
	if err != nil {
		log.Printf("Could not get txs for address: \nADDRESS: %v \nERROR: %v", address, err)
		return nil, err
	}

	return txs, nil
}

func GetTransaction(ctx context.Context, hash string) (bfg.TransactionUTXOs, error) {
	tx, err := client.TransactionUTXOs(ctx, hash)
	if err != nil {
		log.Printf("Could not get txs for address: \nHASH: %v \nERROR: %v", hash, err)
		return bfg.TransactionUTXOs{}, err
	}

	return tx, nil
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


func GetAddress(ctx context.Context, address string) bfg.Address {
	addr, err := client.Address(ctx, address)
	if err != nil {
		log.Printf("Could not get account details: \aADDRESS: %v \nERROR: %v", address, err)
	}

	return addr
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
	for _, addr := range wallets {
		address, err := client.AddressExtended(ctx, addr.String())
		if err != nil {
			return nil, err
		}

		allAddresses = append(allAddresses, address)
	}

	return allAddresses, nil
}

func CountUserAssetsByPolicy(ctx context.Context, policyIDs preeb.PolicyIDs, allAddresses []bfg.AddressExtended) map[string]int {
	var policyCounts = make(map[string]int)

	powInt := func (decimals int) float64 {
		return math.Pow(10, float64(decimals))
	}

	for _, address := range allAddresses {
		total := 0
		for _, utxo := range address.Amount {
			for policyID, policy := range policyIDs {
				if !HasAllGroupedPolicies(policy, allAddresses) {
					continue // Skip this policy if grouping not satisfied
				}

				if strings.HasPrefix(utxo.Unit, policyID) {
					qty, err := strconv.Atoi(utxo.Quantity)
					if err != nil {
						log.Printf("Could not get quantity from utxo: %v\n%v", utxo, err)
					}

					if utxo.Decimals != nil && *utxo.Decimals > 0 {
						qty = int(math.Floor(float64(qty) / powInt(*utxo.Decimals)))
					}

					// ✅ Trait matching logic for NFTs
					if len(policy.Traits) > 0 && utxo.HasNftOnchainMetadata {
						asset, err := client.Asset(ctx, utxo.Unit)
						if err != nil {
							log.Printf("Could not get asset details: \nUNIT: %v \nERROR: %v", utxo.Unit, err)
							continue // Skip this UTXO if asset details cannot be retrieved
						}
						slog.Info("asset", "METADATA", *asset.OnchainMetadata, "POLICY", policy.Traits)
						if !HasMatchingTrait(asset.OnchainMetadata, policy.Traits) {
							continue // Skip this UTXO if no trait matches
						}
					}

					total+= qty
					policyCounts[policyID] += total
				}
			}
		}
	}

	return policyCounts
}

func HasMatchingTrait(metadata *interface{}, requiredTraits map[string][]string) bool {
	// Check if metadata is non-nil and of the right type
	if metadata == nil {
		return false
	}

	// Assert to map[string]interface{} (standard for onchain JSON)
	metaMap, ok := (*metadata).(map[string]interface{})
	if !ok {
		return false
	}

	// Loop through required traits
	for traitKey, allowedValues := range requiredTraits {
		if val, exists := metaMap[traitKey]; exists {
			valStr := fmt.Sprintf("%v", val)
			for _, requiredVal := range allowedValues {
				if valStr == requiredVal {
					return true
				}
			}
		}
	}

	return false
}



func HasAllGroupedPolicies(policy preeb.Policy, userAddresses []bfg.AddressExtended) bool {
	if len(policy.GroupWith) == 0 {
		return true // No grouped policies required
	}

	// Track which required policies the user has
	held := make(map[string]bool)

	// For each address and its UTXOs
	for _, address := range userAddresses {
		for _, utxo := range address.Amount {
			// Check if UTXO matches any policy in GroupWith
			for requiredPolicyID := range policy.GroupWith {
				if strings.HasPrefix(utxo.Unit, requiredPolicyID) {
					held[requiredPolicyID] = true
				}
			}
		}
	}

	// Ensure user holds something from each required group policy
	for requiredPolicyID := range policy.GroupWith {
		if !held[requiredPolicyID] {
			return false
		}
	}

	return true
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

func EpochsDelegatedToPool(ctx context.Context, stakeAddress string, poolID string) (*int, error) {
	var epoch int
	l := logger.Record.WithGroup("EpochsDelegatedToPool")
	APIQueryParams.Order = "desc"
	history, err := client.AccountDelegationHistory(ctx, stakeAddress, APIQueryParams)
	if err != nil {
		l.Error("could not get account history", "ERROR", err)
		return nil, err
	}

	var stakeHistory []bfg.AccountDelegationHistory

	if len(history) > 0 {
		latestPool := history[0].PoolID
		if latestPool == poolID || poolID == "" {
			for i := 0; i < len(history); i++ {
				if history[i].PoolID == latestPool {
					stakeHistory = append(stakeHistory, history[i])
				} else {
					break
				}
			}
		}
	} else {
		l.Info("no history", "HISTORY", history)
	}

	if len(stakeHistory) > 1 {
		sort.Slice(stakeHistory, func(i, j int) bool {
			return  stakeHistory[i].ActiveEpoch < stakeHistory[j].ActiveEpoch
		})

		epoch = int(stakeHistory[0].ActiveEpoch)
	} else if len(stakeHistory) == 1 {
		epoch = int(stakeHistory[0].ActiveEpoch)
	} else {
		l.Info("no stake", "STAKEHISTORY", stakeHistory)
	}

	return &epoch, nil
}

func PoolInfo(ctx context.Context, poolID string) (*bfg.Pool, error) {
	info, err := client.Pool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func PoolHistory(ctx context.Context, poolID string) ([]bfg.PoolHistory, error) {
	APIQueryParams.Order = "desc"
	history, err := client.PoolHistory(ctx, poolID, APIQueryParams)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func PoolMeta(ctx context.Context, poolID string) (*bfg.PoolMetadata, error) {
	info, err := client.PoolMetadata(ctx, poolID)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func PoolBlocks(ctx context.Context, poolID string) (bfg.PoolBlocks, error) {
	APIQueryParams.Order = "desc"
	blocks, err := client.PoolBlocks(ctx, poolID, APIQueryParams)
	if err != nil {
		return nil, err
	}

	return blocks, nil
}