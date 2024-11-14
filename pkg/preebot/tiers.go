package preebot

func GetDelegatorRolesByStake(controlledAmount int, roles DelegatorRoles) []string {
	var rolesToAssign []string
	for role, bounds := range roles {
		if controlledAmount > int(bounds.Min) && controlledAmount < int(bounds.Max) {
			rolesToAssign = append(rolesToAssign, role)
		}
	}

	return rolesToAssign
}

func GetPolicyRoles(assetCount int, roles PolicyRoles) []string {
	var rolesToAssign []string
	for role, bounds := range roles {
		if assetCount > int(bounds.Min) && assetCount < int(bounds.Max) {
			rolesToAssign = append(rolesToAssign, role)
		}
	}

	return rolesToAssign
}