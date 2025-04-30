package preeb

type (
	Policy struct {
		Roles PolicyRoles
		DefaultChannelID string `bson:"default_channel_id,omitempty"`
		BuyNotifications []BuyNotification `bson:"buy_notifications,omitempty"`
		HexName string `bson:"hex_name,omitempty"` // Tokens only
		Notify bool `bson:"notify,omitempty"`
		NFT bool `bson:"nft,omitempty"`
		Mint bool `bson:"mint,omitempty"`
		PriceUpdate bool `bson:"price_update,omitempty"`
		Price float64 `bson:"price,omitempty"`
		PriceChannel string `bson:"price_channel,omitempty"`
		CompareADA bool `bson:"in_ada,omitempty"`
		MetadataKeys []string `bson:"metadata_keys,omitempty"`
		Message string `bson:"message,omitempty"`
	}

	// String here being the policy id
	PolicyIDs map[string]Policy

	// String here being the role id
	PolicyRoles map[string]RoleBounds

	RoleBounds struct {
		Min Bound `bson:"min,omitempty"`
		Max Bound `bson:"max,omitempty"`
		Order int64 `bson:"order,omitempty"`
	}

	BuyNotification struct {
		Min Bound `bson:"min,omitempty"`
		Max Bound `bson:"max,omitempty"`
		Image string `bson:"image,omitempty"` // Tokens only
		Message string `bson:"message,omitempty"`
		ChannelID string `bson:"channel_id,omitempty"`
		Label string `bson:"label,omitempty"`
	}
)