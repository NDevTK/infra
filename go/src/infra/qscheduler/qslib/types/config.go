package types

import "infra/qscheduler/qslib/types/account"

// Config represents configuration information about the behavior of accounts
// for this quota scheduler pool.
type Config struct {
	AccountConfigs map[account.ID]account.Config
}
