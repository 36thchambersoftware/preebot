package blockfrost

import (
	"context"

	bfg "github.com/blockfrost/blockfrost-go"
)

func AssetInfo(ctx context.Context, asset string) (bfg.Asset, error) {
	info, err := client.Asset(ctx, asset)
	if err != nil {
		return bfg.Asset{}, err
	}

	return info, nil
}

func AssetsByPolicy(ctx context.Context, policyID string) ([]bfg.AssetByPolicy, error) {
	assets, err := client.AssetsByPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	return assets, nil
}