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
