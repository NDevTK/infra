package account

// NumPriorities is the number of distinct priority buckets. For performance
// and code complexity reasons, this is a compile-time constant.
const NumPriorities = 3

// FreeBucket is the free priorty bucket, where jobs may run even if they have
// no quota account or have an empty quota account.
const FreeBucket = NumPriorities

// Vector is a NumPriorities-length float array, used to store things like
// account balances or charge rates.
type Vector [NumPriorities]float64

// IntVector is the integer equivalent of QuotaVector, to store things
// like per-bucket counts.
type IntVector [NumPriorities]int

// ID is an opaque globally unique identifier for a quota account.
type ID string

// Balance represents the amount of quota in various priority buckets
// that is currently available to a quota account.
type Balance Vector

// Config represents per-quota-account configuration information, such
// as the recharge parameters. This does not represent anything about the
// current state of an account.
type Config struct {
	ChargeRate Vector
	MaxBalance Vector
	MaxFanout  int
}

// BestPriorityFor determines the highest available priority for a quota
// account, given its balance. If the account is out of quota, this is
// the FreeBucket.
func BestPriorityFor(balance *Balance) int {
	for priority, value := range balance {
		if value > 0 {
			return priority
		}
	}
	return FreeBucket
}

// Advance update the state of a quota account, based on its recharge
// parameters, the number of elapsedSeconds since the last update, and
// the number of currently runningJobs for that account at the different
// priority levels.
func (balance *Balance) Advance(config *Config, elapsedSeconds float64, runningJobs IntVector) {
	for priority, value := range balance {
		value -= elapsedSeconds * float64(runningJobs[priority])
		if value < config.MaxBalance[priority] {
			value += elapsedSeconds * config.ChargeRate[priority]
			if value > config.MaxBalance[priority] {
				value = config.MaxBalance[priority]
			}
		}
		balance[priority] = value
	}
}
