/*
Package account implements a quota account, as part of the quota scheduler
algorithm.
*/
package account

// NumPriorities is the number of distinct priority buckets. For performance
// and code complexity reasons, this is a compile-time constant.
const NumPriorities = 3

// FreeBucket is the free priority bucket, where jobs may run even if they have
// no quota account or have an empty quota account.
const FreeBucket = NumPriorities

// PromoteThreshold is the account Vector at which the scheduler will consider
// promoting jobs.
const PromoteThreshold = 5

// DemoteThreshold is the account balance at which the scheduler will consider
// demoting jobs.
const DemoteThreshold = -5

// Vector is a used to describe per-prioritity-level quantities, such as
// account balances or charge rates.
type Vector [NumPriorities]float64

// IntVector is the integer equivalent of QuotaVector, to store things
// like per-bucket counts.
type IntVector [NumPriorities]int

// ID is an opaque globally unique identifier for a quota account.
type ID string

// Config represents per-quota-account configuration information, such
// as the recharge parameters. This does not represent anything about the
// current state of an account.
type Config struct {
	ChargeRate Vector // The rates (per second) at which per-priority accounts grow
	MaxBalance Vector // The maximum value that per-priority accounts grow to
	MaxFanout  int    // The maximum number of concurrent paid jobs that this account will pay for (0 = no limit)
}

// Less determines whether Vector a is less than b, based on
// priority ordered comparison
func (a Vector) Less(b Vector) bool {
	for i, valA := range a {
		valB := b[i]
		if valA < valB {
			return true
		}
		if valB < valA {
			return false
		}
	}
	return false
}

// Plus returns the sum of two vectors.
func (a Vector) Plus(b Vector) Vector {
	ans := Vector{}
	for i, aV := range a {
		ans[i] = aV + b[i]
	}
	return ans
}

// Minus returns the difference of two vectors.
func (a Vector) Minus(b Vector) Vector {
	ans := Vector{}
	for i, aV := range a {
		ans[i] = aV - b[i]
	}
	return ans
}

// BestPriorityFor determines the highest available priority for a quota
// account, given its balance. If the account is out of quota, this is
// the FreeBucket.
func BestPriorityFor(balance *Vector) int {
	for priority, value := range balance {
		if value > 0 {
			return priority
		}
	}
	return FreeBucket
}

// Update updates the state of a quota account by recharging the account for
// the elapsed time and draining the account for currently running jobs.
// TODO: Use time.Duration instead of float64 to represent elapsed time.
func (balance *Vector) Update(c *Config, elapsedSeconds float64, runningJobs IntVector) {
	for priority, value := range balance {
		value -= elapsedSeconds * float64(runningJobs[priority])
		if value < c.MaxBalance[priority] {
			value += elapsedSeconds * c.ChargeRate[priority]
			if value > c.MaxBalance[priority] {
				value = c.MaxBalance[priority]
			}
		}
		balance[priority] = value
	}
}
