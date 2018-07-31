package types

import "infra/qscheduler/qslib/types/account"

// Config represents configuration information about the behavior of accounts
// for this quota scheduler pool. It is expected to change infrequently,
// compared to State, and through different mechanisms (such as luci-config
// pushes) and thus is represented separately.
type Config struct {
	AccountConfigs map[account.ID]account.Config
}
