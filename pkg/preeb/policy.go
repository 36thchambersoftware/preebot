package preeb

type (
	Policy struct {
		Roles PolicyRoles
		ChannelID string `bson:"channel_id,omitempty"`
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
)