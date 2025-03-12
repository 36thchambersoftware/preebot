package preeb

type (
	Policy struct {
		Roles PolicyRoles
		DefaultChannelID string `bson:"default_channel_id,omitempty"`
		BuyNotifications []BuyNotification `bson:"buy_notifications,omitempty"`
		HexName string `bson:"hex_name,omitempty"`
		Notify bool `bson:"notify,omitempty"`
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
		Image string `bson:"image,omitempty"`
		Message string `bson:"message,omitempty"`
		ChannelID string `bson:"channel_id,omitempty"`
		Label string `bson:"label,omitempty"`
	}
)