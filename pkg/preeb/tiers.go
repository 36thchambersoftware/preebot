package preeb

func GetDelegatorRolesByStake(controlledAmount int, roles DelegatorRoles) []string {
	var rolesToAssign []string
	for role, bounds := range roles {
		if controlledAmount > int(bounds.Min) && controlledAmount < int(bounds.Max) {
			rolesToAssign = append(rolesToAssign, role)
		}
	}

	return rolesToAssign
}

func GetPolicyRoles(assetCount map[string]int, policyIDs PolicyIDs) []string {
	var rolesToAssign []string
	for policyID, policy := range policyIDs {
		for role, bounds := range policy.Roles {
			if assetCount[policyID] >= int(bounds.Min) && assetCount[policyID] < int(bounds.Max) {
				rolesToAssign = append(rolesToAssign, role)
			}
		}
	}

	return rolesToAssign
}