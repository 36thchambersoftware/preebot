package preebot

import "strconv"

// Sequential epochs delegated to pool
const (
	EpochTier1 = 10
	EpochTier2 = 20
	EpochTier3 = 50
	EpochTier4 = 100
)

type EpochTier int

// Amount of ADA delegated to pool
const (
	ADA_TIER_0   = 1_000_000
	ADA_TIER_1   = 500_000_000
	ADA_TIER_2   = 5_000_000_000
	ADA_TIER_3   = 50_000_000_000
	ADA_TIER_4   = 500_000_000_000
	ADA_TIER_5   = 1_000_000_000_000
	DELEGATOR    = "Delegator"
	PANDA        = "PANDA"
	BLACK_BEAR   = "BLACK BEAR"
	GRIZZLY_BEAR = "GRIZZLY BEAR"
	POLAR_BEAR   = "POLAR BEAR"
	CARE_BEAR    = "CARE BEAR"
)

var Tiers = map[int]string{
	ADA_TIER_0: DELEGATOR,
	ADA_TIER_1: PANDA,
	ADA_TIER_2: BLACK_BEAR,
	ADA_TIER_3: GRIZZLY_BEAR,
	ADA_TIER_4: POLAR_BEAR,
	ADA_TIER_5: CARE_BEAR,
}

func GetTier(activeEpoch *int64, controlledAmount string) string {
	balance, _ := strconv.Atoi(controlledAmount)

	if balance >= ADA_TIER_5 {
		return Tiers[ADA_TIER_5]
	}

	if balance >= ADA_TIER_4 {
		return Tiers[ADA_TIER_4]
	}

	if balance >= ADA_TIER_3 {
		return Tiers[ADA_TIER_3]
	}

	if balance >= ADA_TIER_2 {
		return Tiers[ADA_TIER_2]
	}

	if balance >= ADA_TIER_1 {
		return Tiers[ADA_TIER_1]
	}

	if balance < ADA_TIER_1 {
		return Tiers[ADA_TIER_0]
	}

	return ""
}
