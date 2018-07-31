/*
Package account implements a quota account, as part of the quota scheduler
algorithm.
*/
package account

import "infra/qscheduler/qslib/types/vector"

// FreeBucket is the free priority bucket, where jobs may run even if they have
// no quota account or have an empty quota account.
const FreeBucket = int32(vector.NumPriorities)

// PromoteThreshold is the account Vector at which the scheduler will consider
// promoting jobs.
const PromoteThreshold = 5.0

// DemoteThreshold is the account balance at which the scheduler will consider
// demoting jobs.
const DemoteThreshold = -5.0

// BestPriorityFor determines the highest available priority for a quota
// account, given its balance. If the account is out of quota, this is
// the FreeBucket.
func BestPriorityFor(balance vector.Vector) int32 {
	for priority, value := range balance.Values {
		if value > 0 {
			return int32(priority)
		}
	}
	return FreeBucket
}

// UpdateBalance updates the state of a quota account by recharging the account
// for the elapsed time and draining the account for currently running jobs.
// TODO: Use time.Duration instead of float64 to represent elapsed time.
func UpdateBalance(balance *vector.Vector, c Config, elapsedSeconds float64, runningJobs *vector.IntVector) {
	// TODO: rewrite me using vector math primitives. This also will allow
	// the vector package to internally call fix() on anything necessary.
	for priority, value := range balance.Values {
		value -= elapsedSeconds * float64(runningJobs[priority])
		if value < c.MaxBalance.Values[priority] {
			value += elapsedSeconds * c.ChargeRate.Values[priority]
			if value > c.MaxBalance.Values[priority] {
				value = c.MaxBalance.Values[priority]
			}
		}
		balance.Values[priority] = value
	}
}
