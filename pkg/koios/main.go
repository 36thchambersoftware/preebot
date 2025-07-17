package koios

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"preebot/pkg/preeb"
	"sort"

	"github.com/cardano-community/koios-go-client/v4"
)

var client *koios.Client
type EpochNo koios.EpochNo

func init() {
	var err error
	client, err = koios.New()
	if err != nil {
		slog.Error("could not connect koios api", "ERROR", err)
	}
}

func AddressInformation(ctx context.Context, addresses []preeb.Address) ([]koios.AddressInfo, error) {
	var options *koios.RequestOptions
	var addrs []koios.Address

	for _, address := range addresses {
		addrs = append(addrs, koios.Address(address))
	}

	result, err := client.GetAddressesInfo(ctx, addrs, options)
	if err != nil {
		return nil, err
	}

	if result.StatusCode != 200 {
		return nil, errors.New(result.Response.Error.Message)
	}

	return result.Data, nil
}

func AccountHistory(ctx context.Context, stake_addresses []preeb.StakeAddress) ([]koios.AccountHistory, error) {
	var addrs []koios.Address
	var epoch *koios.EpochNo
	var options *koios.RequestOptions
	for _, address := range stake_addresses {
		addrs = append(addrs, koios.Address(address))
	}
	result, err := client.GetAccountHistory(ctx, addrs, epoch, options)
	if err != nil {
		return nil, err
	}

	if result.StatusCode != 200 {
		return nil, errors.New(result.Response.Error.Message)
	}

	return result.Data, nil
}

func Tip(ctx context.Context) (res *koios.Tip, err error) {
	var options *koios.RequestOptions
	result, err := client.GetTip(ctx, options)
	if err != nil {
		return nil, err
	}

	if result.StatusCode != 200 {
		return nil, errors.New(result.Response.Error.Message)
	}

	return &result.Data, nil
}

func GetPolicyTxs(ctx context.Context, policyID, assetName string) ([]koios.AddressTx, error) {
	var options *koios.RequestOptions
	txs, err := client.GetAssetTxs(ctx, koios.PolicyID(policyID), koios.AssetName(assetName), 0, false, options)
	if err != nil {
		return nil, err
	}

	return txs.Data, nil
}

func GetPolicyAssetMints(ctx context.Context, policyID string) ([]koios.PolicyAssetMint, error) {
	var options *koios.RequestOptions
	mints, err := client.GetPolicyAssetMints(ctx, koios.PolicyID(policyID), options)
	if err != nil {
		return nil, err
	}

	if mints.StatusCode != 200 {
		return nil, errors.New(mints.Response.Error.Message)
	}

	sort.Slice(mints.Data, func(i, j int) bool {
		return mints.Data[i].CreationTime.After(mints.Data[j].CreationTime.Time)
	})

	return mints.Data, nil
}

func GetAssetUTXOs(ctx context.Context, policyID, assetName string) (koios.UTxO, error) {
	var options *koios.RequestOptions

	asset := []koios.Asset{{
		PolicyID: koios.PolicyID(policyID),
		AssetName: koios.AssetName(assetName),
	}}
	utxos, err := client.GetAssetUTxOs(ctx, asset, options)
	if err != nil {
		return koios.UTxO{}, err
	}

	if utxos.StatusCode != 200 {
		return koios.UTxO{}, errors.New(utxos.Response.Error.Message)
	}

	if len(utxos.Data) == 0 {
		return koios.UTxO{}, errors.New(fmt.Sprintf("no UTxOs found: %s %v", policyID, utxos))
	}

	return utxos.Data[0], nil
}

func GetDatum(ctx context.Context, datumHash koios.DatumHash) (*koios.DatumInfo, error) {
	var options *koios.RequestOptions
	datum, err := client.GetDatumInfo(ctx, datumHash, options)
	if err != nil {
		return nil, err
	}

	if datum.StatusCode != 200 {
		return nil, errors.New(datum.Response.Error.Message)
	}

	return datum.Data, nil
}

// func GetTransactionDetails(ctx context.Context, txHash string) {
// 	var options *koios.RequestOptions
// 	client.GetTxInfo(ctx, txs []koios.TxHash, opts *koios.RequestOptions)
// }