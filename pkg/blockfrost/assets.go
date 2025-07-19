package blockfrost

import (
	"context"
	"log/slog"

	bfg "github.com/blockfrost/blockfrost-go"
)

func AssetInfo(ctx context.Context, asset string) (bfg.Asset, error) {
	info, err := client.Asset(ctx, asset)
	if err != nil {
		return bfg.Asset{}, err
	}

	if info.OnchainMetadata != nil {
		// Dereference and attempt type assertion
		raw := *info.OnchainMetadata
		if parsed, ok := raw.(map[string]interface{}); ok {
			// Re-box and reassign back to the pointer
			var boxed interface{} = parsed
			info.OnchainMetadata = &boxed
		} else {
			// If it's not the expected format, you could log or nil it out
			slog.Warn("OnchainMetadata is not a map[string]interface{}", "asset", asset)
			info.OnchainMetadata = nil
		}
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