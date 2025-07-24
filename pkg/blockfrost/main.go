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
	logger.Record.Info("BLOCKFROST", "CALL", "GetLastTransaction")
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
	logger.Record.Info("BLOCKFROST", "CALL", "GetAddressTransactions")
	APIQueryParams.Order = "desc"
	txs, err := client.AddressTransactions(ctx, address, APIQueryParams)
	if err != nil {
		log.Printf("Could not get txs for address: \nADDRESS: %v \nERROR: %v", address, err)
		return nil, err
	}

	return txs, nil
}

func GetTransaction(ctx context.Context, hash string) (bfg.TransactionUTXOs, error) {
	logger.Record.Info("BLOCKFROST", "CALL", "GetTransaction")
	tx, err := client.TransactionUTXOs(ctx, hash)
	if err != nil {
		log.Printf("Could not get txs for address: \nHASH: %v \nERROR: %v", hash, err)
		return bfg.TransactionUTXOs{}, err
	}

	return tx, nil
}

func GetAccountByAddress(ctx context.Context, address string) bfg.Account {
	logger.Record.Info("BLOCKFROST", "CALL", "GetAccountByAddress")
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
	logger.Record.Info("BLOCKFROST", "CALL", "GetAddress")
	addr, err := client.Address(ctx, address)
	if err != nil {
		log.Printf("Could not get account details: \aADDRESS: %v \nERROR: %v", address, err)
	}

	return addr
}

func GetStakeInfo(ctx context.Context, stakeAddress string) bfg.Account {
	logger.Record.Info("BLOCKFROST", "CALL", "GetStakeInfo")
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
		logger.Record.Info("BLOCKFROST", "CALL", "GetTotalStake")
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
	logger.Record.Info("BLOCKFROST", "CALL", "GetPoolMetaData")
	metaData, err := client.PoolMetadata(ctx, poolID)
	if err != nil {
		return bfg.PoolMetadata{}, err
	}

	return metaData, nil
}

func GetPolicyAssets(ctx context.Context, policyID string) ([]bfg.AssetByPolicy, error) {
	logger.Record.Info("BLOCKFROST", "CALL", "GetPolicyAssets")
	assets, err := client.AssetsByPolicy(ctx, policyID)
	if err != nil {
		return []bfg.AssetByPolicy{}, err
	}

	return assets, nil
}

// Get all the assets from all addresses for a single stake address
func GetAllUserAddressesAssets(ctx context.Context, stake preeb.StakeAddress, page uint) ([]bfg.AccountAssociatedAsset, error) {
	logger.Record.Info("BLOCKFROST", "CALL", "GetAllUserAddressesAssets")
	APIQueryParams.Page = int(page)
	assets, err := client.AccountAssociatedAssets(ctx, string(stake), APIQueryParams)
	if err != nil {
		log.Printf("Could not get addresses for stake address: \nSTAKEADDR: %v \nERROR: %v", stake, err)
		return nil, err
	}

	if len(assets) == 100 {
		nextPageAssets, err := GetAllUserAddressesAssets(ctx, stake, page+1)
		if err != nil {
			return nil, err
		}
		assets = append(assets, nextPageAssets...)
	}

	return assets, nil
}

// Sum all the assets from the different wallets
func SumAllAssets(ctx context.Context, assets []bfg.AccountAssociatedAsset) (map[string]uint64) {
	allAssets := make(map[string]uint64) // Unit -> Quantity
	for _, asset := range assets {
		qty, err := strconv.Atoi(asset.Quantity)
		if err != nil {
			log.Printf("Could not convert asset quantity to int: \nASSET: %v \nERROR: %v", asset, err)
			continue
		}
		allAssets[asset.Unit] += uint64(qty)
	}

	return allAssets
}

// Create a map of all the user's assets for easy access
func GetAllUserAssets(ctx context.Context, wallets preeb.Wallets) (map[string]uint64, error) {
	var allAssets []bfg.AccountAssociatedAsset
	for stake, _ := range wallets {
		assets, err := GetAllUserAddressesAssets(ctx, stake, 1)
		if err != nil {
			log.Printf("Could not get addresses for stake address: \nSTAKEADDR: %v \nERROR: %v", stake, err)
			return nil, err
		}
		allAssets = append(allAssets, assets...)
	}

	summedAssets := SumAllAssets(ctx, allAssets)

	// slog.Info("Summed assets", "TOTAL", len(summedAssets), "ASSETS", summedAssets)
	return summedAssets, nil
}

func CountUserAssetsByPolicy(ctx context.Context, policyIDs preeb.PolicyIDs, allAssets map[string]uint64) map[string]int {
	var policyCounts = make(map[string]int)

	powInt := func (decimals int) float64 {
		return math.Pow(10, float64(decimals))
	}

	for unit, quantity := range allAssets {
		total := 0
		for policyID, policy := range policyIDs {
			if !HasAllGroupedPolicies(policy, allAssets) {
				continue // Skip this policy if grouping not satisfied
			}

			if strings.HasPrefix(unit, policyID) {
				logger.Record.Info("BLOCKFROST", "CALL", "GetAsset")
				asset, err := client.Asset(ctx, unit)
				if err != nil {
					log.Printf("Could not get asset details: \nUNIT: %v \nERROR: %v", unit, err)
					continue // Skip this UTXO if asset details cannot be retrieved
				}



				qty := int(quantity) // Default to quantity as is
				if asset.Metadata != nil && asset.Metadata.Decimals > 0 {
					qty = int(math.Floor(float64(quantity) / powInt(asset.Metadata.Decimals)))
				}

				// ✅ Trait matching logic for NFTs
				if len(policy.Traits) > 0 && asset.OnchainMetadata != nil {
					if !HasMatchingTrait(asset.OnchainMetadata, policy.Traits) {
						continue // Skip this UTXO if no trait matches
					}
				}

				total+= qty
				policyCounts[policyID] += total
			}
		}
	}

	slog.Info("Policy counts", "TOTAL", len(policyCounts), "COUNTS", policyCounts)
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
					slog.Info("Trait match found", "TRAIT", traitKey, "VALUE", valStr)
					return true
				}
			}
		}
	}

	return false
}



func HasAllGroupedPolicies(policy preeb.Policy, allAssets map[string]uint64) bool {
	if len(policy.GroupWith) == 0 {
		return true // No grouped policies required
	}

	// Track which required policies the user has
	held := make(map[string]bool)

	// For each address and its UTXOs
	for unit, _ := range allAssets {
		// Check if UTXO matches any policy in GroupWith
		for requiredPolicyID := range policy.GroupWith {
			if strings.HasPrefix(unit, requiredPolicyID) {
				held[requiredPolicyID] = true
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
		logger.Record.Info("BLOCKFROST", "CALL", "HandleAddress")
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
	logger.Record.Info("BLOCKFROST", "CALL", "AccountDelegationHistory")
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
	logger.Record.Info("BLOCKFROST", "CALL", "PoolHistory")
	APIQueryParams.Order = "desc"
	history, err := client.PoolHistory(ctx, poolID, APIQueryParams)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func PoolMeta(ctx context.Context, poolID string) (*bfg.PoolMetadata, error) {
	logger.Record.Info("BLOCKFROST", "CALL", "PoolMeta")
	info, err := client.PoolMetadata(ctx, poolID)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func PoolBlocks(ctx context.Context, poolID string) (bfg.PoolBlocks, error) {
	logger.Record.Info("BLOCKFROST", "CALL", "PoolBlocks")
	APIQueryParams.Order = "desc"
	blocks, err := client.PoolBlocks(ctx, poolID, APIQueryParams)
	if err != nil {
		return nil, err
	}

	return blocks, nil
}